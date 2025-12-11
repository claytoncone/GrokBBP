package pi

import (
	"math/big"
	"runtime"
	"sync"
)

func ComputeHexDigits(digits int) []int {
	terms := digits*5 + 100 // safe number of terms

	numWorkers := runtime.NumCPU()
	ch := make(chan *big.Rat, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * (terms / numWorkers)
		end := start + (terms / numWorkers)
		if i == numWorkers-1 {
			end = terms
		}
		wg.Add(1)
		go worker(start, end, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	S := new(big.Rat)
	for part := range ch {
		S.Add(S, part)
	}

	// π ≈ 4*S1 - 2*S4 - S5 - S6
	pi := new(big.Rat).Mul(big.NewRat(4, 1), S)
	pi.Sub(pi, new(big.Rat).Mul(big.NewRat(2, 1), S))
	pi.Sub(pi, S)
	pi.Sub(pi, S)

	// {pi}
	intPart := new(big.Int).Quo(pi.Num(), pi.Denom())
	pi.Sub(pi, new(big.Rat).SetInt(intPart))

	result := make([]int, digits)
	sixteen := big.NewRat(16, 1)
	for i := 0; i < digits; i++ {
		pi.Mul(pi, sixteen)
		digit := new(big.Int).Quo(pi.Num(), pi.Denom())
		result[i] = int(digit.Int64() & 15)
		pi.Sub(pi, new(big.Rat).SetInt(digit))
	}
	return result
}

func worker(start, end int, ch chan<- *big.Rat, wg *sync.WaitGroup) {
	defer wg.Done()
	sum := new(big.Rat)
	pow := new(big.Int).SetUint64(1)

	for j := start; j < end; j++ {
		d1 := int64(8*j + 1)
		d4 := int64(8*j + 4)
		d5 := int64(8*j + 5)
		d6 := int64(8*j + 6)

		// 4/(8j+1)
		t := new(big.Rat).SetFrac(pow, big.NewInt(d1))
		sum.Add(sum, t)
		sum.Add(sum, t)
		sum.Add(sum, t)
		sum.Add(sum, t) // 4×

		// -2/(8j+4)
		t.SetFrac(pow, big.NewInt(d4))
		sum.Sub(sum, t)
		sum.Sub(sum, t) // 2×

		// -1/(8j+5)
		t.SetFrac(pow, big.NewInt(d5))
		sum.Sub(sum, t)

		// -1/(8j+6)
		t.SetFrac(pow, big.NewInt(d6))
		sum.Sub(sum, t)

		pow.Div(pow, big.NewInt(16))
		if pow.Sign() == 0 {
			break
		}
	}
	ch <- sum
}
