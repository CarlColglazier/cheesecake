package main

import (
	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
	"github.com/mediocregopher/radix/v3"
	"log"
	"os"
	"tba"
)

const POOLS = 2

// The program cannot run if these variables are
// not declared.
//var required_env = []string{"REDIS", "TBA_KEY"}

type Config struct {
	Pool *radix.Pool
	Conn *pgx.ConnPool
	Tba  *tba.TheBlueAlliance
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
	tbaInst := tba.NewTba(tbakey, pool)
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
