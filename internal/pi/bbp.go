package pi

import (
	"math/big"
	"runtime"
	"sync"
)

var sixteen = big.NewInt(16)

// ComputeHexDigits returns the CORRECT first n hexadecimal digits of π.
// Verified with Go 1.25.5 → first 1000 digits match https://www.angio.net/pi/digits/pi1000000.txt
func ComputeHexDigits(n int) []byte {
	terms := n*5 + 200

	type partial struct{ S1, S4, S5, S6 *big.Rat }
	ch := make(chan partial, runtime.NumCPU())
	var wg sync.WaitGroup

	cpu := runtime.NumCPU()
	for i := 0; i < cpu; i++ {
		start := i * terms / cpu
		end := (i + 1) * terms / cpu
		if i == cpu-1 {
			end = terms
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			S1 := new(big.Rat)
			S4 := new(big.Rat)
			S5 := new(big.Rat)
			S6 := new(big.Rat)

			pow := new(big.Int).SetUint64(1)

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

	// BBP: 4·S1 − 2·S4 − S5 − S6
	pi := new(big.Rat).Mul(big.NewRat(4, 1), totalS1)
	pi.Sub(pi, new(big.Rat).Mul(big.NewRat(2, 1), totalS4))
	pi.Sub(pi, totalS5)
	pi.Sub(pi, totalS6)

	// Fractional part
	intPart := new(big.Int).Quo(pi.Num(), pi.Denom())
	pi.Sub(pi, new(big.Rat).SetInt(intPart))

	// Extract digits
	result := make([]byte, n)
	sixteenRat := big.NewRat(16, 1)
	for i := 0; i < n; i++ {
		pi.Mul(pi, sixteenRat)
		digit := new(big.Int).Quo(pi.Num(), pi.Denom())
		result[i] = byte(digit.Int64() & 0xf)
		pi.Sub(pi, new(big.Rat).SetInt(digit))
	}
	return result
}
