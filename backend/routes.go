package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func runServer(config *Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index).Methods("GET")
	router.HandleFunc("/", config.Webhook).Methods("POST")
	router.HandleFunc("/matches/{event}", config.GetEventMatchesReq)
	router.HandleFunc("/events", config.EventReq)
	router.HandleFunc("/events/{year}", config.EventYearReq)
	router.HandleFunc("/forecasts/{event}", config.getEventForecastsReq)
	router.HandleFunc("/brier", config.Brier)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	handler := handlers.CORS(corsObj)(router)
	http.ListenAndServe(":8080", handler)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{}`)
}

type WebhookData struct {
	MessageData map[string]interface{} `json:"message_data"`
	MessageType string                 `json:"message_type"`
}

func (config *Config) Webhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data WebhookData
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err)
	}
	switch data.MessageType {
	case "schedule_updated":
		key := data.MessageData["event_key"]
		matches, _ := config.tba.GetEventMatches(key.(string))
		config.insertMatches(matches)
		//config.tba.
	case "match_score":
		var m tba.Match
		mData := data.MessageData["match"]
		jStr, _ := json.Marshal(mData)
		log.Println(string(jStr))
		json.Unmarshal([]byte(jStr), &m)
		log.Printf("%v", m)
		// This code exists because TBA uses team_keys for the API
		// and team for the Webhooks.
		mD := mData.(map[string]interface{})
		mD = mD["alliances"].(map[string]interface{})
		mDred := mD["red"].(map[string]interface{})
		mDblue := mD["blue"].(map[string]interface{})
		mDredTeams := mDred["teams"].([]interface{})
		mDblueTeams := mDblue["teams"].([]interface{})
		for _, team := range mDredTeams {
			m.Alliances.Red.TeamKeys = append(m.Alliances.Red.TeamKeys, team.(string))
		}
		for _, team := range mDblueTeams {
			m.Alliances.Blue.TeamKeys = append(m.Alliances.Blue.TeamKeys, team.(string))
		}
		log.Printf("%v", m.Alliances.Red.TeamKeys)
		matchList := []tba.Match{m}
		config.insertMatches(matchList)
		// TODO: This should be some kind of real-time version.
		config.predict()
	case "verification":
		log.Println("Verification")
		log.Printf("%v", data)
	case "ping":
		log.Println("Pinged")
		fallthrough
	default:
		log.Println("Default")
		log.Printf("%v", data)
	}
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
inner join event on match.event_key=event.key
where match.winning_alliance is not null and length(match.winning_alliance) > 0
and model='eloscore2020' and match.event_key like '2020%' and event.event_type < 7`,
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
