package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/remvn/go-pgx-sqlc/database"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func createContainer() (*postgres.PostgresContainer, error) {
	ctx := context.Background()
	dbUsername := "user"
	dbPassword := "password"
	dbName := "testing"

	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUsername),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container %w", err)
	}

	return postgresContainer, nil
}

func TestDatabase(t *testing.T) {
	ctx := context.Background()
	container, err := createContainer()
	assert.NoError(t, err, "unable to create container")

	connStr, err := container.ConnectionString(ctx)
	assert.NoError(t, err, "unable to get connStr")

	db := database.NewDatabase(connStr)
	err = db.Pool.Ping(ctx)
	assert.NoError(t, err, "Unable to ping the database.")
}
