package pi

import (
	"math"
	"sync"
)

func frac(x float64) float64 {
	return x - math.Floor(x)
}

func ComputeHexDigits(digits int) []int {
	d := digits * 4 // bits per hex digit margin
	n := d + 20     // extra terms for rounding safety

	numWorkers := 8 // or runtime.NumCPU()
	termsPerWorker := (n + numWorkers - 1) / numWorkers

	ch := make(chan float64, 4*n) // buffer all terms
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * termsPerWorker
		terms := termsPerWorker
		if start+terms > n {
			terms = n - start
		}
		if terms <= 0 {
			continue
		}
		wg.Add(1)
		go worker(start, terms, d, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var sum float64
	for term := range ch {
		sum += term
	}

	// BBP coefficients (mod 1)
	sum = 4*frac(sum) - 2*frac(2*sum) - frac(3*sum) - frac(4*sum)
	sum = frac(sum)

	hex := make([]int, digits)
	for i := range hex {
		sum *= 16
		hex[i] = int(sum)
		sum = frac(sum)
	}
	return hex
}

func worker(start, terms, d int, ch chan<- float64, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := start; j < start+terms; j++ {
		p := uint64(d + j)
		den1 := uint64(8*j + 1)
		den4 := uint64(8*j + 4)
		den5 := uint64(8*j + 5)
		den6 := uint64(8*j + 6)

		pow1 := modPow(16, p, den1)
		pow4 := modPow(16, p, den4)
		pow5 := modPow(16, p, den5)
		pow6 := modPow(16, p, den6)

		// Exact fractional terms: {16^{d+j} / den}
		ch <- frac(float64(pow1) / float64(den1))
		ch <- frac(float64(pow4) / float64(den4))
		ch <- frac(float64(pow5) / float64(den5))
		ch <- frac(float64(pow6) / float64(den6))
	}
}
