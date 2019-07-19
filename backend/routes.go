package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func runServer(config Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/matches", config.MatchReq)
	router.HandleFunc("/reset", config.ResetReq)
	router.HandleFunc("/elo", config.CalcElo)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func (config *Config) ResetReq(w http.ResponseWriter, r *http.Request) {
	reset(config)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Done")
}

func (config *Config) MatchReq(w http.ResponseWriter, r *http.Request) {
	matches, err := config.getMatches()
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(matches)
}

func (config *Config) CalcElo(w http.ResponseWriter, r *http.Request) {
	matches, err := config.getMatches()
	if err != nil {
		log.Println(err)
	}
	pred := NewEloScriptPredictor()
	for _, match := range matches {
		pred.AddResult(match)
	}
	json.NewEncoder(w).Encode(pred.CurrentValues())
}
