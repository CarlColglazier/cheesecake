package main

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"sync"

	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx/pgtype"
)

// Upserts a list of teams into the database.
func (config *Config) insertTeams(teamList []tba.Team) {
	batch := config.conn.BeginBatch()
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
	batch := config.conn.BeginBatch()
	for _, row := range eventList {
		batch.Queue(
			"INSERT INTO event (key, end_date, event_type, short_name, year) values ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT event_pkey DO UPDATE set end_date = $2, event_type = $3, short_name = $4, year = $5",
			[]interface{}{row.Key, row.EndDate, row.EventType, row.ShortName, row.Year},
			[]pgtype.OID{pgtype.VarcharOID, pgtype.VarcharOID, pgtype.Int4OID, pgtype.VarcharOID, pgtype.Int4OID},
			nil,
		)
	}
	err := batch.Send(context.Background(), nil)
	if err != nil {
		log.Printf("Error sending batch: %s", err)
	}
	for i := 0; i < len(eventList); i++ {
		_, err := batch.ExecResults()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err = batch.Close()
	if err != nil {
		log.Println("Error closing batch.")
		log.Println(err)
	}
}

func (config *Config) version() int {
	row, err := config.conn.Query("select \"version\" from schema_migrations")
	if err != nil {
		return 0
	}
	defer row.Close()
	row.Next()
	var n int
	row.Scan(&n)
	return n
}

func (config *Config) insertMatches(matchList []tba.Match) {
	batch := config.conn.BeginBatch()
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Println(err)
	}
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
			if reg.Match([]byte(team[3:])) {
				//log.Println(team)
				team = "frc9990"
				//log.Println(team)
			}
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
			if reg.Match([]byte(team[3:])) {
				team = "frc9990"
			}
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
	err = batch.Send(context.Background(), nil)
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
	for i := 2003; i <= 2020; i++ {
		go func(year int) {
			defer wg.Done()
			rows, _ := config.tba.GetAllOfficialEvents(year)
			config.insertEvents(rows)
		}(i)

		wg.Add(1)
	}
	wg.Wait()
}

// TODO: Change this loop to run within config?
func calculateModel(config *Config, pred Model, modelkey string) ([]byte, error) {
	matches, err := config.getMatches()
	if err != nil {
		return nil, err
	}
	//pred.Dampen()
	batch := config.conn.BeginBatch()
	for _, match := range matches {
		p := pred.Predict(match)
		batch.Queue("insert into prediction_history (match, model, prediction) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT prediction_history_pkey DO UPDATE set prediction = $3",
			[]interface{}{match.Match.Key, modelkey, p},
			[]pgtype.OID{pgtype.VarcharOID, pgtype.VarcharOID, pgtype.JSONOID},
			nil,
		)
		if match.Official {
			pred.AddResult(match)
		}
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
	teamList, err := config.tba.GetAllTeams()
	if err != nil {
		log.Printf("Error getting teams: %s", err)
		return
	}
	config.insertTeams(teamList)
	config.syncEvents()
	// Sync 2019-2020 events.
	for year := 2019; year <= 2020; year++ {
		matchChan, _, numEvents := config.tba.GetAllEventMatches(year)
		for i := 0; i < numEvents; i++ {
			log.Printf("Upserting event #%d of %d", i, numEvents)
			matches := <-matchChan
			if len(matches) > 0 {
				log.Printf(matches[0].Key)
			}
			config.insertMatches(matches)
		}
	}
	config.predict()
}

// Sync predictors
func (config *Config) predict() {
	for key, val := range config.predictors {
		_, err := calculateModel(config, val, key)
		if err != nil {
			log.Println(err)
		}
	}
}
