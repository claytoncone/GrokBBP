package pi

// kahanSum performs compensated summation to minimize floating-point error
func kahanSum(ch <-chan float64) float64 {
	var sum, c float64 = 0.0, 0.0
	for frac := range ch {
		y := frac - c
		t := sum + y
		c = (t - sum) - y
		sum = t
	}
	return sum
}
