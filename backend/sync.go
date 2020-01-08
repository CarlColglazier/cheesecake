package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"log"
	"sync"
)

// Upserts a list of teams into the database.
func (config *Config) insertTeams(teamList []tba.Team) {
	batch := config.Conn.BeginBatch()
	for _, row := range teamList {
		batch.Queue(
			"insert into team (key, team_number, name) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT team_pkey DO UPDATE set team_number = $2, name = $3",
			[]interface{}{row.Key, row.TeamNumber, row.Name},
			[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID, pgtype.VarcharOID},
			nil,
		)
	}
	err := batch.Send(context.Background(), nil)
	if err != nil {
		log.Printf("Error sending batch: %s", err)
	}
	for i := 0; i < len(teamList); i++ {
		_, err := batch.ExecResults()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err = batch.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

// Upserts a list of events into the database.
func (config *Config) insertEvents(eventList []tba.Event) {
	var r [][]interface{}
	for _, row := range eventList {
		r = append(r, []interface{}{row.EndDate, row.Key, row.ShortName, row.Year})
	}
	_, err := config.Conn.CopyFrom(
		pgx.Identifier{"event"},
		[]string{"end_date", "key", "short_name", "year"},
		pgx.CopyFromRows(r),
	)
	if err != nil {
		log.Println(err)
	}
}

func (config *Config) insertMatches(matchList []tba.Match) {
	batch := config.Conn.BeginBatch()
	for _, row := range matchList {
		batch.Queue(
			"insert into match (key, comp_level, set_number, match_number, winning_alliance, event_key, time, actual_time, predicted_time, post_result_time, score_breakdown) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT ON CONSTRAINT match_pkey DO NOTHING",
			[]interface{}{
				row.Key,
				row.CompLevel,
				row.SetNumber,
				row.MatchNumber,
				row.WinningAlliance,
				row.EventKey,
				row.Time,
				row.ActualTime,
				row.PredictedTime,
				row.PostResultTime,
				row.ScoreBreakdown,
			},
			[]pgtype.OID{
				pgtype.VarcharOID,
				pgtype.VarcharOID,
				pgtype.Int4OID,
				pgtype.Int4OID,
				pgtype.VarcharOID,
				pgtype.VarcharOID,
				pgtype.Int4OID,
				pgtype.Int4OID,
				pgtype.Int4OID,
				pgtype.Int4OID,
				pgtype.JSONOID,
			},
			nil,
		)
		allianceId := row.Key + "_red"
		batch.Queue(
			"insert into alliance (key, score, color, match_key) values ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT alliance_pkey DO UPDATE set score = $2",
			[]interface{}{
				allianceId,
				row.Alliances.Red.Score,
				"red",
				row.Key,
			},
			[]pgtype.OID{
				pgtype.VarcharOID,
				pgtype.Int4OID,
				pgtype.VarcharOID,
				pgtype.VarcharOID,
			},
			nil,
		)
		for _, team := range row.Alliances.Red.TeamKeys {
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key) values ($1, $2) ON CONFLICT DO NOTHING",
				[]interface{}{
					allianceId,
					team,
				},
				[]pgtype.OID{
					pgtype.VarcharOID,
					pgtype.VarcharOID,
				},
				nil,
			)
		}
		allianceId = row.Key + "_blue"
		batch.Queue(
			"insert into alliance (key, score, color, match_key) values ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT alliance_pkey DO UPDATE set score = $2",
			[]interface{}{
				allianceId,
				row.Alliances.Blue.Score,
				"blue",
				row.Key,
			},
			[]pgtype.OID{
				pgtype.VarcharOID,
				pgtype.Int4OID,
				pgtype.VarcharOID,
				pgtype.VarcharOID,
			},
			nil,
		)
		for _, team := range row.Alliances.Blue.TeamKeys {
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key) values ($1, $2) ON CONFLICT DO NOTHING",
				[]interface{}{
					allianceId,
					team,
				},
				[]pgtype.OID{
					pgtype.VarcharOID,
					pgtype.VarcharOID,
				},
				nil,
			)
		}
	}
	err := batch.Send(context.Background(), nil)
	if err != nil {
		log.Printf("Error sending batch: %s", err)
	}
	for i := 0; i < len(matchList)*9; i++ {
		_, err := batch.ExecResults()
		if err != nil {
			log.Printf("Error upserting: %s", err)
		}
	}
	err = batch.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

func (config *Config) syncEvents() {
	var wg sync.WaitGroup
	for i := 2003; i <= 2019; i++ {
		go func(year int) {
			defer wg.Done()
			rows, _ := config.Tba.GetAllEvents(year)
			config.insertEvents(rows)
		}(i)

		wg.Add(1)
	}
	wg.Wait()
}

func calculatePredictor(config *Config, pred Predictor, modelkey string) ([]byte, error) {
	vals, err := config.CacheGet(modelkey)
	if err != nil {
		log.Println(err)
		log.Println("Could not fetch scores from cache. Calculating...")
	} else if len(vals) > 2 {
		b, err := json.Marshal(vals)
		if err != nil {
			log.Println("Could not marshal Elo win ratings")
		} else {
			return b, nil
		}
	} else {
		log.Println("Empty response. Calculating...")
	}
	matches, err := config.getMatches()
	if err != nil {
		return nil, err
	}
	pred.Dampen()
	batch := config.Conn.BeginBatch()
	for _, match := range matches {
		p := pred.Predict(match)
		batch.Queue("insert into prediction_history (match, model, prediction) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT prediction_history_pkey DO UPDATE set prediction = $3",
			[]interface{}{match.Match.Key, modelkey, p},
			[]pgtype.OID{pgtype.VarcharOID, pgtype.VarcharOID, pgtype.Float8OID},
			nil,
		)
		pred.AddResult(match)
	}
	err = batch.Send(context.Background(), nil)
	if err != nil {
		log.Printf("Error sending batch: %s", err)
	}
	for i := 0; i < len(matches); i++ {
		_, err := batch.ExecResults()
		if err != nil {
			log.Printf("Error upserting %s: %s", modelkey, err)
		}
	}
	err = batch.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
	values := pred.CurrentValues()
	j, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	err = config.CacheSetStr(modelkey, string(j))
	if err != nil {
		log.Printf("Could not set '%s' in cache: %s", modelkey, err)
	}
	return j, nil
}

func reset(config *Config) {
	config.Migrate("db", "cheesecake")
	teamList, err := config.Tba.GetAllTeams()
	if err != nil {
		log.Printf("Error getting teams: %s", err)
		return
	}
	config.insertTeams(teamList)
	config.syncEvents()
	matchChan, _, numEvents := config.Tba.GetAllEventMatches(2019)
	for i := 0; i < numEvents; i++ {
		matches := <-matchChan
		config.insertMatches(matches)
	}
	pred := NewEloScorePredictor()
	fmt.Println("Calculating elo scores...")
	_, err = calculatePredictor(config, pred, "eloscores")
	if err != nil {
		log.Println(err)
	}
}
