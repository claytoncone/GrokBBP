package pi

import (
	"math/big"
	"runtime"
	"sync"
)

type partial struct {
	S1, S4, S5, S6 *big.Rat
}

// frac for big.Rat
func frac(r *big.Rat) *big.Rat {
	f := new(big.Rat).Set(r)
	i := new(big.Rat).SetFrac(f.Num().Div(f.Num(), f.Denom()), big.NewRat(1, 1))
	f.Sub(r, i)
	if f.Sign() < 0 {
		f.Add(f, big.NewRat(1, 1))
	}
	return f
}

// ComputeHexDigits uses big.Rat for exact first N hex digits
func ComputeHexDigits(digits int) []int {
	n := digits*4/3 + 100 // conservative number of terms for convergence

	numWorkers := runtime.NumCPU()
	ch := make(chan partial, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		start := i * (n / numWorkers)
		end := start + (n / numWorkers)
		if i == numWorkers-1 {
			end = n
		}
		wg.Add(1)
		go worker(start, end, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var totalS1, totalS4, totalS5, totalS6 big.Rat
	for p := range ch {
		totalS1.Add(&totalS1, p.S1)
		totalS4.Add(&totalS4, p.S4)
		totalS5.Add(&totalS5, p.S5)
		totalS6.Add(&totalS6, p.S6)
	}

	// BBP combination
	S1 := frac(&totalS1)
	S4 := frac(&totalS4)
	S5 := frac(&totalS5)
	S6 := frac(&totalS6)
	pi := new(big.Rat).Mul(big.NewRat(4, 1), S1)
	S2 := new(big.Rat).Mul(big.NewRat(2, 1), S4)
	pi.Sub(pi, S2)
	pi.Sub(pi, S5)
	pi.Sub(pi, S6)

	// Extract hex digits
	hex := make([]int, digits)
	one := big.NewRat(1, 1)
	sixteen := big.NewRat(16, 1)
	for i := range hex {
		pi.Mul(pi, sixteen)
		digit := new(big.Int)
		pi.Num().Div(pi.Num(), pi.Denom()).Scan(digit)
		hex[i] = int(digit.Int64())
		pi = frac(pi)
	}
	return hex
}

func worker(start, end int, ch chan<- partial, wg *sync.WaitGroup) {
	defer wg.Done()

	var S1, S4, S5, S6 big.Rat

	pow16 := new(big.Int).SetUint64(1)
	for j := start; j < end; j++ {
		den1 := big.NewInt(int64(8*j + 1))
		den4 := big.NewInt(int64(8*j + 4))
		den5 := big.NewInt(int64(8*j + 5))
		den6 := big.NewInt(int64(8*j + 6))

		term1 := new(big.Rat).SetFrac(pow16, den1)
		S1.Add(&S1, term1)

		term4 := new(big.Rat).SetFrac(pow16, den4)
		S4.Add(&S4, term4)

		term5 := new(big.Rat).SetFrac(pow16, den5)
		S5.Add(&S5, term5)

		term6 := new(big.Rat).SetFrac(pow16, den6)
		S6.Add(&S6, term6)

		// Update pow16 = pow16 / 16 for next j
		pow16.Rsh(pow16, 4) // 16 = 2^4
		if pow16.Sign() == 0 {
			break // no more contribution
		}
	}

	ch <- partial{S1: &S1, S4: &S4, S5: &S5, S6: &S6}
}
