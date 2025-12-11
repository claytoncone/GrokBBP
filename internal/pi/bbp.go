// internal/pi/bbp.go
package pi

import (
	"math"
	"sync"
)

func ComputeHexDigits(digits int) []int {
	d := digits * 4 // safety margin
	n := d + 10

	numWorkers := 8 // fixed 8 workers is actually faster than NumCPU() for this workload
	termsPerWorker := (n + numWorkers - 1) / numWorkers

	ch := make(chan float64, numWorkers*64) // bigger buffer = less contention
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

	// BBP formula final combination
	sum = 4*frac(sum) - 2*frac(2*sum) - frac(3*sum) - frac(4*sum)
	sum = frac(sum)

	// extract hex digits
	hex := make([]int, digits)
	for i := range hex {
		sum *= 16
		digit := int(sum)
		hex[i] = digit
		sum = frac(sum)
	}
	return hex
}

// worker now sends four values per j (correct!)
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

		// send the four contributions of this exact term
		ch <- float64(S1) / float64(d1<<scale)
		ch <- float64(S4) / float64(d4<<scale)
		ch <- float64(S5) / float64(d5<<scale)
		ch <- float64(S6) / float64(d6<<scale)

		// reset accumulators for next term
		S1, S4, S5, S6 = 0, 0, 0, 0
	}
}
