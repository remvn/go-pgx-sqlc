-- name: GetAuthor :one
SELECT *
FROM author
WHERE id = $1
LIMIT 1;

-- name: ListAuthors :many
SELECT *
FROM author
ORDER BY name;

-- name: CreateAuthor :one
INSERT INTO author (name, bio)
VALUES (lower(@name), @bio)
RETURNING *;
