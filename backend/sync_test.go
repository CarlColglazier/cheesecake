package main

import (
	"github.com/carlcolglazier/cheesecake/tba"
	"testing"
)

func TestInsertTeams(t *testing.T) {
	conn, _ := Connect("testdb", "cheesecaketest")
	tbaInst := tba.NewTba("key")
	defer tbaInst.Close()
	config := Config{Conn: conn, Tba: tbaInst}
	//config.Migrate("testdb", "cheesecaketest")
	teams := []tba.Team{
		{Key: "frc1", TeamNumber: 1, Name: "One"},
		{Key: "frc2", TeamNumber: 2, Name: "Two"},
	}
	config.insertTeams(teams)
	rows, err := config.Conn.Query(`SELECT key, team_number FROM team`)
	defer rows.Close()
	if err != nil {
		t.Errorf("%s", err)
	}
	var key string
	var num int
	b := rows.Next()
	if b != true {
		t.Errorf("No row returned for first line")
	}
	rows.Scan(&key, &num)
	if key != "frc1" {
		t.Errorf("Expected %v got %v", "frc1", key)
	}
	if num != 1 {
		t.Errorf("Expected %v got %v", 1, num)
	}
	b = rows.Next()
	if b != true {
		t.Errorf("No row returned for second line")
	}
	rows.Scan(&key, &num)
	if key != "frc2" {
		t.Errorf("Expected %v got %v", "frc2", key)
	}
	if num != 2 {
		t.Errorf("Expected %v got %v", 2, num)
	}
}

func TestInsertEvents(t *testing.T) {
	conn, _ := Connect("testdb", "cheesecaketest")
	tbaInst := tba.NewTba("key")
	defer tbaInst.Close()
	config := Config{Conn: conn, Tba: tbaInst}
	events := []tba.Event{
		{Key: "2019abcd", ShortName: "A Big C Deal", Year: 2019},
		{Key: "2019ef", ShortName: "Everyman Fortune", Year: 2019},
	}
	config.insertEvents(events)
	rows, err := config.Conn.Query(`SELECT key FROM event`)
	defer rows.Close()
	if err != nil {
		t.Errorf("%s", err)
	}
	var key string
	b := rows.Next()
	if b != true {
		t.Errorf("No row returned for first line")
	}
	rows.Scan(&key)
	if key != "2019abcd" {
		t.Errorf("Expected %v got %v", "2019abcd", key)
	}
	b = rows.Next()
	if b != true {
		t.Errorf("No row returned for second line")
	}
	rows.Scan(&key)
	if key != "2019ef" {
		t.Errorf("Expected %v got %v", "2019ef", key)
	}
}
