package main

import (
	"math/rand"
)

type future struct {
	rp map[string]int
}

// future constructor
func NewFuture() future {
	rp := make(map[string]int)
	return future{rp: rp}
}

// Helper function to add points for a team or add the team
// to the map if it does not exist.
func (f future) addPoints(team string, points int) {
	if _, ok := f.rp[team]; !ok {
		f.rp[team] = 0
	}
	f.rp[team] += points
}

func (f future) leader() string {
	val := ""
	m := 0
	for key, v := range f.rp {
		if v > m {
			val = key
			m = v
		}
	}
	return val
}

// Calculate one possible future for an event.
func findFuture(time int, matches []MatchEntry) future {
	f := NewFuture()
	for _, match := range matches {
		if match.Match.CompLevel != "qm" {
			continue
		}
		/*
			if match.Alliances["red"].Alliance.Score != -1 {
				if match.Alliances["red"].Alliance.Score > match.Alliances["blue"].Alliance.Score {

				}
			}
		*/
		if match.Match.Time < time {
			if len(match.Match.WinningAlliance) > 0 {
				winner := match.Match.WinningAlliance
				for _, team := range match.Alliances[winner].Teams {
					f.addPoints(team, 2)
				}
				continue
				// TODO: handle ties
			}
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
		r := rand.Float64()
		if r <= pred {
			randWin = "red"
		} else {
			randWin = "blue"
		}
		for _, team := range match.Alliances[randWin].Teams {
			f.addPoints(team, 2)
		}
		//fmt.Printf("%f %f\n", r, pred)
	}
	/*
		if matches[0].Match.EventKey == "2019ncwak" {
			fmt.Printf("%v\n", f)
		}
	*/
	return f
}

func (config *Config) forecastEvent(time int, matches []MatchEntry) map[string]int {
	futures := make([]future, 100)
	for i, _ := range futures {
		futures[i] = findFuture(time, matches)
	}
	leaders := make(map[string]int)
	for _, val := range futures {
		leader := val.leader()
		leaders[leader] += 1
	}
	return leaders
}

/*
func (config *Config) forecast() {
	events, err := config.getEvents(2019)
	if err != nil {
		log.Println("Could not get events.")
		return
	}
	for _, event := range events {
		matches, _ := config.getEventMatches2019(event.Key)
		fmt.Println(event.Key)
		config.forecastEvent(matches)
	}
}
*/
