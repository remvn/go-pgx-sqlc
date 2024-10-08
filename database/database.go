package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remvn/go-pgx-sqlc/database/sqlc"
)

type Database struct {
	Pool  *pgxpool.Pool
	Query *sqlc.Queries
}

func NewDatabase(connStr string) *Database {
	// this only create pgxpool struct, you may need to ping the database to
	// initialize a connection and check availability
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}

	// this is generated by sqlc cli
	query := sqlc.New(pool)

	database := Database{
		Pool:  pool,
		Query: query,
	}
	return &database
}
