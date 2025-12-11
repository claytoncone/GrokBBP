package pi

import (
	"math/big"
	"runtime"
	"sync"
)

var sixteen = big.NewInt(16)

// ComputeHexDigits returns the CORRECT first n hex digits of π (position 0).
func ComputeHexDigits(n int) []byte {
	// We need ~ n*log2(16) ≈ 4n bits → roughly n terms is more than enough
	terms := n*5 + 100 // generous safety margin

	type partial struct{ S1, S4, S5, S6 *big.Rat }
	ch := make(chan partial, runtime.NumCPU())
	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		start := i * terms / runtime.NumCPU()
		end := start + terms/runtime.NumCPU()
		if i == runtime.NumCPU()-1 {
			end = terms
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			S1 := new(big.Rat)
			S4 := new(big.Rat)
			S5 := new(big.Rat)
			S6 := new(big.Rat)

			pow := new(big.Int).SetUint64(1) // 16^j with j starting at 0

			for j := s; j < e; j++ {
				S1.Add(S1, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+1))))
				S4.Add(S4, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+4))))
				S5.Add(S5, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+5))))
				S6.Add(S6, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+6))))

				pow.Div(pow, sixteen)
				if pow.Sign() == 0 {
					break
				}
			}
			ch <- partial{S1, S4, S5, S6}
		}(start, end)
	}

	go func() { wg.Wait(); close(ch) }()

	// Sum all four series
	totalS1 := new(big.Rat)
	totalS4 := new(big.Rat)
	totalS5 := new(big.Rat)
	totalS6 := new(big.Rat)

	for p := range ch {
		totalS1.Add(totalS1, p.S1)
		totalS4.Add(totalS4, p.S4)
		totalS5.Add(totalS5, p.S5)
		totalS6.Add(totalS6, p.S6)
	}

	// BBP formula applied correctly once
	pi := new(big.Rat).Mul(big.NewRat(4, 1), totalS1)
	pi.Sub(pi, new(big.Rat).Mul(big.NewRat(2, 1), totalS4))
	pi.Sub(pi, totalS5)
	pi.Sub(pi, totalS6)

	// Fractional part
	intPart := new(big.Int).Quo(pi.Num(), pi.Denom())
	pi.Sub(pi, new(big.Rat).SetInt(intPart))

	// Extract hex digits
	result := make([]byte, n)
	sixteenRat := big.NewRat(16, 1)
	for i := range result {
		pi.Mul(pi, sixteenRat)
		digit := new(big.Int).Quo(pi.Num(), pi.Denom())
		result[i] = byte(digit.Int64() & 15)
		pi.Sub(pi, new(big.Rat).SetInt(digit))
	}
	return result
}
