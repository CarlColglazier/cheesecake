package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func ReadEloRecords() (map[string]float64, error) {
	file, err := ioutil.ReadFile("elo2018.json")
	if err != nil {
		return nil, err
	}
	var records map[string]float64
	err = json.Unmarshal([]byte(file), &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

type predictor interface {
	//predictMatch() float64
	predict() float64
	addResult()
	currentValues()
}

type EloScriptPredictor struct {
	current map[string]float64
}

func NewEloScriptPredictor() *EloScriptPredictor {
	scores, err := ReadEloRecords()
	if err != nil {
		log.Println("Could not read Elo scores")
		return &EloScriptPredictor{}
	}
	return &EloScriptPredictor{scores}
}

func (pred *EloScriptPredictor) currentValues() map[string]float64 {
	return pred.currentValues()
}

type RedPredictor struct{}

func (pred *RedPredictor) currentValues() {
	return
}
