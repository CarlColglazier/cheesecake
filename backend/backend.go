package main

import (
	"fmt"
	//"github.com/carlcolglazier/cheesecake/tba"
	"github.com/joho/godotenv"
	"github.com/mediocregopher/radix"
	"log"
	"os"
)

const POOLS = 2

func syncTba() {
	tbakey := os.Getenv("TBA_KEY")
	pool, err := radix.NewPool("tcp", "127.0.0.1:6379", POOLS)
	defer pool.Close()
	if err != nil {
		log.Fatal(err)
	}
	tbaInst := NewTba(tbakey, pool)
	defer tbaInst.Close()
	for i := 0; i < 10; i++ {
		go tbaInst.Team(fmt.Sprintf("frc254%d", i))
	}
	for i := 0; i < 10; i++ {
		go tbaInst.Team(fmt.Sprintf("frc254%d", i))
	}
	err = tbaInst.GetAllTeams()
	if err != nil {
		log.Fatal("get teams", err)
	}
	_ = tbaInst.GetAllEventMatches(2019)

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "sync" {
			syncTba()
		}
	}
}
