package main

import (
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const POOLS = 2

type Config struct {
	Conn *pgx.ConnPool
	Tba  *tba.TheBlueAlliance
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	conn := Connect("cheesecake")
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := tba.NewTba(tbakey)
	defer tbaInst.Close()
	if err != nil {
		log.Fatal(err)
	}
	config := Config{Conn: conn, Tba: tbaInst}
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "server" {
			runServer(config)
		} else if args[0] == "reset" {
			reset(&config)
		}
	}
}
