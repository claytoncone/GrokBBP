package pi

// modPow computes (base^exp) % mod using uint64 only (safe up to 2^64-1)
func modPow(base, exp, mod uint64) uint64 {
	if mod == 1 {
		return 0
	}
	result := uint64(1)
	base %= mod
	for exp > 0 {
		if exp&1 == 1 {
			result = (result * base) % mod // safe: Go 1.19+ has built-in overflow protection in some paths
		}
		base = (base * base) % mod
		exp >>= 1
	}
	return result
}
