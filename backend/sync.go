package main

import (
	"encoding/json"
	"fmt"
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
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

func calculateElo(config *Config) ([]byte, error) {
	log.Println("Could not fetch scores from cache. Calculating...")
	// TODO: Check the cache here. This runs multiple times otherwise.
	matches, err := config.getMatches()
	if err != nil {
		return nil, err
	}
	log.Println(len(matches))
	pred := NewEloScorePredictor()
	pred.Dampen()
	var predictions [][]interface{}
	for _, match := range matches {
		p := pred.Predict(match)
		predictions = append(predictions, []interface{}{match.Match.Key, p, "elo_score"})
		pred.AddResult(match)
	}
	config.Conn.CopyFrom(
		pgx.Identifier{"prediction_history"},
		[]string{"match", "prediction", "model"},
		pgx.CopyFromRows(predictions),
	)
	j, err := json.Marshal(pred.CurrentValues())
	return j, err
}

func reset(config *Config) {
	config.Migrate("db", "cheesecake")
	teamList, err := config.Tba.GetAllTeams()
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	fmt.Println(copyCount)
}
