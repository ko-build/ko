package fastlog

import "math"

// The following two functions were copied from fastapprox (BSD license).
// They do not calculate the precise value - and we do not need it.

// Log2 calculates the approximate base-2 logarithm.
func Log2(x float32) float32 {
	vx := math.Float32bits(x)
	mx := math.Float32frombits((vx & 0x007FFFFF) | 0x3f000000)
	y := float32(vx) * 1.1920928955078125e-7
	return y - 124.22551499 - 1.498030302*mx - 1.72587999/(0.3520887068+mx)
}

// Log calculates the approximate natural logarithm.
func Log(x float32) float32 {
	return 0.69314718 * Log2(x)
}
