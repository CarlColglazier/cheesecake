package main

import (
	"context"
	"fmt"

	//"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect(host, applicationName string) (*pgxpool.Pool, error) {
	connString := "postgres://cheese:cheesepass4279@" + host + ":5432/" + applicationName + "?pool_max_conns=16&pool_min_conns=1"
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	fmt.Println(connString)
	//config.MaxConnLifetime = time.Duration(2 * 1000000000)
	return pgxpool.ConnectConfig(context.Background(), config)
}
