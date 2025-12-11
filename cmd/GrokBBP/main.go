package main

import (
	"fmt"

	"GrokBBP/internal/pi"
)

const digits = 50 // change whenever you want

func main() {
	hex := pi.ComputeHexDigits(digits)

	fmt.Printf("π — first %d hexadecimal digits after the point:\n3.", digits)
	for i, d := range hex {
		fmt.Printf("%x", d)
		if (i+1)%80 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}
