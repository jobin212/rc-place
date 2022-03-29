package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func setupPostgresConnection() error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO tile_info(username, x, y, color) VALUES ($1, $2, $3, $4)", "user123", 5, 9, "purple")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return err
	}

	return nil
}
