package main

import (
	"context"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

func (me *MatchEntry) Diff() (int, error) {
	if _, ok := me.Alliances["red"]; !ok {
		return 0, errors.New("No red alliance")
	}
	if _, ok := me.Alliances["blue"]; !ok {
		return 0, errors.New("No blue alliance")
	}
	return me.Alliances["red"].Alliance.Score - me.Alliances["blue"].Alliance.Score, nil
}

func (config *Config) getEventMatches2019(event string) ([]MatchEntry, error) {
	rows, err := config.conn.Query(
		context.Background(),
		`SELECT "match".*, alliance.*, alliance_teams.*, ph.prediction as EloScorePrediction, phr.prediction as RocketPrediction, phh.prediction as HabPrediction, event.event_type < 7 as official FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
LEFT JOIN prediction_history ph on ph."match" = alliance.match_key and ph.model = 'eloscore2019'
LEFT JOIN prediction_history phr on phr."match" = alliance.match_key and phr.model = 'rocket'
LEFT JOIN prediction_history phh on phh."match" = alliance.match_key and phh.model = 'hab'
join event on (event.key = match.event_key)
where match.event_key = '`+event+`'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make(map[string]MatchEntry)
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		var eloPrediction PredictionHistory
		var rocketPrediction PredictionHistory
		var habPrediction PredictionHistory
		var official bool
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
			&aTeam.AllianceId,
			&aTeam.TeamKey,
			&eloPrediction.Prediction,
			&rocketPrediction.Prediction,
			&habPrediction.Prediction,
			&official,
		)
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			preds := make(map[string]*PredictionHistory)
			matches[match.Key] = MatchEntry{&match, dict, preds, official}
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
		matches[key].Predictions["elo_score"] = &eloPrediction
		matches[key].Predictions["rocket"] = &rocketPrediction
		matches[key].Predictions["hab"] = &habPrediction
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

func (config *Config) getEventMatches2020(event string) ([]MatchEntry, error) {
	rows, err := config.conn.Query(
		context.Background(),
		`SELECT "match".*, alliance.*, alliance_teams.*, ph.prediction as EloScorePrediction, phe.prediction as Energized, phs.prediction as Shield, event.event_type < 7 as official FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
LEFT JOIN prediction_history ph on ph."match" = alliance.match_key and ph.model = 'eloscore2020'
LEFT JOIN prediction_history phe on phe."match" = alliance.match_key and phe.model = 'shieldeng'
LEFT JOIN prediction_history phs on phs."match" = alliance.match_key and phs.model = 'shieldop'
join event on (event.key = match.event_key)
where match.event_key = '`+event+`'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make(map[string]MatchEntry)
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		var eloPrediction PredictionHistory
		var energizedPrediction PredictionHistory
		var shieldPrediction PredictionHistory
		var official bool
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
			&aTeam.AllianceId,
			&aTeam.TeamKey,
			&eloPrediction.Prediction,
			&energizedPrediction.Prediction,
			&shieldPrediction.Prediction,
			&official,
		)
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			preds := make(map[string]*PredictionHistory)
			matches[match.Key] = MatchEntry{&match, dict, preds, official}
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
		matches[key].Predictions["elo_score"] = &eloPrediction
		matches[key].Predictions["energized"] = &energizedPrediction
		matches[key].Predictions["shield"] = &shieldPrediction
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
	rows, err := config.conn.Query(
		context.Background(),
		`SELECT
  match.key, match.comp_level, match.set_number, match.match_number, match.winning_alliance, match.event_key, match.time, match.actual_time, match.predicted_time, match.post_result_time, match.score_breakdown,
  alliance.key, alliance.score, alliance.color, alliance.match_key,
  alliance_teams.alliance_id, alliance_teams.team_key,
  event.event_type < 7 as official
  FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
join event on (event.key = match.event_key)
--where event.event_type < 7`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make(map[string]MatchEntry)
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		var official bool
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
			&aTeam.AllianceId,
			&aTeam.TeamKey,
			&official,
		)
		// temp: This takes up too much memory.
		//match.ScoreBreakdown = nil
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			preds := make(map[string]*PredictionHistory)
			matches[match.Key] = MatchEntry{&match, dict, preds, official}
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

func (config *Config) getEvents(year int) ([]Event, error) {
	rows, err := config.conn.Query(
		context.Background(),
		"SELECT key, short_name FROM event where event.year="+strconv.Itoa(year))
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
