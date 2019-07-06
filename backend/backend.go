package main

import (
	"encoding/json"
	"fmt"
	// "github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"github.com/mediocregopher/radix"
	"io/ioutil"
	"log"
	// "net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const POOLS = 4

type Config struct {
	Pool *radix.Pool
	Conn *pgx.ConnPool
	Tba  *TheBlueAlliance
}

/*
func runServer(config Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/team/{key}", TeamReq)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func TeamReq(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["key"])
}
*/
func ReadEloRecords() map[string]float64 {
	file, err := ioutil.ReadFile("elo2018.json")
	if err != nil {
		log.Fatal("Could not read elo file ", err)
	}
	var records map[string]float64
	_ = json.Unmarshal([]byte(file), &records)
	return records
}

func reset(config Config) {
	path := filepath.Join("sql", "create.sql")

	c, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		// handle error.
		log.Fatal(ioErr)
	}
	sql := string(c)
	_, err := config.Conn.Exec(sql)
	if err != nil {
		// handle error.
		log.Fatal(err)
	}
	teamList, err := config.Tba.GetAllTeams()
	if err != nil {
		log.Fatal(err)
	}
	var teams [][]interface{}
	for _, row := range teamList {
		teams = append(teams, []interface{}{
			row.Key,
			row.TeamNumber,
			row.Name,
		})
	}
	copyCount, _ := config.Conn.CopyFrom(
		pgx.Identifier{"team"},
		[]string{"key", "team_number", "name"},
		pgx.CopyFromRows(teams),
	)
	fmt.Println(copyCount)
	time.Sleep(2 * time.Second)
	var t int
	err = config.Conn.QueryRow("SELECT COUNT('key') FROM team").Scan(&t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)
	var wg sync.WaitGroup
	for i := 2003; i <= 2019; i++ {
		fmt.Println(i)
		go func(year int) {
			defer wg.Done()
			rows, _ := config.Tba.GetAllEvents(year)

			var r [][]interface{}
			for _, row := range rows {
				r = append(r, []interface{}{row.EndDate, row.Key, row.Year})
			}

			copyCount, _ := config.Conn.CopyFrom(
				pgx.Identifier{"event"},
				[]string{"end_date", "key", "year"},
				pgx.CopyFromRows(r),
			)
			fmt.Println(copyCount)
		}(i)

		wg.Add(1)
	}
	wg.Wait()
	rows, _ := config.Tba.GetAllEventMatches(2019)
	var r [][]interface{}
	var a [][]interface{}
	var aTeams [][]interface{}
	for _, row := range rows {
		r = append(r, []interface{}{
			row.Key,
			row.CompLevel,
			row.SetNumber,
			row.MatchNumber,
			row.WinningAlliance,
			row.EventKey,
		})
		a = append(a, []interface{}{
			row.Key + "_blue",
			row.Alliances.Blue.Score,
			"blue",
			row.Key,
		})
		for _, team := range row.Alliances.Blue.TeamKeys {
			aTeams = append(aTeams,
				[]interface{}{
					row.Key + "_blue",
					team,
				})
		}
		a = append(a, []interface{}{
			row.Key + "_red",
			row.Alliances.Red.Score,
			"red",
			row.Key,
		})
	}
	copyCount, _ = config.Conn.CopyFrom(
		pgx.Identifier{"match"},
		[]string{
			"key", "comp_level", "set_number", "match_number", "winning_alliance", "event_key",
		},
		pgx.CopyFromRows(r),
	)
	fmt.Println(copyCount)
	copyCount, _ = config.Conn.CopyFrom(
		pgx.Identifier{"alliance"},
		[]string{
			"key", "score", "color", "match_key",
		},
		pgx.CopyFromRows(a),
	)
	fmt.Println(copyCount)
	copyCount, err = config.Conn.CopyFrom(
		pgx.Identifier{"alliance_teams"},
		[]string{
			"alliance_id", "team_key",
		},
		pgx.CopyFromRows(aTeams),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(copyCount)

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	pool, err := radix.NewPool("tcp", "127.0.0.1:6379", POOLS)
	conn := Connect("cheesecake")
	defer pool.Close()
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := NewTba(tbakey, pool)
	defer tbaInst.Close()
	if err != nil {
		log.Fatal(err)
	}
	config := Config{Pool: pool, Conn: conn, Tba: tbaInst}
	args := os.Args[1:]
	if len(args) == 1 {
		/*if args[0] == "server" {
			runServer(config)
		} else*/
		if args[0] == "reset" {
			reset(config)
		}
	}
}
