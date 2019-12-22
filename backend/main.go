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

func main() {
	//log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
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
