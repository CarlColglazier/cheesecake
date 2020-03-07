package main

import "strconv"

// Match represents a single match in both the database and the API.
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

// Alliance represents a collection of AllianceTeam objects,
// including the results in the match.
type Alliance struct {
	Key      string `db:"key" json:"key"`
	Score    int    `db:"score" json:"score"`
	Color    string `db:"color" json:"color"`
	MatchKey string `db:"match_key" json:"match_key"`
}

// AllianceTeam maps Team objects to Alliance objects.
type AllianceTeam struct {
	AllianceId string `db:"alliance_id" json:"alliance_id"`
	TeamKey    string `db:"team_key" json:"team_key"`
}

//
type AllianceEntry struct {
	Alliance Alliance `json:"alliance"`
	Teams    []string `json:"teams"`
}

//
type MatchEntry struct {
	Match       *Match                        `json:"match"`
	Alliances   map[string]*AllianceEntry     `json:"alliances"`
	Predictions map[string]*PredictionHistory `json:"predictions"`
	Official    bool
}

func (me MatchEntry) year() int {
	yearStr := me.Match.Key[0:4]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return 0
	}
	return year
}

func (m MatchEntry) played() bool {
	return m.Alliances["blue"].Alliance.Score > 0 &&
		m.Alliances["red"].Alliance.Score > 0
}

//
type PredictionHistory struct {
	//Key        int     `db:"key"`
	//Match      string          `db:"match" json:"match"`
	Prediction map[string]interface{} `db:"prediction" json:"prediction"`
	//Model      string          `db:"model" json:"model"`
}

type ForecastHistory struct {
	Model    string                 `db:"model" json:"model"`
	MatchKey string                 `db:"match_key" json:"match_key"`
	TeamKey  string                 `db:"team_key" json:"team_key"`
	Forecast map[string]interface{} `db:"forecast" json:"forecast"`
}

//
type Event struct {
	ShortName *string `json:"short_name"`
	Key       string  `json:"key"`
	EndDate   string  `json:"end_date"`
}
