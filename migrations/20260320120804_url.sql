-- +goose Up
CREATE TABLE urls(
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    alias VARCHAR(20) NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE urls
