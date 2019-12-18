package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
	"os"
	"sort"
	"time"
)

func Connect(applicationName string) (conn *pgx.ConnPool) {
	var runtimeParams map[string]string
	runtimeParams = make(map[string]string)
	runtimeParams["application_name"] = applicationName
	connConfig := pgx.ConnConfig{
		User:              "postgres",
		Password:          "postgres",
		Host:              "db",
		Port:              5432,
		Database:          "cheesecake",
		TLSConfig:         nil,
		UseFallbackTLS:    false,
		FallbackTLSConfig: nil,
		RuntimeParams:     runtimeParams,
	}
	connPoolConfig := pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 8}
	errors := 0
	for errors < 10 {
		conn, err := pgx.NewConnPool(connPoolConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to establish connection: %v\n", err)
			time.Sleep(2000 * time.Millisecond)
			errors += 1
		} else {
			return conn
		}

	}
	os.Exit(1)
	return nil
}

type Match struct {
	Key             string                 `db:"key" json:"key"`
	CompLevel       string                 `db:"comp_level" json:"comp_level"`
	SetNumber       int                    `db:"set_number" json:"set_number"`
	MatchNumber     int                    `db:"match_number" json:"match_number"`
	WinningAlliance string                 `db:"winning_alliance" json:"winning_alliance"`
	EventKey        string                 `db:"event_key" json:"event_key"`
	Time            int                    `db:"time" json:"time"`
	ActualTime      int                    `db:"actual_time" json:"actual_time"`
	PredictedTime   int                    `db:"predicted_time" json:"predicted_time"`
	PostResultTime  int                    `db:"post_result_time" json:"post_result_time"`
	ScoreBreakdown  map[string]interface{} `db:"score_breakdown" json:"score_breakdown"`
}

type Alliance struct {
	Key      string `db:"key" json:"key"`
	Score    int    `db:"score" json:"score"`
	Color    string `db:"color" json:"color"`
	MatchKey string `db:"match_key" json:"match_key"`
}

type AllianceTeam struct {
	Position   int    `db:"position" json:"position"`
	AllianceId string `db:"alliance_id" json:"alliance_id"`
	TeamKey    string `db:"team_key" json:"team_key"`
}

type AllianceEntry struct {
	Alliance Alliance `json:"alliance"`
	Teams    []string `json:"teams"`
}

type MatchEntry struct {
	Match       *Match                        `json:"match"`
	Alliances   map[string]*AllianceEntry     `json:"alliances"`
	Predictions map[string]*PredictionHistory `json:"predictions"`
}

type PredictionHistory struct {
	//Key        int     `db:"key"`
	//Match      string          `db:"match" json:"match"`
	Prediction null.Float `db:"prediction" json:"prediction"`
	//Model      string          `db:"model" json:"model"`
}

type Event struct {
	ShortName *string `json:"short_name"`
	Key       string  `json:"key"`
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

func (config *Config) getEventMatches(event string) ([]MatchEntry, error) {
	rows, err := config.Conn.Query( //`SELECT * FROM match
		//JOIN alliance on (match.key = alliance.match_key)
		//JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
		//where match.event_key = '` + event + `'`)
		`SELECT "match".*, alliance.*, alliance_teams.*, ph.prediction as EloScorePrediction FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
LEFT JOIN prediction_history ph on ph."match" = alliance.match_key and ph.model = 'elo_score'
where match.event_key = '` + event + `'`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make(map[string]MatchEntry)
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		var prediction PredictionHistory
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
			&prediction.Prediction,
		)
		// temp: This takes up too much memory.
		//match.ScoreBreakdown = nil
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			preds := make(map[string]*PredictionHistory)
			matches[match.Key] = MatchEntry{&match, dict, preds}
		}
		key := match.Key
		if _, ok := matches[key].Alliances[alliance.Color]; !ok {
			list := make([]string, 0)
			matches[key].Alliances[alliance.Color] = &AllianceEntry{alliance, list}
		}
		matches[key].Alliances[alliance.Color].Teams = append(
			matches[key].Alliances[alliance.Color].Teams,
			aTeam.TeamKey,
		)
		//prediction.Model = "EloScore"
		matches[key].Predictions["elo_score"] = &prediction
		//var score float64
		//if err := prediction.Prediction.Scan(&score); !err {
		//	matches[key].Predictions["EloScore"] =
		//}/ else {
		//	return nil, err
		//}
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

func (config *Config) getMatches() ([]MatchEntry, error) {
	rows, err := config.Conn.Query(`SELECT * FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
where match.event_key like '2019%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make(map[string]MatchEntry)
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
		// temp: This takes up too much memory.
		match.ScoreBreakdown = nil
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			preds := make(map[string]*PredictionHistory)
			matches[match.Key] = MatchEntry{&match, dict, preds}
		}
		key := match.Key
		if _, ok := matches[key].Alliances[alliance.Color]; !ok {
			list := make([]string, 0)
			matches[key].Alliances[alliance.Color] = &AllianceEntry{alliance, list}
		}
		matches[key].Alliances[alliance.Color].Teams = append(
			matches[key].Alliances[alliance.Color].Teams,
			aTeam.TeamKey,
		)
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
	rows, err := config.Conn.Query(`SELECT key, short_name FROM event
where event.key like '2019%'`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var events []Event
	for rows.Next() {
		var event Event
		rows.Scan(
			&event.Key,
			&event.ShortName,
		)
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return events, nil
}
