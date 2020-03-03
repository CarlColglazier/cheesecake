package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect(host, applicationName string) (*pgxpool.Pool, error) {
	connString := "postgres://cheese:cheesepass4279@" + host + ":5432/" + applicationName
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	fmt.Println(connString)
	config.MinConns = 1
	config.MaxConns = 5
	config.MaxConnLifetime = 10 * time.Second
	config.HealthCheckPeriod = 15 * time.Second
	config.ConnConfig.LogLevel = pgx.LogLevelDebug
	return pgxpool.ConnectConfig(context.Background(), config)
}
