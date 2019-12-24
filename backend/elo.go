package main

import "math"

/// Calculate win probability using a logistic curve.
func EloPredict(a, b float64) float64 {
	return 1.0 / (1 + math.Pow(10, (b-a)/100))
}

func EloChange(a, b, k, r float64) float64 {
	prediction := EloPredict(a, b)
	diff := float64(r) - prediction
	return k * diff
}
