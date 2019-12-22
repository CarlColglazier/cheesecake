package main

import (
	"fmt"
	"github.com/carlcolglazier/cheesecake/tba"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	time.Sleep(2500 * time.Millisecond)
	conn, err := Connect("testdb", "cheesecaketest")
	if err != nil {
		t.Errorf("Could not connect to database: %s", err)
	}
	tbaInst := tba.NewTba("key")
	defer tbaInst.Close()
	config := Config{Conn: conn, Tba: tbaInst}
	fmt.Printf("%+v\n", config)
}
