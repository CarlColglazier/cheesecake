package tba

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
}

// NewTba creates a new TheBlueAlliance object and initialize the
// cache.  Note: the Key is super required because you can't access
// the API without it.
func NewTba(key string) *TheBlueAlliance {
	var tba TheBlueAlliance
	tba.Key = key
	// Down the road, this will connect to the database.
	tba.cache = make(map[string]TBACall)
	//tba.pool = pool
	return &tba
}

// tbaRequest is an internal function that fetches data from The Blue
// Alliance. It integrates with the cache to increase speed and reduce
// network load.
func (tba *TheBlueAlliance) tbaRequest(url string) (string, error) {
	fmt.Println(url)
	lastTime := "Sun, 30 Jun 2000 09:07:40 GMT"
	val := TBACall{Modified: "", Body: ""}
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
		body, _ := ioutil.ReadAll(resp.Body)
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

func (tba *TheBlueAlliance) GetAllEventMatches(year int) (chan []Match, error, int) {
	events, err := tba.GetAllOfficialEvents(year)
	if err != nil {
		log.Println(err)
		return nil, err, 0
	}
	channel := make(chan []Match)
	for _, event := range events {
		url := fmt.Sprintf("event/%s/matches", event.Key)
		go func(url string) {
			matchesString, _ := tba.tbaRequest(url)
			var matchList []Match
			_ = json.Unmarshal([]byte(matchesString), &matchList)
			channel <- matchList
		}(url)
	}
	return channel, nil, len(events)
}

func (tba *TheBlueAlliance) GetEventMatches(key string) ([]Match, error) {
	url := fmt.Sprintf("event/%s/matches", key)
	matchesString, _ := tba.tbaRequest(url)
	var matchList []Match
	err = json.Unmarshal([]byte(matchesString), &matchList)
	if err != nil {
		return nil, err
	}
	return matchList
}

func (tba *TheBlueAlliance) GetAllEvents(year int) ([]Event, error) {
	events, err := tba.tbaRequest(fmt.Sprintf("events/%d", year))
	if err != nil {
		return nil, err
	}
	var e []Event
	err = json.Unmarshal([]byte(events), &e)
	return e, err
}

func (tba *TheBlueAlliance) GetAllOfficialEvents(year int) ([]Event, error) {
	e, err := tba.GetAllEvents(year)
	if err != nil {
		return nil, err
	}
	var eventSlice []Event
	for i := range e {
		if e[i].EventType != -1 && e[i].EventType != 99 {
			eventSlice = append(eventSlice, e[i])
		}
	}
	return eventSlice, err
}

func (tba *TheBlueAlliance) Close() {
	// Closes and saves
}
