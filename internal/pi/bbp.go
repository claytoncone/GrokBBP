// internal/pi/bbp.go
package pi

import (
	"math"
	"sync"
)

// frac returns the fractional part of x
func frac(x float64) float64 {
	return x - math.Floor(x)
}

// ComputeHexDigits – the only public function
func ComputeHexDigits(digits int) []int {
	d := digits*4 + 10 // safety margin
	n := d + 20        // a few extra terms never hurt
	numWorkers := 12   // 8–16 is sweet spot for BBP in 2025
	termsPerWorker := (n + numWorkers - 1) / numWorkers

	ch := make(chan float64, numWorkers*128)
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

	sum := kahanSum(ch)

	// BBP final combination
	sum = 4*frac(sum) - 2*frac(2*sum) - frac(3*sum) - frac(4*sum)
	sum = frac(sum)

	hex := make([]int, digits)
	for i := range hex {
		sum *= 16
		digit := int(sum)
		hex[i] = digit
		sum = frac(sum)
	}
	return hex
}

func worker(start, terms, d int, ch chan<- float64, wg *sync.WaitGroup) {
	defer wg.Done()
	const scale = 24

	var S1, S4, S5, S6 uint64

	for j := start; j < start+terms; j++ {
		p := uint64(d + j)
		d1 := uint64(8*j + 1)
		d4 := uint64(8*j + 4)
		d5 := uint64(8*j + 5)
		d6 := uint64(8*j + 6)

		S1 = (S1 + modPow(16, p, d1)<<scale) % d1
		S4 = (S4 + modPow(16, p, d4)<<scale) % d4
		S5 = (S5 + modPow(16, p, d5)<<scale) % d5
		S6 = (S6 + modPow(16, p, d6)<<scale) % d6

		// send each term with its own denominator
		ch <- float64(S1) / float64(d1<<scale)
		ch <- float64(S4) / float64(d4<<scale)
		ch <- float64(S5) / float64(d5<<scale)
		ch <- float64(S6) / float64(d6<<scale)

		S1, S4, S5, S6 = 0, 0, 0, 0 // reset for next term
	}
}
