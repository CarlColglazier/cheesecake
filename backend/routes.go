package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func runServer(config Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/matches", config.MatchReq)
	router.HandleFunc("/matches/{event}", config.GetEventMatchesReq)
	router.HandleFunc("/reset", config.ResetReq)
	router.HandleFunc("/events", config.EventReq)
	router.HandleFunc("/elo", config.CalcElo)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	handler := handlers.CORS(corsObj)(router)
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
	j, err := calculateElo(config)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
