package main

type predictor interface {
	predictMatch() float64
	predict() float64
	addResult()
	currentValues()
}

type EloScriptPredictor struct {
	current map[string]float64
}
