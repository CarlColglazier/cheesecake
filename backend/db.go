package main

import (
	"github.com/jackc/pgx"
)

func Connect(host, applicationName string) (*pgx.ConnPool, error) {
	var runtimeParams map[string]string
	runtimeParams = make(map[string]string)
	runtimeParams["application_name"] = applicationName
	connConfig := pgx.ConnConfig{
		User:              "cheese",
		Password:          "cheesepass4279",
		Host:              host,
		Port:              5432,
		Database:          applicationName,
		TLSConfig:         nil,
		UseFallbackTLS:    false,
		FallbackTLSConfig: nil,
		RuntimeParams:     runtimeParams,
	}
	connPoolConfig := pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 8}
	return pgx.NewConnPool(connPoolConfig)
}
