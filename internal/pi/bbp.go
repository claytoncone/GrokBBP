package pi

import (
	"math"
	"sync"
)

// ComputeHexDigits returns the first `digits` hexadecimal digits of π (after 3.)
func ComputeHexDigits(digits, numWorkers, termsPerWorker int) []int {
	d := digits * 4 // safety margin
	n := d + 10

	ch := make(chan float64, numWorkers*4)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * termsPerWorker
		terms := termsPerWorker
		if start+terms > n {
			terms = n - start
		}
		if terms <= 0 {
			break
		}
		wg.Add(1)
		go worker(start, terms, d, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// Kahan summation for high precision
	sum := kahanSum(ch)

	// Apply BBP formula coefficients
	sum = 4*frac(sum) - 2*frac(2*sum) - frac(3*sum) - frac(4*sum)
	sum = frac(sum)

	// Extract hex digits
	hex := make([]int, digits)
	for i := 0; i < digits; i++ {
		sum *= 16
		digit := int(sum)
		hex[i] = digit
		sum = frac(sum)
	}
	return hex
}

func worker(start, terms, d int, ch chan<- float64, wg *sync.WaitGroup) {
	defer wg.Done()

	var S1, S4, S5, S6 uint64 = 0, 0, 0, 0

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

		// Fixed-point scaling with 2^24 (empirically excellent)
		const scale = 24
		S1 = (S1 + pow1<<scale) % den1
		S4 = (S4 + pow4<<scale) % den4
		S5 = (S5 + pow5<<scale) % den5
		S6 = (S6 + pow6<<scale) % den6
	}

	const shift = 24
	ch <- float64(S1) / float64(den1<<shift)
	ch <- float64(S4) / float64(den4<<shift)
	ch <- float64(S5) / float64(den5<<shift)
	ch <- float64(S6) / float64(den6<<shift)
}

func frac(x float64) float64 {
	return x - math.Floor(x)
}
