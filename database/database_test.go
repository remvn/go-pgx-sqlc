package database_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/remvn/go-pgx-sqlc/database"
	"github.com/remvn/go-pgx-sqlc/database/sqlc"
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

	schemaFile := filepath.Join("../sqlc/", "schema.sql")
	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithInitScripts(schemaFile),
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

func createTestDb() (*database.Database, error) {
	ctx := context.Background()
	container, err := createContainer()
	if err != nil {
		return nil, err
	}

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	db := database.NewDatabase(connStr)
	err = db.Pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestSqlc(t *testing.T) {
	ctx := context.Background()
	db, err := createTestDb()
	assert.NoError(t, err, "unable to create test database")

	author, err := db.Query.CreateAuthor(ctx, sqlc.CreateAuthorParams{
		Name: "Rob Pike",
	})
	assert.NoError(t, err, "unable to create author")
	t.Logf("created author %s with id %d\n", author.Name, author.ID)

	authors, err := db.Query.ListAuthors(ctx)
	assert.NoError(t, err, "unable to get author lists")
	t.Logf("author table has %d row\n", len(authors))
}

func TestPgx(t *testing.T) {
	db, err := createTestDb()
	assert.NoError(t, err, "unable to create test database")

	_, err = pgxInsert(db, "Rob Pike", pgtype.Text{})
	assert.NoError(t, err, "unable to create author")
}

func pgxInsert(db *database.Database, name string, bio pgtype.Text) (any, error) {
	query := `INSERT INTO author (name, bio) VALUES (@name, @bio) RETURNING *`
	args := pgx.NamedArgs{
		"name": name,
		"bio":  bio,
	}
	row := db.Pool.QueryRow(context.Background(), query, args)
	author := &sqlc.Author{}
	err := row.Scan(author)
	if err != nil {
		return nil, err
	}

	return author, nil
}

func pgxSelect(db *database.Database, id int) (any, error) {
	query := `SELECT name, bio from author where id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}
	row := db.Pool.QueryRow(context.Background(), query, args)
	author := &sqlc.Author{}
	err := row.Scan(author)
	if err != nil {
		return nil, err
	}

	return author, nil
}
