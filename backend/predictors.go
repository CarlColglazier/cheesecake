package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
	"io/ioutil"
	"log"
)

/// Read the score cache from a file.
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

/// Predictor provides functions for updating results and
/// returning match predictions.
type Predictor interface {
	Predict(MatchEntry) map[string]interface{}
	AddResult(MatchEntry)
	CurrentValues() map[string]float64
	Dampen()
}

type EloScorePredictor struct {
	current map[string]float64
}

func NewEloScorePredictor() *EloScorePredictor {
	scores, err := ReadEloRecords()
	if err != nil {
		log.Println("Could not read Elo scores")
		return &EloScorePredictor{}
	}
	return &EloScorePredictor{scores}
}

func NewEloScorePredictorFromCache(scores map[string]interface{}) *EloScorePredictor {
	mapString := make(map[string]float64)
	for key, value := range scores {
		strKey := fmt.Sprintf("%v", key)
		//val := float64(value)
		val, ok := value.(float64)
		if !ok {
			val = 0.0
		}
		mapString[strKey] = val
	}
	return &EloScorePredictor{mapString}
}

func (pred *EloScorePredictor) Dampen() {
	for k, v := range pred.current {
		pred.current[k] = 0.5*v + 15
	}
}

func (pred *EloScorePredictor) CurrentValues() map[string]float64 {
	return pred.current
}

func (pred *EloScorePredictor) Predict(me MatchEntry) map[string]interface{} {
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

func (pred *EloScorePredictor) AddResult(me MatchEntry) {
	std := 21.1
	k := 12.0
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

type MarblePredictor struct {
	current map[string]float64
}

func NewMarblePredictor() *MarblePredictor {
	scores := map[string]float64{}
	return &MarblePredictor{scores}
}

func (mp *MarblePredictor) MarbleScore(teamKey string) float64 {
	if _, ok := mp.current[teamKey]; !ok {
		mp.current[teamKey] = 100.0
	}
	return mp.current[teamKey]
}

func (mp *MarblePredictor) teamMarbles(me MatchEntry) map[string]float64 {
	marbles := make(map[string]float64)
	marbles["red"] = 0.0
	marbles["blue"] = 0.0
	for key, val := range me.Alliances {
		for i := range val.Teams {
			marbles[key] += mp.MarbleScore(val.Teams[i])
		}
		//marbles[key] /= float64(len(val.Teams))
	}
	return marbles
}

func (mp *MarblePredictor) Predict(me MatchEntry) map[string]interface{} {
	marbles := mp.teamMarbles(me)
	ret := make(map[string]interface{})
	ret["red"] = marbles["red"] / (marbles["red"] + marbles["blue"])
	ret["blue"] = marbles["blue"] / (marbles["red"] + marbles["blue"])
	return ret
}

func (mp *MarblePredictor) AddResult(me MatchEntry) {
	marbles := mp.teamMarbles(me)
	for key, val := range me.Alliances {
		for i := range val.Teams {
			teamKey := val.Teams[i]
			if me.Match.WinningAlliance == "red" {
				if key == "red" {
					mp.current[teamKey] += 0.2 * marbles["blue"] / 3.0
				} else if key == "blue" {
					mp.current[teamKey] *= 0.8
				}
			} else if me.Match.WinningAlliance == "blue" {
				if key == "red" {
					mp.current[teamKey] *= 0.8
				} else if key == "blue" {
					mp.current[teamKey] += 0.2 * marbles["blue"] / 3.0
				}
			}
		}
	}
}

func (mp *MarblePredictor) CurrentValues() map[string]float64 {
	return mp.current
}

func (mp *MarblePredictor) Dampen() {
	// Do nothing
}
