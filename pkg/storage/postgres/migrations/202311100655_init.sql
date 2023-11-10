-- +goose Up
CREATE TABLE IF NOT EXISTS passwords (
    guid VARCHAR(36) NOT NULL,
    hashed_password VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(guid)
);

-- +goose Down
DROP TABLE IF EXISTS passwords;