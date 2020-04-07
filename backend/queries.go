package main

import (
	"context"
	"sort"
	"strconv"
)

func (config *Config) predictionQuery(query string) (map[string]map[string]*PredictionHistory, error) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	v := make(map[string]map[string]*PredictionHistory)
	for rows.Next() {
		var match string
		var model string
		var pred PredictionHistory
		rows.Scan(
			&match,
			&model,
			&pred.Prediction,
		)
		if _, ok := v[match]; !ok {
			v[match] = make(map[string]*PredictionHistory)
		}
		v[match][model] = &pred
	}
	return v, nil
}

func (config *Config) matchQuery(query string) ([]MatchEntry, error) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		query,
	)
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
			&aTeam.Position,
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

func (config *Config) getEventMatches(event string) ([]MatchEntry, error) {
	query := `SELECT "match".*, alliance.*, alliance_teams.*, event.event_type < 7 as official FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
join event on (event.key = match.event_key)
where match.event_key = '` + event + `'`
	matches, err := config.matchQuery(query)
	if err != nil {
		return nil, err
	}
	preds, err := config.predictionQuery(`
		select ph.match, ph.model, ph.prediction from prediction_history ph
		join match on (match.key = ph.match)
    where match.event_key = '` + event + `'`)
	if err != nil {
		return matches, err
	}
	for _, m := range matches {
		for p_key, p := range preds[m.Match.Key] {
			m.Predictions[p_key] = p
		}
	}
	return matches, nil
}

func (config *Config) getTeamMatchesYear(team, year string) ([]MatchEntry, error) {
	query := `
SELECT "match".*, alliance.*, alliance_teams.*, event.event_type < 7 as official FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
join event on (event.key = match.event_key)
where match.key in
(select match."key" from alliance_teams at2
join alliance on (alliance.key = at2.alliance_id)
join match on (match.key = alliance.match_key)
where at2.team_key = '` + team + `' and match.key like '` + year + `%')`
	matches, err := config.matchQuery(query)
	if err != nil {
		return nil, err
	}
	preds, err := config.predictionQuery(`
select ph.match, ph.model, ph.prediction from prediction_history ph
join match on (match.key = ph.match)
where match.key in
(select match."key" from alliance_teams at2
join alliance on (alliance.key = at2.alliance_id)
join match on (match.key = alliance.match_key)
where at2.team_key = '` + team + `' and match.key like '` + year + `%')`)
	if err != nil {
		return matches, err
	}

	for _, m := range matches {
		for p_key, p := range preds[m.Match.Key] {
			m.Predictions[p_key] = p
		}
	}
	return matches, nil
}

func (config *Config) getMatches() ([]MatchEntry, error) {
	query := `SELECT
  match.key, match.comp_level, match.set_number, match.match_number, match.winning_alliance, match.event_key, match.time, match.actual_time, match.predicted_time, match.post_result_time, match.score_breakdown,
  alliance.key, alliance.score, alliance.color, alliance.match_key,
  alliance_teams.alliance_id, alliance_teams.team_key, alliance_teams.position,
  event.event_type < 7 as official
  FROM match
JOIN alliance on (match.key = alliance.match_key)
JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)
join event on (event.key = match.event_key)`
	return config.matchQuery(query)
}

func (config *Config) getEvents(year int) ([]Event, error) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		"SELECT key, short_name, end_date FROM event where event.year="+strconv.Itoa(year))
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
			&event.EndDate,
		)
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return events, nil
}

type ForecastEntry struct {
	Model    string  `json:"model"`
	Match    int     `json:"match"`
	Team     string  `json:"team"`
	Forecast float64 `json:"forecast"`
}

func (config *Config) getEventForecasts(event string) ([]ForecastEntry, error) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		`select model, match.match_number, team_key, forecast from forecast_history fh
join match on (match.key = fh.match_key )
where
fh.match_key like '`+event+`_%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	forecasts := make([]ForecastEntry, 0)
	for rows.Next() {
		var cast ForecastEntry
		rows.Scan(
			&cast.Model,
			&cast.Match,
			&cast.Team,
			&cast.Forecast,
		)
		forecasts = append(forecasts, cast)
	}
	return forecasts, nil
}

func (config *Config) getTeamEventBreakdown2020(team, event string) (map[string][]int, error) {
	query := `
select score_breakdown->color->'autoCellPoints' as autoCells,
(case
	when (score_breakdown->color->format('initLineRobot%s', (position+1)))::text = '"Exited"' 
	then 5
	else 0
end) as init,
score_breakdown->color->'teleopCellPoints' as teleCells,
score_breakdown->color->format('endgameRobot%s', (position+1)) as endgame
from alliance_teams at2 
join alliance on (alliance."key" = at2.alliance_id )
join match on (match.key = alliance.match_key )
where at2.team_key =$1 and event_key =$2
order by match.actual_time
`
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		query,
		team,
		event,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string][]int)
	for rows.Next() {
		var auto, init, tele int
		var end string
		rows.Scan(
			&auto,
			&init,
			&tele,
			&end,
		)
		ret["autoCells"] = append(ret["autoCells"], auto)
		ret["init"] = append(ret["init"], init)
		ret["teleCells"] = append(ret["teleCells"], tele)
	}
	return ret, nil
}

func (config *Config) getEventBreakdown2020(event string) (map[string][]int, error) {
	query := `
select score_breakdown->color->'autoCellPoints' as autoCells,
(case
	when (score_breakdown->color->format('initLineRobot%s', (position+1)))::text = '"Exited"' 
	then 5
	else 0
end) as init,
score_breakdown->color->'teleopCellPoints' as teleCells,
score_breakdown->color->format('endgameRobot%s', (position+1)) as endgame
from alliance_teams at2 
join alliance on (alliance."key" = at2.alliance_id )
join match on (match.key = alliance.match_key )
where at2.team_key =$1 and event_key =$2
order by match.actual_time
`
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(
		context.Background(),
		query,
		event,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make(map[string][]int)
	for rows.Next() {
		var auto, init, tele int
		var end string
		rows.Scan(
			&auto,
			&init,
			&tele,
			&end,
		)
		ret["autoCells"] = append(ret["autoCells"], auto)
		ret["init"] = append(ret["init"], init)
		ret["teleCells"] = append(ret["teleCells"], tele)
	}
	return ret, nil
}
