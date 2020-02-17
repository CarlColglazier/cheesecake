package main

import (
	"fmt"
	"log"
	"math/rand"
)

type future struct {
	rp map[string]int
}

func NewFuture() future {
	rp := make(map[string]int)
	return future{rp: rp}
}

func (f future) addPoints(team string, points int) {
	if _, ok := f.rp[team]; !ok {
		f.rp[team] = 0
	}
	f.rp[team] += points
}

func findFuture(matches []MatchEntry) future {
	f := NewFuture()
	for _, match := range matches {
		if match.Match.CompLevel != "qm" {
			continue
		}
		var randWin string
		var pred float64
		p := match.Predictions["elo_score"]
		if p == nil {
			continue
		}
		if val, ok := p.Prediction["red"]; ok {
			pred = val.(float64)
		} else {
			pred = 0.5
		}
		if rand.Float64() < pred {
			randWin = "red"
		} else {
			randWin = "blue"
		}
		for _, team := range match.Alliances[randWin].Teams {
			f.addPoints(team, 2)
		}
	}
	return f
}

func (config *Config) forecast() {
	events, err := config.getEvents(2019)
	if err != nil {
		log.Println("Could not get events.")
		return
	}
	for _, event := range events {
		futures := make([]future, 1000)
		matches, _ := config.getEventMatches2019(event.Key)
		for i, _ := range futures {
			futures[i] = findFuture(matches)
		}
		fmt.Printf("%s: %v\n", event.Key, futures[0])
	}
}
