-- +goose Up
ALTER TABLE urls
ADD COLUMN user_id INT NOT NULL REFERENCES users(id);


-- +goose Down
ALTER TABLE urls
DROP COLUMN user_id;
