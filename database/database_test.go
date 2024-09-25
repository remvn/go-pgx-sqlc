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
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "unable to create test database")

	author, err := db.Query.CreateAuthor(ctx, sqlc.CreateAuthorParams{
		Name: "Rob Pike",
	})
	require.NoError(t, err, "unable to create author")
	t.Logf("created author %s with id %d\n", author.Name, author.ID)

	authors, err := db.Query.ListAuthors(ctx)
	require.NoError(t, err, "unable to get author lists")
	t.Logf("author table has %d row\n", len(authors))
}

func TestPgx(t *testing.T) {
	db, err := createTestDb()
	require.NoError(t, err, "unable to create test database")

	insertedAuthor, err := pgxInsert(db, "Rob Pike", pgtype.Text{})
	require.NoError(t, err, "unable to create author")

	selectedAuthor, err := pgxSelect(db, insertedAuthor.Id)
	require.NoError(t, err, "unable to get author")

	require.Equal(t, insertedAuthor.Name, selectedAuthor.Name)
}

func TestPgxBulkInsert(t *testing.T) {
	db, err := createTestDb()
	require.NoError(t, err, "unable to create test database")

	authors := []Author{
		{Name: "User 1"},
		{Name: "User 2"},
		{Name: "User 3"},
	}

	copyCount, err := pgxCopyInsert(db, authors)
	require.NoError(t, err, "unable to copy")
	require.Equal(t, int(copyCount), len(authors))
}

// this struct is for demonstration of
// pgx collect rows helper function
type Author struct {
	Id   int         `db:"id"`
	Name string      `db:"name"`
	Bio  pgtype.Text `db:"bio"`
}

func pgxInsert(db *database.Database, name string, bio pgtype.Text) (Author, error) {
	// use named arguments instead $1, $2, $3...
	query := `INSERT INTO author (name, bio) VALUES (@name, @bio) RETURNING *`
	args := pgx.NamedArgs{
		"name": name,
		"bio":  bio,
	}
	rows, err := db.Pool.Query(context.Background(), query, args)
	if err != nil {
		return Author{}, nil
	}
	defer rows.Close()

	// use collect helper function instead of scanning rows
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Author])
}

func pgxCopyInsert(db *database.Database, authors []Author) (int64, error) {
	rows := [][]any{}
	columns := []string{"name", "bio"}
	tableName := "author"

	for _, author := range authors {
		rows = append(rows, []any{author.Name, author.Bio})
	}

	return db.Pool.CopyFrom(
		context.Background(),
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(rows),
	)
}

func pgxSelect(db *database.Database, id int) (Author, error) {
	// notice that I dont select id
	// and use RowToStructByNameLax to allows some of the column missing
	query := `SELECT name, bio from author where id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}

	rows, err := db.Pool.Query(context.Background(), query, args)
	if err != nil {
		return Author{}, err
	}
	defer rows.Close()

	return pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Author])
}
