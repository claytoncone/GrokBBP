package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/claytoncone/GrokBBP/internal/pi"
)

const (
	// Change this to compute more/fewer hex digits
	Digits = 1000
)

func main() {
	// Position: we compute digits starting after the initial "3."
	d := Digits * 4 // 4x safety margin in binary digits
	n := d + 10     // terms needed

	numWorkers := runtime.NumCPU()
	termsPerWorker := (n + numWorkers - 1) / numWorkers

	hexDigits := pi.ComputeHexDigits(Digits, numWorkers, termsPerWorker)

	fmt.Printf("π (first %d hex digits after 3.):\n3.", Digits)
	for i, d := range hexDigits {
		fmt.Printf("%x", d)
		if (i+1)%80 == 0 {
			fmt.Println()
		}
	}
	fmt.Println("\nDone!")
}
