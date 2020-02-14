package main

import (
	"encoding/json"
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

const POOLS = 2

type Config struct {
	Conn       *pgx.ConnPool
	Tba        *tba.TheBlueAlliance
	predictors map[string]Predictor
}

func NewConfig(conn *pgx.ConnPool, tba *tba.TheBlueAlliance) *Config {
	predictors := make(map[string]Predictor)
	config := Config{Conn: conn, Tba: tba, predictors: predictors}
	return &config
}

func main() {
	log.SetFlags(log.Ldate | log.Llongfile)
	log.Println("Starting...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	conn, err := Connect("db", "cheesecake")
	error_count := 0
	for error_count < 5 {
		if err != nil {
			log.Println("Could not load database")
			time.Sleep(1000 * time.Millisecond)
			conn, err = Connect("db", "cheesecake")
			error_count += 1
		} else {
			log.Println("Connected to the database.")
			break
		}
	}
	if error_count == 5 {
		log.Fatal("Could not connect to database. Exiting.")
	}
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := tba.NewTba(tbakey)
	defer tbaInst.Close()
	config := NewConfig(conn, tbaInst)
	log.Println("Config created")
	log.Println("Adding predictors")
	// EloScore predictor
	scores, err := config.CacheGet("pred_eloscores")
	if err != nil {
		log.Println("Could not get elo scores from cache.")
		scores = make(map[string]interface{})
	}
	eloPred := NewEloScorePredictorFromCache(scores)
	config.predictors["eloscore"] = eloPred
	config.predictors["rocket"] = NewBetaPredictor(0.5, 12.0, "completeRocketRankingPoint", 2019)
	config.predictors["hab"] = NewBetaPredictor(0.7229, 2.4517, "habDockingRankingPoint", 2019)
	// ---
	// Check: do we need to reset?
	dbVersion := config.version()
	if dbVersion == 0 {
		reset(config)
		vals := config.predictors["eloscore"].CurrentValues()
		b, err := json.Marshal(vals)
		if err != nil {
			config.CacheSetStr("pred_eloscores", string(b))
		}
	}
	// Parse command line input
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "server" {
			runServer(config)
		} else if args[0] == "reset" {
			reset(config)
		}
	}
}
