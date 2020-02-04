package main

import (
	"log"
	"math"
)

type BetaPredictor struct {
	success      map[string]int
	attempts     map[string]int
	a            float64
	b            float64
	breakdownKey string
}

func NewBetaPredictor(a, b float64, breakdownKey string) *BetaPredictor {
	succ := make(map[string]int)
	att := make(map[string]int)
	return &BetaPredictor{succ, att, a, b, breakdownKey}
}

/*
func NewBetaPredictorFromCache(key string) *BetaPredictor {
	succ := make(map[string]int)
	att := make(map[string]int)
	return &BetaPredictor{succ, att, a, b}
}
*/

func (pred *BetaPredictor) Dampen() {
	// Do nothing.
}

func (pred *BetaPredictor) CurrentValues() map[string]float64 {
	dic := make(map[string]float64)
	return dic
}

func (pred *BetaPredictor) getTeamBetaPred(key string) float64 {
	if _, ok := pred.success[key]; !ok {
		pred.success[key] = 0
		pred.attempts[key] = 0
	}
	// alpha plus hits over beta plus misses
	return ((pred.a + float64(pred.success[key])) /
		(pred.b + float64(pred.attempts[key]) - float64(pred.success[key])))
}

func (pred *BetaPredictor) Predict(me MatchEntry) map[string]interface{} {
	bpred := make(map[string]float64)
	for key, val := range me.Alliances {
		bpred[key] = 0.0
		for i := range val.Teams {
			bpred[key] = math.Max(bpred[key], pred.getTeamBetaPred(val.Teams[i]))
		}
	}
	// So that the types match.
	ret := make(map[string]interface{})
	for key, _ := range me.Alliances {
		ret[key] = bpred[key]
	}
	return ret
}

func (pred *BetaPredictor) AddResult(me MatchEntry) {
	breakdown := me.Match.ScoreBreakdown
	for key, val := range me.Alliances {
		bd, ok := breakdown[key].(map[string]interface{})
		if !ok {
			log.Println("Issue with breakdown.")
		}
		success, ok := bd[pred.breakdownKey].(bool)
		if !ok {
			log.Println("Issue casting success to bool.")
		}
		for i := range val.Teams {
			teamKey := val.Teams[i]
			pred.attempts[teamKey] += 1
			if success {
				pred.success[teamKey] += 1
			}
		}
	}
}
