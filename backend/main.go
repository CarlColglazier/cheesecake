package main

import (
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

const POOLS = 2

type Config struct {
	Conn *pgx.ConnPool
	Tba  *tba.TheBlueAlliance
}

func init() {
	log.SetFlags(0)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	log.Println("init started")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	conn, err := Connect("db", "cheesecake")
	errors := 0
	for errors < 5 {
		if err != nil {
			log.Println("Could not load database")
			time.Sleep(1000 * time.Millisecond)
		}
	}
	log.Println("Connected to the database.")
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := tba.NewTba(tbakey)
	defer tbaInst.Close()
	config := Config{Conn: conn, Tba: tbaInst}
	log.Println("Config created")
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "server" {
			runServer(config)
		} else if args[0] == "reset" {
			reset(&config)
		}
	}
}
