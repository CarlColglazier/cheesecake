package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"os"
)

func Connect(applicationName string) (conn *pgx.ConnPool) {
	var runtimeParams map[string]string
	runtimeParams = make(map[string]string)
	runtimeParams["application_name"] = applicationName
	connConfig := pgx.ConnConfig{
		User:              "postgres",
		Password:          "postgres",
		Host:              "localhost",
		Port:              5432,
		Database:          "postgres",
		TLSConfig:         nil,
		UseFallbackTLS:    false,
		FallbackTLSConfig: nil,
		RuntimeParams:     runtimeParams,
	}
	connPoolConfig := pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 8}
	conn, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to establish connection: %v\n", err)
		os.Exit(1)
	}
	return conn
}
