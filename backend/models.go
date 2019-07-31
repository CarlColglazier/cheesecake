package main

import (
	"github.com/pkg/errors"
	"sort"
)

type Match struct {
	Key             string                 `db:"key"`
	CompLevel       string                 `db:"comp_level"`
	SetNumber       int                    `db:"set_number"`
	MatchNumber     int                    `db:"match_number"`
	WinningAlliance string                 `db:"winning_alliance"`
	EventKey        string                 `db:"event_key"`
	Time            int                    `db:"time"`
	ActualTime      int                    `db:"actual_time"`
	PredictedTime   int                    `db:"predicted_time"`
	PostResultTime  int                    `db:"post_result_time"`
	ScoreBreakdown  map[string]interface{} `db:"score_breakdown"`
}

type Alliance struct {
	Key      string `db:"key"`
	Score    int    `db:"score"`
	Color    string `db:"color"`
	MatchKey string `db:"match_key"`
}

type AllianceTeam struct {
	Position   int    `db:"position"`
	AllianceId string `db:"alliance_id"`
	TeamKey    string `db:"team_key"`
}

type AllianceEntry struct {
	Alliance Alliance
	Teams    []string
}

type MatchEntry struct {
	Match     Match
	Alliances map[string]*AllianceEntry
}

type PredictionHistory struct {
	//Key        int     `db:"key"`
	Match      string  `db:"match"`
	Prediction float64 `db:"prediction"`
	Model      string  `db:"model"`
}

type Event struct {
	//ShortName string
	Key string
}

func (me *MatchEntry) Diff() (int, error) {
	if _, ok := me.Alliances["red"]; !ok {
		return 0, errors.New("No red alliance")
	}
	if _, ok := me.Alliances["blue"]; !ok {
		return 0, errors.New("No blue alliance")
	}
	return me.Alliances["red"].Alliance.Score - me.Alliances["blue"].Alliance.Score, nil
}

func (config *Config) getMatches() ([]MatchEntry, error) {
	rows, err := config.Conn.Query(`SELECT * FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
where match.event_key like '2019%'
--order by actual_time ASC, time ASC`)
	if err != nil {
		return nil, err
	}
	matches := make(map[string]MatchEntry)
	//var keys []string
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		rows.Scan(
			&match.Key,
			&match.CompLevel,
			&match.SetNumber,
			&match.MatchNumber,
			&match.WinningAlliance,
			&match.EventKey,
			&match.Time,
			&match.ActualTime,
			&match.PredictedTime,
			&match.PostResultTime,
			&match.ScoreBreakdown,
			&alliance.Key,
			&alliance.Score,
			&alliance.Color,
			&alliance.MatchKey,
			&aTeam.Position,
			&aTeam.AllianceId,
			&aTeam.TeamKey,
		)
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			matches[match.Key] = MatchEntry{match, dict}
		}
		if _, ok := matches[match.Key].Alliances[alliance.Color]; !ok {
			list := make([]string, 0)
			matches[match.Key].Alliances[alliance.Color] = &AllianceEntry{alliance, list}
		}
		matches[match.Key].Alliances[alliance.Color].Teams = append(
			matches[match.Key].Alliances[alliance.Color].Teams,
			aTeam.TeamKey,
		)
		/*
			if len(keys) == 0 || keys[len(keys)-1] != match.Key {
				keys = append(keys, match.Key)
			}
		*/
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	var list []MatchEntry
	for _, value := range matches {
		list = append(list, value)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Match.Time < list[j].Match.Time
	})
	return list, nil
}

func (config *Config) getEvents() ([]Event, error) {
	rows, err := config.Conn.Query(`SELECT key FROM event
where event.key like '2019%'`)
	if err != nil {
		return nil, err
	}
	var events []Event
	for rows.Next() {
		var event Event
		rows.Scan(
			//&event.ShortName,
			&event.Key,
		)
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return events, nil
}
