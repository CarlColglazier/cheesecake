package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Connect(host, applicationName string) (*pgxpool.Pool, error) {
	connString := "postgres://cheese:cheesepass4279@" + host + ":5432/" + applicationName + "?"
	fmt.Println(connString)
	return pgxpool.Connect(context.Background(), connString)
}
