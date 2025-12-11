package pi

import (
	"math/big"
	"runtime"
	"sync"
)

var sixteen = big.NewInt(16)

// ComputeHexDigits returns the first n hexadecimal digits of π after 3.
func ComputeHexDigits(n int) []byte {
	terms := n + 20 // more than enough

	type partial struct{ s *big.Rat }
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
			r := new(big.Rat)
			pow := new(big.Int).SetUint64(1)
			for j := s; j < e; j++ {
				// 4/(8j+1)
				r.Add(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+1))))
				r.Add(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+1))))
				r.Add(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+1))))
				r.Add(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+1))))

				// –2/(8j+4)
				t := new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+4)))
				r.Sub(r, t)
				r.Sub(r, t)

				// –1/(8j+5)
				r.Sub(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+5))))

				// –1/(8j+6)
				r.Sub(r, new(big.Rat).SetFrac(pow, big.NewInt(int64(8*j+6))))

				pow.Div(pow, sixteen)
				if pow.Sign() == 0 {
					break
				}
			}
			ch <- partial{r}
		}(start, end)
	}

	go func() { wg.Wait(); close(ch) }()

	total := new(big.Rat)
	for p := range ch {
		total.Add(total, p.s)
	}

	// fractional part only
	total.Sub(total, new(big.Rat).SetInt(new(big.Int).Quo(total.Num(), total.Denom())))

	digits := make([]byte, n)
	for i := range digits {
		total.Mul(total, big.NewRat(16, 1))
		d := new(big.Int).Quo(total.Num(), total.Denom())
		digits[i] = byte(d.Int64() & 15)
		total.Sub(total, new(big.Rat).SetInt(d))
	}
	return digits
}
