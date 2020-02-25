package main

import (
	"log"
	"os"
	"time"

	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

// Main struct to contain application connections.
type Config struct {
	conn   *pgxpool.Pool
	tba    *tba.TheBlueAlliance
	models map[string]Model
}

func NewConfig(conn *pgxpool.Pool, tba *tba.TheBlueAlliance) *Config {
	predictors := make(map[string]Model)
	config := Config{conn: conn, tba: tba, models: predictors}
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
	config.models["eloscore2019"] = NewEloScoreModel(2019)
	config.models["eloscore2020"] = NewEloScoreModel(2020)
	config.models["rocket"] = NewBetaModel(0.5, 12.0, "completeRocketRankingPoint", 2019)
	config.models["hab"] = NewBetaModel(0.7229, 2.4517, "habDockingRankingPoint", 2019)
	config.models["shieldop"] = NewBetaModel(0.5, 12.0, "shieldOperationalRankingPoint", 2020)
	config.models["shieldeng"] = NewBetaModel(0.5, 12.0, "shieldEnergizedRankingPoint", 2020)
	// ---
	// Check: do we need to reset?
	dbVersion := config.version()
	if dbVersion == 0 {
		reset(config)
	}
	// Parse command line input
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "server" {
			runServer(config)
		} else if args[0] == "reset" {
			reset(config)
		} else if args[0] == "predict" {
			config.predict()
		} else if args[0] == "forecast" {
			config.forecast()
		} else if args[0] == "forecastfull" {
			config.forecast2019()
			config.forecast2020()
		}
	}
}
