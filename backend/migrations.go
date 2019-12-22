package main

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

/// Migrate drops the entire database and rebuilds it.
/// This is a temoporary function. This should instead connect to an admin
/// panel in the future.
func (config *Config) Migrate() {
	m, err := migrate.New(
		"file://migrations",
		"postgres://postgres:postgres@db:5432/cheesecake?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Down(); err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}