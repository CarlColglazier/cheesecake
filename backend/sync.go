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
	batch := &pgx.Batch{}
	for _, row := range teamList {
		batch.Queue(
			"insert into team (key, team_number, name) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT team_pkey DO UPDATE set team_number = $2, name = $3",
			row.Key, row.TeamNumber, row.Name,
		)
	}
	res := config.conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(teamList); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err := res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
}

// Upserts a list of events into the database.
func (config *Config) insertEvents(eventList []tba.Event) {
	batch := &pgx.Batch{}
	for _, row := range eventList {
		batch.Queue(
			"INSERT INTO event (key, end_date, event_type, short_name, year) values ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT event_pkey DO UPDATE set end_date = $2, event_type = $3, short_name = $4, year = $5",
			row.Key, row.EndDate, row.EventType, row.ShortName, row.Year,
		)
	}
	res := config.conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(eventList); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err := res.Close()
	if err != nil {
		log.Println("Error closing batch.")
		log.Println(err)
	}
}

func (config *Config) version() int {
	row, err := config.conn.Query(context.Background(), "select \"version\" from schema_migrations")
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
	batch := &pgx.Batch{}
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Println(err)
	}
	for _, row := range matchList {
		batch.Queue(
			"insert into match (key, comp_level, set_number, match_number, winning_alliance, event_key, time, actual_time, predicted_time, post_result_time, score_breakdown) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT ON CONSTRAINT match_pkey DO NOTHING",
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
		for _, team := range row.Alliances.Red.TeamKeys {
			if reg.Match([]byte(team[3:])) {
				//log.Println(team)
				team = "frc9990"
				//log.Println(team)
			}
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key) values ($1, $2) ON CONFLICT DO NOTHING",
				allianceId,
				team,
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
		for _, team := range row.Alliances.Blue.TeamKeys {
			if reg.Match([]byte(team[3:])) {
				team = "frc9990"
			}
			batch.Queue(
				"insert into alliance_teams (alliance_id, team_key) values ($1, $2) ON CONFLICT DO NOTHING",
				allianceId,
				team,
			)
		}
	}
	res := config.conn.SendBatch(context.Background(), batch)
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
	batch := &pgx.Batch{}
	for _, row := range forecasts {
		batch.Queue(
			"INSERT INTO forecast_history (model, match_key, team_key, forecast) VALUES ($1, $2, $3, $4) ON CONFLICT CONSTRAINT forecast_pkey DO UPDATE set forecast = $4",
			row.Model, row.MatchKey, row.TeamKey, row.Forecast,
		)
	}
	res := config.conn.SendBatch(context.Background(), batch)
	for i := 0; i < len(forecasts); i++ {
		_, err := res.Exec()
		if err != nil {
			log.Printf("Error upserting: %s", err)
			return
		}
	}
	err := res.Close()
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

// Sync predictors
func (config *Config) predict() {
	batch := &pgx.Batch{}
	matches, _ := config.getMatches()
	qCount := 0
	simp_model := NewEloScoreModel(2019)
	simp_model2 := NewBetaModel(0.5, 12.0, "completeRocketRankingPoint", 2019)
	simp_model3 := NewBetaModel(0.7229, 2.4517, "habDockingRankingPoint", 2019)
	eloModel2020 := NewEloScoreModel(2020)
	for _, match := range matches {
		if match.Match.Key[0:4] == "2020" {
			// Run the forecast
			if match.Match.CompLevel == "qm" && (match.Match.MatchNumber)%5 == 1 {

			}
			eloModel2020.AddResult(match)
		}
		if match.Match.Key[0:4] == "2019" {
			// Run the forecast here
			if match.Match.CompLevel == "qm" && (match.Match.MatchNumber)%5 == 1 {
				forecastMatches := make([]MatchEntry, 0)
				//matchesCopy := make([]MatchEntry, len(matches))
				//copy(matchesCopy, matches)
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
				leadersCast, capsCast := config.forecastEvent(match.Match.Time, forecastMatches)
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
			}
			simp_model.AddResult(match)
			simp_model2.AddResult(match)
			simp_model2.AddResult(match)
		}
		for modelkey, model := range config.models {
			if !model.SupportsYear(match.year()) {
				continue
			}
			p := model.Predict(match)
			batch.Queue("insert into prediction_history (match, model, prediction) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT prediction_history_pkey DO UPDATE set prediction = $3",
				match.Match.Key, modelkey, p,
			)
			qCount += 1
			if match.Official {
				model.AddResult(match)
			}
		}
	}
	log.Printf("Sending %d predictions to database...", qCount)
	res := config.conn.SendBatch(context.Background(), batch)
	for i := 0; i < qCount; i++ {
		_, err := res.Exec()
		if err != nil {
			log.Println(err)
			//break
		}
	}
	err := res.Close()
	if err != nil {
		log.Println("Error closing batch.")
	}
	//config.forecast()
}
