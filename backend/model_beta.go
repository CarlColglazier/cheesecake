package main

import (
	"log"
	"math"
	"strconv"
)

type BetaModel struct {
	success      map[string]int
	attempts     map[string]int
	a            float64
	b            float64
	breakdownKey string
	year         int
}

func NewBetaModel(a, b float64, breakdownKey string, year int) *BetaModel {
	succ := make(map[string]int)
	att := make(map[string]int)
	return &BetaModel{succ, att, a, b, breakdownKey, year}
}

/*
func NewBetaModelFromCache(key string) *BetaModel {
	succ := make(map[string]int)
	att := make(map[string]int)
	return &BetaModel{succ, att, a, b}
}
*/

func (pred *BetaModel) SupportsYear(year int) bool {
	return year == pred.year
}

func (pred *BetaModel) CurrentValues() map[string]float64 {
	dic := make(map[string]float64)
	return dic
}

// Shortcut function to return a team's current values, or create
// a new, blank entry if it does not exist.
func (pred *BetaModel) getTeamBetaPred(key string) float64 {
	if _, ok := pred.success[key]; !ok {
		pred.success[key] = 0
		pred.attempts[key] = 0
	}
	// alpha plus hits over beta plus misses
	return ((pred.a + float64(pred.success[key])) /
		(pred.a + pred.b + float64(pred.attempts[key])))
}

func (pred *BetaModel) Predict(me MatchEntry) map[string]interface{} {
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

func (pred *BetaModel) AddResult(me MatchEntry) {
	if me.Match.Key[0:4] != strconv.Itoa(pred.year) {
		return
	}
	if me.Match.CompLevel != "qm" {
		return
	}
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
