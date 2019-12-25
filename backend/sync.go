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

func (config *Config) insertTeams(teamList []tba.Team) {
	var teams [][]interface{}
	for _, row := range teamList {
		teams = append(teams, []interface{}{
			row.Key,
			row.TeamNumber,
			row.Name,
		})
	}
	copyCount, _ := config.Conn.CopyFrom(
		pgx.Identifier{"team"},
		[]string{"key", "team_number", "name"},
		pgx.CopyFromRows(teams),
	)
	fmt.Println(copyCount)
}

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
	rows, _ := config.Tba.GetAllEventMatches(2019)
	var r [][]interface{}
	var a [][]interface{}
	var aTeams [][]interface{}
	for _, row := range rows {
		r = append(r, []interface{}{
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
		})
		a = append(a, []interface{}{
			row.Key + "_blue",
			row.Alliances.Blue.Score,
			"blue",
			row.Key,
		})
		for _, team := range row.Alliances.Blue.TeamKeys {
			aTeams = append(aTeams,
				[]interface{}{
					row.Key + "_blue",
					team,
				})
		}
		a = append(a, []interface{}{
			row.Key + "_red",
			row.Alliances.Red.Score,
			"red",
			row.Key,
		})
		for _, team := range row.Alliances.Red.TeamKeys {
			aTeams = append(aTeams,
				[]interface{}{
					row.Key + "_red",
					team,
				})
		}
	}
	copyCount, _ := config.Conn.CopyFrom(
		pgx.Identifier{"match"},
		[]string{
			"key", "comp_level", "set_number", "match_number",
			"winning_alliance", "event_key", "time", "actual_time",
			"predicted_time", "post_result_time", "score_breakdown",
		},
		pgx.CopyFromRows(r),
	)
	fmt.Println(copyCount)
	copyCount, _ = config.Conn.CopyFrom(
		pgx.Identifier{"alliance"},
		[]string{
			"key", "score", "color", "match_key",
		},
		pgx.CopyFromRows(a),
	)
	fmt.Println(copyCount)
	copyCount, err = config.Conn.CopyFrom(
		pgx.Identifier{"alliance_teams"},
		[]string{
			"alliance_id", "team_key",
		},
		pgx.CopyFromRows(aTeams),
	)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(copyCount)
	pred := NewEloScorePredictor()
	fmt.Println("Calculating elo scores...")
	_, err = calculatePredictor(config, pred, "eloscores")
	if err != nil {
		log.Println(err)
	}
}
