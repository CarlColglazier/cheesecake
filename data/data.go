package main

import (
	"fmt"
	"github.com/carlcolglazier/cheesecake/tba"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tbakey := os.Getenv("TBA_KEY")
	tbaInst := tba.NewTba(tbakey)
	for i := 0; i < 10; i++ {
		tbaInst.Team(fmt.Sprintf("frc254%d", i))
		fmt.Println(i)
	}
	tbaInst.Close()
}
