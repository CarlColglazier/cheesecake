package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mediocregopher/radix"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const POOLS = 2

type Config struct {
	Pool *radix.Pool
}

// syncTba implements the command line option to fetch all data from
// The Blue Alliance.
func syncTba(config Config) {
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := NewTba(tbakey, config.Pool)
	defer tbaInst.Close()
	for i := 0; i < 10; i++ {
		go tbaInst.Team(fmt.Sprintf("frc254%d", i))
	}
	for i := 0; i < 10; i++ {
		go tbaInst.Team(fmt.Sprintf("frc254%d", i))
	}
	err := tbaInst.GetAllTeams()
	if err != nil {
		log.Fatal("get teams", err)
	}
	_ = tbaInst.GetAllEventMatches(2019)
}

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

func ReadEloRecords() map[string]float64 {
	file, err := ioutil.ReadFile("elo2018.json")
	if err != nil {
		log.Fatal("Could not read elo file ", err)
	}
	var records map[string]float64
	_ = json.Unmarshal([]byte(file), &records)
	return records
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	pool, err := radix.NewPool("tcp", "127.0.0.1:6379", POOLS)
	defer pool.Close()
	if err != nil {
		log.Fatal(err)
	}
	config := Config{Pool: pool}
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "sync" {
			syncTba(config)
		} else if args[0] == "server" {
			runServer(config)
		}
	}
}
