package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func runServer(config *Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/matches/{event}", config.GetEventMatchesReq)
	router.HandleFunc("/events", config.EventReq)
	router.HandleFunc("/events/{year}", config.EventYearReq)
	router.HandleFunc("/forecasts/{event}", config.getEventForecastsReq)
	//router.HandleFunc("/elo", config.CalcEloScores)
	router.HandleFunc("/brier", config.Brier)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	handler := handlers.CORS(corsObj)(router)
	http.ListenAndServe(":8080", handler)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{}`)
}

func (config *Config) GetEventMatchesReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	year := vars["event"][0:4]
	var matches []MatchEntry
	var err error
	if year == "2020" {
		matches, err = config.getEventMatches2020(vars["event"])
	} else if year == "2019" {
		matches, err = config.getEventMatches2019(vars["event"])
	} else {
		matches, err = config.getEventMatches2019(vars["event"])
	}
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(matches)
}

func (config *Config) getEventForecastsReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	fore, err := config.getEventForecasts(vars["event"])
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(fore)
}

func (config *Config) EventReq(w http.ResponseWriter, r *http.Request) {
	events, err := config.getEvents(2019)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(events)
}

func (config *Config) EventYearReq(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode("{}")
	}
	events, err := config.getEvents(year)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(events)
}

/*
func (config *Config) CalcEloScores(w http.ResponseWriter, r *http.Request) {
	pred := NewEloScoreModel(2019)
	j, err := calculateModel(config, pred, "eloscores")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
*/

func (config *Config) Brier(w http.ResponseWriter, r *http.Request) {
	rows, err := config.conn.Query(
		context.Background(),
		`select 
  avg(power((winning_alliance='red')::int - (prediction->'red')::text::float, 2)) as brier,
	count(*) filter 
		(where (winning_alliance='red' and (prediction->'red')::text::float > 0.5) or (winning_alliance='blue' and (prediction->'red')::text::float < 0.5)) as correct,
	count(*) as count
from match
inner join prediction_history on prediction_history."match"=match."key"
where match.winning_alliance is not null and length(match.winning_alliance) > 0
and model='eloscores' and match.event_key like '2019%'`,
	)
	if err != nil {
		log.Println(err)
	}
	var brier float32
	var correct int
	var count int
	rows.Next()
	rows.Scan(&brier, &correct, &count)
	fmt.Fprintf(w, `{"score": %v, "correct": %d, "count": %d}`, brier, correct, count)
}
