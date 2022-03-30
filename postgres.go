package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var postgresClient *sql.DB

func setupPostgresConnection() error {
	var err error
	postgresClient, err = sql.Open("pgx", os.Getenv("PG_DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}
	return postgresClient.Ping()
}
