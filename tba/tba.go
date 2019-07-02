package tba

import "net/http"
import "fmt"
import "io/ioutil"
import "log"
import "errors"

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
}

/// Create a new TheBlueAlliance object and initialize the cache.
/// Note: the Key is super required because you can't access the API
/// without it.
func NewTba(key string) *TheBlueAlliance {
	var tba TheBlueAlliance
	tba.Key = key
	// Down the road, this will connect to the database.
	tba.cache = make(map[string]TBACall)
	return &tba
}

func (tba *TheBlueAlliance) tbaRequest(url string) (string, error) {
	lastTime := "Sun, 30 Jun 2000 09:07:40 GMT"
	if val, ok := tba.cache[url]; ok {
		log.Println(val)
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
		body := tba.cache[url].Body
		return body, nil
	} else if resp.StatusCode == 200 {
		respTime := resp.Header.Get("Last-Modified")
		body, _ := ioutil.ReadAll(resp.Body)
		cacheEntry := TBACall{Modified: respTime, Body: string(body)}
		tba.cache[url] = cacheEntry
		return string(body), nil
	} else {
		return "", errors.New("Page not found")
	}
}

func (tba *TheBlueAlliance) Team(s string) (string, error) {
	url := fmt.Sprintf("team/%s/simple", s)
	return tba.tbaRequest(url)
}

func (tba *TheBlueAlliance) GetAllTeams() (string, error) {
	s := ""
	for i := 0; i < 20; i++ {
		url := fmt.Sprintf("teams/%d", i)
		r, err := tba.tbaRequest(url)
		if err != nil {
			return s, err
		}
		s += r
	}
	return s, nil
}

func (tba *TheBlueAlliance) Close() {
	// Closes and saves
}
