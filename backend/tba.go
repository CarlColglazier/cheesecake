package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix"
	"io/ioutil"
	"log"
	"net/http"
	//"sync"
)

// BASE URL for The Blue Alliance API. We are on v3 at the moment.
const BASE = "https://thebluealliance.com/api/v3/"

/// TBACall is n API call item for the cache.
type TBACall struct {
	Modified string
	Body     string
}

/// TheBlueAlliance strores the access key; caches responses.
type TheBlueAlliance struct {
	Key   string
	cache map[string]TBACall
	pool  *radix.Pool
}

// NewTba creates a new TheBlueAlliance object and initialize the
// cache.  Note: the Key is super required because you can't access
// the API without it.
func NewTba(key string, pool *radix.Pool) *TheBlueAlliance {
	var tba TheBlueAlliance
	tba.Key = key
	// Down the road, this will connect to the database.
	tba.cache = make(map[string]TBACall)
	tba.pool = pool
	return &tba
}

// tbaRequest is an internal function that fetches data from The Blue
// Alliance. It integrates with the cache to increase speed and reduce
// network load.
func (tba *TheBlueAlliance) tbaRequest(url string) (string, error) {
	fmt.Println(url)
	lastTime := "Sun, 30 Jun 2000 09:07:40 GMT"
	var s []byte
	val := TBACall{Modified: "", Body: ""}
	err := tba.pool.Do(radix.Cmd(&s, "GET", url))
	if err != nil {
		log.Println(":( ", err)
	}
	err = json.Unmarshal(s, &val)
	if err != nil {
		log.Println("Unmarshal ", err)
	} else if len(val.Modified) > 0 {
		lastTime = val.Modified
	}
	tbaurl := fmt.Sprintf("%s%s", BASE, url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", tbaurl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-TBA-Auth-Key", tba.Key)
	req.Header.Set("If-Modified-Since", lastTime)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 304 {
		return val.Body, nil
	} else if resp.StatusCode == 200 {
		respTime := resp.Header.Get("Last-Modified")
		body, _ := ioutil.ReadAll(resp.Body)
		cacheEntry := TBACall{Modified: respTime, Body: string(body)}
		marsh, err := json.Marshal(cacheEntry)
		if err != nil {
			log.Fatal("JSON Marshal", err)
		}
		err = tba.pool.Do(radix.Cmd(nil, "SET", url, string(marsh)))
		if err != nil {
			log.Fatal("pool set", err)
		}
		return string(body), nil
	}
	return "", errors.New("Page not found")
}

func (tba *TheBlueAlliance) Team(s string) (string, error) {
	url := fmt.Sprintf("team/%s/simple", s)
	return tba.tbaRequest(url)
}

func (tba *TheBlueAlliance) GetAllTeams() ([]Team, error) {
	pages := 20
	channel := make(chan []Team)
	for i := 0; i < pages; i++ {
		url := fmt.Sprintf("teams/%d", i)
		go func(url string) error {
			log.Println(url)
			teamString, err := tba.tbaRequest(url)
			if err != nil {
				log.Println(err)
				return err
			}
			var teamList []Team
			err = json.Unmarshal([]byte(teamString), &teamList)
			if err != nil {
				log.Println(err)
				return err
			}
			channel <- teamList
			return nil
		}(url)
	}
	var teamList []Team
	for i := 0; i < pages; i++ {
		teams := <-channel
		teamList = append(teamList, teams...)
	}
	return teamList, nil
}

func (tba *TheBlueAlliance) GetAllEventMatches(year int) ([]Match, error) {
	events, err := tba.GetAllEvents(year)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println("len events ", len(events))
	channel := make(chan []Match)
	for _, event := range events {
		fmt.Println(event.Key)
		url := fmt.Sprintf("event/%s/matches", event.Key)
		go func(url string) {
			matchesString, _ := tba.tbaRequest(url)
			var matchList []Match
			_ = json.Unmarshal([]byte(matchesString), &matchList)
			channel <- matchList
		}(url)
	}
	var matchList []Match
	for i := 0; i < len(events); i++ {
		matches := <-channel
		matchList = append(matchList, matches...)
	}
	return matchList, nil
}

func (tba *TheBlueAlliance) GetAllEvents(year int) ([]Event, error) {
	events, err := tba.tbaRequest(fmt.Sprintf("events/%d", year))
	if err != nil {
		return nil, err
	}
	var e []Event
	err = json.Unmarshal([]byte(events), &e)
	// Keep only official events.
	var eventSlice []Event
	for i := range e {
		if e[i].EventType >= 0 && e[i].EventType <= 6 {
			eventSlice = append(eventSlice, e[i])
		}
	}
	return eventSlice, err
}

func (tba *TheBlueAlliance) Close() {
	// Closes and saves
}
