package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Read the score cache from a file.
func ReadEloRecords(year int) (map[string]float64, error) {
	fileName := fmt.Sprintf("elo%d.json", year-1)
	file, err := ioutil.ReadFile(fileName)
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

// Predictor provides functions for updating results and
// returning match predictions.
type Model interface {
	Predict(MatchEntry) map[string]interface{}
	AddResult(MatchEntry)
	CurrentValues() map[string]float64
	//Dampen()
	SupportsYear(year int) bool
}

type EloScoreModel struct {
	current map[string]float64
	year    int
	K       float64
	Std     float64
}

func NewEloScoreModel(year int, std float64) *EloScoreModel {
	scores, err := ReadEloRecords(year)
	if err != nil {
		log.Printf("Could not read Elo scores: %v\n", err)
		return &EloScoreModel{}
	}
	pred := &EloScoreModel{scores, year, 12.0, std}
	pred.Dampen()
	return pred
}

/*
func NewEloScoreModelFromCache(scores map[string]interface{}) *EloScoreModel {
	mapString := make(map[string]float64)
	for key, value := range scores {
		strKey := fmt.Sprintf("%v", key)
		val, ok := value.(float64)
		if !ok {
			val = 0.0
		}
		mapString[strKey] = val
	}
	return &EloScoreModel{mapString, 2019}
}
*/

func (pred *EloScoreModel) Dampen() {
	for k, v := range pred.current {
		pred.current[k] = 0.5*v + 15
	}
}

func (pred *EloScoreModel) SupportsYear(year int) bool {
	return year == pred.year
}

func (pred *EloScoreModel) CurrentValues() map[string]float64 {
	return pred.current
}

func (pred *EloScoreModel) Predict(me MatchEntry) map[string]interface{} {
	elos := make(map[string]float64)
	elos["red"] = 0.0
	elos["blue"] = 0.0
	for key, val := range me.Alliances {
		for i := range val.Teams {
			teamKey := val.Teams[i]
			if _, ok := pred.current[teamKey]; !ok {
				pred.current[teamKey] = 0.0
			}
			elos[key] += pred.current[teamKey]

		}
		elos[key] /= float64(len(val.Teams))
	}
	//return EloPredict(elos["red"], elos["blue"])
	red := EloPredict(elos["red"], elos["blue"])
	ret := make(map[string]interface{})
	ret["red"] = red
	ret["blue"] = 1 - red
	return ret
}

func (pred *EloScoreModel) AddResult(me MatchEntry) {
	std := pred.Std
	k := pred.K
	oddsMap := pred.Predict(me)
	odds, _ := oddsMap["red"].(float64)
	randx := rand.NewSource(372984243789)
	dist := distuv.Normal{0.0, std, randx}
	diff, err := me.Diff()
	if err != nil {
		return
	}
	expected := dist.Quantile(odds)
	change := k * (float64(diff) - expected) / std
	for key, val := range me.Alliances {
		for i := range val.Teams {
			teamKey := val.Teams[i]
			if _, ok := pred.current[teamKey]; !ok {
				pred.current[teamKey] = 0.0
			}
			if key == "red" {
				pred.current[teamKey] += change
			} else {
				pred.current[teamKey] -= change
			}
		}
	}
}
