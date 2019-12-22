package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"os"
	"time"
)

func Connect(applicationName string) (conn *pgx.ConnPool) {
	var runtimeParams map[string]string
	runtimeParams = make(map[string]string)
	runtimeParams["application_name"] = applicationName
	connConfig := pgx.ConnConfig{
		User:              "postgres",
		Password:          "postgres",
		Host:              "db",
		Port:              5432,
		Database:          "cheesecake",
		TLSConfig:         nil,
		UseFallbackTLS:    false,
		FallbackTLSConfig: nil,
		RuntimeParams:     runtimeParams,
	}
	connPoolConfig := pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 8}
	errors := 0
	for errors < 10 {
		conn, err := pgx.NewConnPool(connPoolConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to establish connection: %v\n", err)
			time.Sleep(2000 * time.Millisecond)
			errors += 1
		} else {
			return conn
		}

	}
	os.Exit(1)
	return nil
}
