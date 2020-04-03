package main

import (
	"context"
	"log"
	"regexp"
	"sync"

	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx/v4"
)

// Upserts a list of teams into the database.
func (config *Config) insertTeams(teamList []tba.Team) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	for _, row := range teamList {
		batch.Queue(
			"insert into team (key, team_number, name) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT team_pkey DO UPDATE set team_number = $2, name = $3",
			row.Key, row.TeamNumber, row.Name,
		)
	}
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(teamList); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

// Upserts a list of events into the database.
func (config *Config) insertEvents(eventList []tba.Event) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	for _, row := range eventList {
		batch.Queue(
			"INSERT INTO event (key, end_date, event_type, short_name, year) values ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT event_pkey DO UPDATE set end_date = $2, event_type = $3, short_name = $4, year = $5",
			row.Key, row.EndDate, row.EventType, row.ShortName, row.Year,
		)
	}
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(eventList); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
		log.Println(err)
	}
}

func (config *Config) version() int {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return -1
	}
	defer conn.Release()
	row, err := conn.Query(context.Background(), "select \"version\" from schema_migrations")
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
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Println(err)
	}
	for _, row := range matchList {
		batch.Queue(
			"insert into match (key, comp_level, set_number, match_number, winning_alliance, event_key, time, actual_time, predicted_time, post_result_time, score_breakdown) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT ON CONSTRAINT match_pkey DO UPDATE set winning_alliance = $5, actual_time = $8, post_result_time = $10, score_breakdown = $11",
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
		)
		allianceId := row.Key + "_red"
		batch.Queue(
			"insert into alliance (key, score, color, match_key) values ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT alliance_pkey DO UPDATE set score = $2",
			allianceId,
			row.Alliances.Red.Score,
			"red",
			row.Key,
		)
		for pos, team := range row.Alliances.Red.TeamKeys {
			if reg.Match([]byte(team[3:])) {
				//log.Println(team)
				team = "frc9990"
				//log.Println(team)
			}
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key, position) values ($1, $2, $3) ON CONFLICT DO NOTHING",
				allianceId,
				team,
				pos,
			)
		}
		allianceId = row.Key + "_blue"
		batch.Queue(
			"insert into alliance (key, score, color, match_key) values ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT alliance_pkey DO UPDATE set score = $2",
			allianceId,
			row.Alliances.Blue.Score,
			"blue",
			row.Key,
		)
		for pos, team := range row.Alliances.Blue.TeamKeys {
			if reg.Match([]byte(team[3:])) {
				team = "frc9990"
			}
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key, position) values ($1, $2, $3) ON CONFLICT DO NOTHING",
				allianceId,
				team,
				pos,
			)
		}
	}
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(matchList)*9; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

func (config *Config) insertForecasts(forecasts []ForecastHistory) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	for _, row := range forecasts {
		batch.Queue(
			"INSERT INTO forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
			row.Model, row.MatchKey, row.TeamKey, row.Forecast,
		)
	}
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(forecasts); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err = res.Close()
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

func (config *Config) predictMatch(match MatchEntry) {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	qCount := 0
	for modelkey, model := range config.models {
		if !model.SupportsYear(match.year()) {
			continue
		}
		p := model.Predict(match)
		// This does not do anything if the prediction
		// already exists.
		batch.Queue("insert into prediction_history (match, model, prediction) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT prediction_history_pkey DO UPDATE set prediction = $3",
			match.Match.Key, modelkey, p,
		)
		qCount += 1
	}
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < qCount; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Println(err)
			//break
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

// Sync predictors
func (config *Config) predict() {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	matches, err := config.getMatches()
	if err != nil {
		log.Println(err)
		return
	}
	qCount := 0
	for _, match := range matches {
		for modelkey, model := range config.models {
			if !model.SupportsYear(match.year()) {
				continue
			}
			p := model.Predict(match)
			batch.Queue("insert into prediction_history (match, model, prediction) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT prediction_history_pkey DO UPDATE set prediction = $3",
				match.Match.Key, modelkey, p,
			)
			qCount += 1
			if match.Official && match.played() {
				model.AddResult(match)
			}
		}
	}
	log.Printf("Sending %d predictions to database...", qCount)
	log.Printf("This is out of %d matches.", len(matches))
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < qCount; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Println(err)
			//break
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

func (config *Config) forecast2019() {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	matches, _ := config.getMatches()
	qCount := 0
	simp_model := NewEloScoreModel(2019, 21.1)
	simp_model2 := NewBetaModel(0.5, 12.0, "completeRocketRankingPoint", 2019)
	simp_model3 := NewBetaModel(0.7229, 2.4517, "habDockingRankingPoint", 2019)
	for _, match := range matches {
		if match.Match.Key[0:4] != "2019" {
			continue
		}
		if match.Match.CompLevel == "qm" && (match.Match.MatchNumber)%5 == 1 {
			forecastMatches := make([]MatchEntry, 0)
			for _, fmatch := range matches {
				// Only look at matches in the same event
				if fmatch.Match.EventKey != match.Match.EventKey {
					continue
				}
				p := simp_model.Predict(fmatch)
				p2 := simp_model2.Predict(fmatch)
				p3 := simp_model3.Predict(fmatch)
				fmatch.Predictions["elo_score"] = &PredictionHistory{Prediction: p}
				fmatch.Predictions["rocket"] = &PredictionHistory{Prediction: p2}
				fmatch.Predictions["hab"] = &PredictionHistory{Prediction: p3}
				forecastMatches = append(forecastMatches, fmatch)
			}
			ret1, ret2 := config.forecastEvent(match.Match.Time, forecastMatches)
			leadersCast := ret1[0]
			capsCast := ret1[1]
			avgRp := ret2[0]
			stdRp := ret2[1]
			for team, times := range leadersCast {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"rpleader", match.Match.Key, team, float64(times)/100.0,
				)
				qCount += 1
			}
			for team, times := range capsCast {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"cap", match.Match.Key, team, float64(times)/100.0,
				)
				qCount += 1
			}
			for team, rp := range avgRp {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"meanrp", match.Match.Key, team, rp,
				)
				qCount += 1
			}
			for team, rp := range stdRp {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"stdrp", match.Match.Key, team, rp,
				)
				qCount += 1
			}
		}
		if match.Official {
			simp_model.AddResult(match)
			simp_model2.AddResult(match)
			simp_model2.AddResult(match)
		}
	}
	log.Printf("Sending %d predictions to database...", qCount)
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < qCount; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Println(err)
			//break
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

func lastPlayed(event string, matches []MatchEntry) int {
	last := 0
	for _, match := range matches {
		if match.Match.EventKey == event && match.Match.CompLevel == "qm" {
			if match.played() && match.Match.MatchNumber > last {
				last = match.Match.MatchNumber
			}
		}
	}
	return last
}

func (config *Config) forecast2020() {
	conn, err := config.conn.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	matches, _ := config.getMatches()
	qCount := 0
	eloModel2020 := NewEloScoreModel(2020, 40.0)
	shieldModel := NewBetaModel(0.5, 12.0, "shieldOperationalRankingPoint", 2020)
	energizedModel := NewBetaModel(0.5, 12.0, "shieldEnergizedRankingPoint", 2020)
	for _, match := range matches {
		if match.Match.Key[0:4] != "2020" {
			continue
		}
		if match.Match.MatchNumber != 1 {
			if match.Match.MatchNumber > lastPlayed(match.Match.EventKey, matches) {
				continue
			}
		}
		// Run the forecast
		if match.Match.CompLevel == "qm" {
			forecastMatches := make([]MatchEntry, 0)
			for _, fmatch := range matches {
				// Only look at matches in the same event
				if fmatch.Match.EventKey != match.Match.EventKey {
					continue
				}
				p := eloModel2020.Predict(fmatch)
				p2 := shieldModel.Predict(fmatch)
				p3 := energizedModel.Predict(fmatch)
				fmatch.Predictions["elo_score"] = &PredictionHistory{Prediction: p}
				fmatch.Predictions["shield"] = &PredictionHistory{Prediction: p2}
				fmatch.Predictions["energized"] = &PredictionHistory{Prediction: p3}
				forecastMatches = append(forecastMatches, fmatch)
			}
			ret1, ret2 := config.forecastEvent(match.Match.Time, forecastMatches)
			leadersCast := ret1[0]
			capsCast := ret1[1]
			avgRp := ret2[0]
			stdRp := ret2[1]
			for team, times := range leadersCast {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"rpleader", match.Match.Key, team, float64(times)/100.0,
				)
				qCount += 1
			}
			for team, times := range capsCast {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"cap", match.Match.Key, team, float64(times)/100.0,
				)
				qCount += 1
			}
			for team, rp := range avgRp {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"meanrp", match.Match.Key, team, rp,
				)
				qCount += 1
			}
			for team, rp := range stdRp {
				batch.Queue("insert into forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
					"stdrp", match.Match.Key, team, rp,
				)
				qCount += 1
			}
		}
		if match.Official {
			eloModel2020.AddResult(match)
		}
	}
	log.Printf("Sending %d predictions to database...", qCount)
	res := conn.SendBatch(context.Background(), batch)
	for i := 0; i < qCount; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Println(err)
			//break
		}
	}
	err = res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

func (config *Config) forecast() {
	config.forecast2020()
}
