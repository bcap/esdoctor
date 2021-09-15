package math

func Pct(a int, b int) float64 {
	return float64(a) / float64(b) * 100.0
}

func Pct64(a int64, b int64) float64 {
	return float64(a) / float64(b) * 100.0
}

func Fraction(a int64, b int64) (int64, int64, float64) {
	gcd := GreatestCommonDivisor(a, b)
	return a / gcd, b / gcd, Pct64(a, b)
}
