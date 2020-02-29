package main

import (
	"log"
	"math/rand"
	"sort"
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

// From https://github.com/indraniel/go-learn/blob/master/09-sort-map-keys-by-values.go
type pair struct {
	key   string
	value int
}

type pairList []pair

func (p pairList) Len() int           { return len(p) }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pairList) Less(i, j int) bool { return p[i].value < p[j].value }

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

func (f future) rankOrder() []string {
	p := make(pairList, len(f.rp))
	i := 0
	for k, v := range f.rp {
		p[i] = pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	var keys []string
	for _, k := range p {
		keys = append(keys, k.key)
	}
	return keys
}

// Calculate one possible future for an event.
func findFuture(time int, matches []MatchEntry) future {
	f := NewFuture()
	for _, match := range matches {
		if match.Match.CompLevel != "qm" {
			continue
		}
		if match.Match.Time <= time {
			if len(match.Match.WinningAlliance) > 0 {
				winner := match.Match.WinningAlliance
				for _, team := range match.Alliances[winner].Teams {
					f.addPoints(team, 2)
				}
				// TODO: handle ties
			}
			// RP things by year.
			if match.Match.Key[0:4] == "2020" {
				breakdown := match.Match.ScoreBreakdown
				for key, val := range match.Alliances {
					bd, ok := breakdown[key].(map[string]interface{})
					if !ok {
						continue
					}
					success, ok := bd["shieldOperationalRankingPoint"].(bool)
					if !ok {
						continue
					}
					if success {
						for i := range val.Teams {
							teamKey := val.Teams[i]
							f.addPoints(teamKey, 1)
						}
					}
					success, ok = bd["shieldEnergizedRankingPoint"].(bool)
					if !ok {
						continue
					}
					if success {
						for i := range val.Teams {
							teamKey := val.Teams[i]
							f.addPoints(teamKey, 1)
						}
					}
				}
			}
			if match.Match.Key[0:4] == "2019" {
				breakdown := match.Match.ScoreBreakdown
				for key, val := range match.Alliances {
					bd, ok := breakdown[key].(map[string]interface{})
					if !ok {
						continue
					}
					success, ok := bd["completeRocketRankingPoint"].(bool)
					if !ok {
						continue
					}
					if success {
						for i := range val.Teams {
							teamKey := val.Teams[i]
							f.addPoints(teamKey, 1)
						}
					}
					success, ok = bd["habDockingRankingPoint"].(bool)
					if !ok {
						continue
					}
					if success {
						for i := range val.Teams {
							teamKey := val.Teams[i]
							f.addPoints(teamKey, 1)
						}
					}
				}
			}
		} else {
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
			if match.Match.Key[0:4] == "2019" {
				rocket := match.Predictions["rocket"]
				hab := match.Predictions["hab"]
				if rocket == nil || hab == nil {
					log.Println("Issue getting predictors.")
					continue
				}
				for key, val := range rocket.Prediction {
					k := val.(float64)
					if rand.Float64() < k {
						for _, team := range match.Alliances[key].Teams {
							f.addPoints(team, 1)
						}
					}
				}
				for key, val := range hab.Prediction {
					k := val.(float64)
					if rand.Float64() < k {
						for _, team := range match.Alliances[key].Teams {
							f.addPoints(team, 1)
						}
					}
				}
			}
		}
	}
	return f
}

func (config *Config) forecastEvent(time int, matches []MatchEntry) (map[string]int, map[string]int) {
	futures := make([]future, 100)
	for i, _ := range futures {
		futures[i] = findFuture(time, matches)
	}
	leaders := make(map[string]int)
	captains := make(map[string]int)
	for _, val := range futures {
		leader := val.leader()
		c := val.rankOrder()
		mc := 8
		if len(c) < 8 {
			mc = len(c)
		}
		caps := c[0:mc]
		leaders[leader] += 1
		for _, c := range caps {
			captains[c] += 1
		}
	}
	return leaders, captains
}
