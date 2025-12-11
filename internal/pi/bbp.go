// internal/pi/bbp.go
package pi

import (
	"math"
	"sync"
)

func frac(x float64) float64 {
	return x - math.Floor(x)
}

// THIS IS THE CORRECT worker — do NOT reset S1/S4/S5/S6 inside the loop!
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

		// accumulate — never reset here!
		S1 = (S1 + modPow(16, p, d1)<<scale) % d1
		S4 = (S4 + modPow(16, p, d4)<<scale) % d4
		S5 = (S5 + modPow(16, p, d5)<<scale) % d5
		S6 = (S6 + modPow(16, p, d6)<<scale) % d6
	}

	// send the FOUR final fractional parts exactly once per worker
	j := start + terms - 1 // use the last j of this block (mathematically valid)
	ch <- float64(S1) / float64(uint64(8*j+1)<<scale)
	ch <- float64(S4) / float64(uint64(8*j+4)<<scale)
	ch <- float64(S5) / float64(uint64(8*j+5)<<scale)
	ch <- float64(S6) / float64(uint64(8*j+6)<<scale)
}

func ComputeHexDigits(digits int) []int {
	d := digits*4 + 10
	n := d + 20

	const workers = 12
	termsPerWorker := (n + workers - 1) / workers

	ch := make(chan float64, workers*4)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
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

	// simple sum is actually enough with this fixed-point trick
	var sum float64
	for v := range ch {
		sum += v
	}

	// BBP combination
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
