package main

import (
	"github.com/carlcolglazier/cheesecake/tba"
	"testing"
)

func loadTestDB() Config {
	conn, _ := Connect("testdb", "cheesecaketest")
	tbaInst := tba.NewTba("key")
	defer tbaInst.Close()
	config := Config{conn: conn, tba: tbaInst}
	return config
}

func TestInsertTeams(t *testing.T) {
	config := loadTestDB()
	//config.Migrate("testdb", "cheesecaketest")
	teams := []tba.Team{
		{Key: "frc1", TeamNumber: 1, Name: "One"},
		{Key: "frc2", TeamNumber: 2, Name: "Two"},
	}
	config.insertTeams(teams)
	rows, err := config.conn.Query(`SELECT key, team_number FROM team`)
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
	// Now try upserting.
	teams = []tba.Team{
		{Key: "frc1", TeamNumber: 1, Name: "One 1"},
		{Key: "frc2", TeamNumber: 2, Name: "Two"},
		{Key: "frc3", TeamNumber: 3, Name: "Three"},
	}
	config.insertTeams(teams)
	rows, err = config.conn.Query(`SELECT key, team_number, name FROM team`)
	defer rows.Close()
	if err != nil {
		t.Errorf("%s", err)
	}
	var name string
	b = rows.Next()
	if b != true {
		t.Errorf("No row returned for first line")
	}
	rows.Scan(&key, &num, &name)
	if key != "frc1" {
		t.Errorf("Expected %v got %v", "frc1", key)
	}
	if num != 1 {
		t.Errorf("Expected %v got %v", 1, num)
	}
	if name != "One 1" {
		t.Errorf("Expected %v got %v", "One 1", name)
	}
	b = rows.Next()
	if b != true {
		t.Errorf("No row returned for second line")
	}
	rows.Scan(&key, &num, &name)
	if key != "frc2" {
		t.Errorf("Expected %v got %v", "frc2", key)
	}
	if num != 2 {
		t.Errorf("Expected %v got %v", 2, num)
	}
	if name != "Two" {
		t.Errorf("Expected %v got %v", "Two", name)
	}
	b = rows.Next()
	if b != true {
		t.Errorf("No row returned for third line")
	}
	rows.Scan(&key, &num, &name)
	if key != "frc3" {
		t.Errorf("Expected %v got %v", "frc3", key)
	}
	if num != 3 {
		t.Errorf("Expected %v got %v", 3, num)
	}
	if name != "Three" {
		t.Errorf("Expected %v got %v", "Three", name)
	}
}

func TestInsertEvents(t *testing.T) {
	config := loadTestDB()
	events := []tba.Event{
		{Key: "2019abcd", ShortName: "A Big C Deal", Year: 2019},
		{Key: "2019ef", ShortName: "Everyman Fortune", Year: 2019},
	}
	config.insertEvents(events)
	rows, err := config.conn.Query(`SELECT key FROM event`)
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
	confevents, err := config.getEvents()
	if err != nil {
		t.Error(err)
	}
	if len(confevents) != 2 {
		t.Errorf("Expected %v got %v", 2, len(confevents))
	}
	// TODO: Check for equality.
}
