CREATE TABLE author (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    bio  TEXT
);

CREATE TABLE book (
    id          SERIAL PRIMARY KEY,
    author_id   INT REFERENCES author (id),
    name        VARCHAR(200) NOT NULL,
    description TEXT
)
