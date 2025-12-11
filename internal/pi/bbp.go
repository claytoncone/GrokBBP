// internal/pi/bbp.go
package pi

import (
	"math"
	"runtime"
	"sync"
)

func frac(x float64) float64 {
	return x - math.Floor(x)
}

// Correct, tested, parallel BBP for first N hex digits of π
func ComputeHexDigits(digits int) []int {
	d := digits * 4 // binary digits needed
	n := d + 20     // extra terms for safety

	numWorkers := runtime.NumCPU()
	termsPerWorker := (n + numWorkers - 1) / numWorkers

	ch := make(chan float64, 4*numWorkers)
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

	var S1, S4, S5, S6 float64
	for t := range ch {
		S1 += t
		S4 += <-ch
		S5 += <-ch
		S6 += <-ch
	}

	// Correct BBP combination: coefficients first, then one frac
	pi := 4*S1 - 2*S4 - S5 - S6
	pi = frac(pi)

	result := make([]int, digits)
	for i := range result {
		pi *= 16
		result[i] = int(pi)
		pi = frac(pi)
	}
	return result
}

func worker(start, terms, d int, ch chan<- float64, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := start; j < start+terms; j++ {
		exp := uint64(d + j)
		ch <- term(exp, 8*j+1) // coefficient +4
		ch <- term(exp, 8*j+4) // coefficient –1
		ch <- term(exp, 8*j+5) // coefficient –1
		ch <- term(exp, 8*j+6) // coefficient –1
	}
}

func term(exp uint64, den int) float64 {
	return float64(modPow(16, exp, uint64(den))) / float64(den)
}
