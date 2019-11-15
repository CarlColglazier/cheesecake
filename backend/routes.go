package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/mediocregopher/radix/v3"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func runServer(config Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/matches", config.MatchReq)
	router.HandleFunc("/matches/{event}", config.GetEventMatchesReq)
	router.HandleFunc("/reset", config.ResetReq)
	router.HandleFunc("/train", config.TrainReq)
	router.HandleFunc("/events", config.EventReq)
	router.HandleFunc("/elo", config.CalcElo)
	handler := cors.Default().Handler(router)
	//log.Fatal(http.ListenAndServe(":8080", router), c.Handler(router))
	http.ListenAndServe(":8080", handler)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{status: "good 1"}`)
}

func (config *Config) ResetReq(w http.ResponseWriter, r *http.Request) {
	reset(config)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Done")
}

func (config *Config) TrainReq(w http.ResponseWriter, r *http.Request) {
	train(config)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Done")
}

func (config *Config) GetEventMatchesReq(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matches, err := config.getEventMatches(vars["event"])
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(matches)
}

func (config *Config) MatchReq(w http.ResponseWriter, r *http.Request) {
	matches, err := config.getMatches()
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(matches)
}

func (config *Config) EventReq(w http.ResponseWriter, r *http.Request) {
	events, err := config.getEvents()
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(events)
}

func (config *Config) CalcElo(w http.ResponseWriter, r *http.Request) {
	var s []byte
	err := config.Pool.Do(radix.Cmd(&s, "GET", "EloRating"))
	if err != nil || len(s) == 0 {
		log.Println("Could not fetch scores from cache. Calculating...")
		matches, err := config.getMatches()
		if err != nil {
			log.Println(err)
		}
		log.Println(len(matches))
		pred := NewEloScorePredictor()
		pred.Dampen()
		var predictions [][]interface{}
		for _, match := range matches {
			p := pred.Predict(match)
			//ph := PredictionHistory{match.Match.Key, p, "EloScore"}
			predictions = append(predictions, []interface{}{match.Match.Key, p, "EloScore"})
			pred.AddResult(match)
		}
		config.Conn.CopyFrom(
			pgx.Identifier{"prediction_history"},
			[]string{"match", "prediction", "model"},
			pgx.CopyFromRows(predictions),
		)
		json.NewEncoder(w).Encode(pred.CurrentValues())
		j, _ := json.Marshal(pred.CurrentValues())
		config.Pool.Do(radix.Cmd(nil, "SET", "EloRating", fmt.Sprintf("%s", j)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}
