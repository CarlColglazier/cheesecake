package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// URL for The Blue Alliance API. We are on v3 at the moment.
const BASE = "https://thebluealliance.com/api/v3/"

/// An API call item for the cache.
type TBACall struct {
	Modified string
	Body     string
}

/// Strores the access key; caches responses.
type TheBlueAlliance struct {
	Key   string
	cache map[string]TBACall
	pool  *radix.Pool
}

/// Create a new TheBlueAlliance object and initialize the cache.
/// Note: the Key is super required because you can't access the API
/// without it.
func NewTba(key string, pool *radix.Pool) *TheBlueAlliance {
	var tba TheBlueAlliance
	tba.Key = key
	// Down the road, this will connect to the database.
	tba.cache = make(map[string]TBACall)
	tba.pool = pool
	return &tba
}

func (tba *TheBlueAlliance) tbaRequest(url string) (string, error) {
	fmt.Println(url)
	lastTime := "Sun, 30 Jun 2000 09:07:40 GMT"
	//var val TBACall = TBACall{Modified: "", Body: ""}
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
		body := val.Body
		return body, nil
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

func (tba *TheBlueAlliance) GetAllTeams() error {
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		url := fmt.Sprintf("teams/%d", i)

		go func() {
			defer wg.Done()
			tba.tbaRequest(url)
		}()

		wg.Add(1)
	}
	wg.Wait()
	return nil
}

func (tba *TheBlueAlliance) GetAllEventMatches(year int) (err error) {
	events, err := tba.tbaRequest(fmt.Sprintf("events/%d/keys", year))
	if err != nil {
		return
	}

	var eventStrings []string
	err = json.Unmarshal([]byte(events), &eventStrings)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	for _, key := range eventStrings {
		url := fmt.Sprintf("event/%s/matches", key)
		go func() {
			defer wg.Done()
			tba.tbaRequest(url)
		}()
		wg.Add(1)
	}
	wg.Wait()
	return nil
}

func (tba *TheBlueAlliance) Close() {
	// Closes and saves
}
