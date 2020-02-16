package main

import (
	"log"
	"os"
	"time"

	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
)

// Number of database connections.
const POOLS = 2

// Main struct to contain application connections.
type Config struct {
	conn       *pgx.ConnPool
	tba        *tba.TheBlueAlliance
	predictors map[string]Model
}

func NewConfig(conn *pgx.ConnPool, tba *tba.TheBlueAlliance) *Config {
	predictors := make(map[string]Model)
	config := Config{conn: conn, tba: tba, predictors: predictors}
	return &config
}

func main() {
	log.SetFlags(log.Ldate | log.Llongfile)
	log.Println("Starting...")
	// Load environment details.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Connect to the database.
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
	/*
		scores, err := config.CacheGet("pred_eloscores")
		if err != nil {
			log.Println("Could not get elo scores from cache.")
			scores = make(map[string]interface{})
		}
	*/
	//eloPred := NewEloScoreModelFromCache(scores)
	config.predictors["eloscore2019"] = NewEloScoreModel(2019)
	config.predictors["eloscore2020"] = NewEloScoreModel(2020)
	config.predictors["rocket"] = NewBetaModel(0.5, 12.0, "completeRocketRankingPoint", 2019)
	config.predictors["hab"] = NewBetaModel(0.7229, 2.4517, "habDockingRankingPoint", 2019)
	config.predictors["shieldop"] = NewBetaModel(0.5, 12.0, "shieldOperationalRankingPoint", 2020)
	config.predictors["shieldeng"] = NewBetaModel(0.5, 12.0, "shieldEnergizedRankingPoint", 2020)
	// ---
	// Check: do we need to reset?
	dbVersion := config.version()
	if dbVersion == 0 {
		reset(config)
	}
	//config.predict()
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
