package main

import (
	"testing"
)

const float64Tresh = 1e-9

func TestEloPredict(t *testing.T) {
	var tests = []struct {
		a    float64
		b    float64
		want float64
	}{
		{0, 0, 0.5},
		{25, 0, 0.640065},
		{0, 25, 0.359935},
		{1025, 1000, 0.640065},
		{1000, 1025, 0.359935},
		{-25, 0, 0.359935},
	}
	for _, test := range tests {
		if got := EloPredict(test.a, test.b); got > test.want+float64Tresh || got < test.want-float64Tresh {
			t.Errorf("EloPredict(%f, %f) = %f, expected %f", test.a, test.b, got, test.want)
		}
	}
}
