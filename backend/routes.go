package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func runServer(config *Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/matches/{event}", config.GetEventMatchesReq)
	router.HandleFunc("/events", config.EventReq)
	router.HandleFunc("/events/{year}", config.EventYearReq)
	router.HandleFunc("/elo", config.CalcEloScores)
	router.HandleFunc("/marbles", config.CalcMarbles)
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
	matches, err := config.getEventMatches(vars["event"])
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(matches)
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

func (config *Config) CalcEloScores(w http.ResponseWriter, r *http.Request) {
	pred := NewEloScorePredictor()
	j, err := calculatePredictor(config, pred, "eloscores")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (config *Config) CalcMarbles(w http.ResponseWriter, r *http.Request) {
	pred := NewMarblePredictor()
	j, err := calculatePredictor(config, pred, "marbles")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (config *Config) Brier(w http.ResponseWriter, r *http.Request) {
	rows, err := config.Conn.Query(
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
