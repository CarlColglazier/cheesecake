package main

import (
	"encoding/json"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"github.com/mediocregopher/radix"
	"io/ioutil"
	"log"
	"os"
)

const POOLS = 2

// The program cannot run if these variables are
// not declared.
//var required_env = []string{"REDIS", "TBA_KEY"}

type Config struct {
	Pool *radix.Pool
	Conn *pgx.ConnPool
	Tba  *TheBlueAlliance
}

// Not used right now?
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
	pool, err := radix.NewPool("tcp", os.Getenv("REDIS"), POOLS)
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
		if args[0] == "server" {
			runServer(config)
		} else if args[0] == "reset" {
			reset(&config)
		}
	}
}
