package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/carlcolglazier/cheesecake/tba"
)

func TestConnect(t *testing.T) {
	time.Sleep(500 * time.Millisecond)
	conn, err := Connect("testdb", "cheesecaketest")
	if err != nil {
		t.Errorf("Could not connect to database: %s", err)
	}
	tbaInst := tba.NewTba("key")
	defer tbaInst.Close()
	config := Config{conn: conn, tba: tbaInst}
	fmt.Printf("%+v\n", config)
	config.Migrate("testdb", "cheesecaketest")
	err = config.CacheSetStr("key", `{"value": 0}`)
	if err != nil {
		t.Errorf("Could not set key in cache, %s", err)
	}
	value, err := config.CacheGetStr("key")
	if err != nil {
		t.Error("Could not get 'key' in cache")
	}
	if value != `{"value": 0}` {
		t.Errorf("Incorrect value for 'key': %s", value)
	}
	// Do a second entry.
	err = config.CacheSetStr("key2", `{"value": 1}`)
	if err != nil {
		t.Errorf("Could not set key2 in cache, %s", err)
	}
	value, err = config.CacheGetStr("key2")
	if err != nil {
		t.Errorf("Could not get 'key2' in cache: %s", err)
	}
	if value != `{"value": 1}` {
		t.Errorf("Incorrect value for 'key2': %s", value)
	}
}
