-- +goose Up
CREATE TABLE drivers (
    id             UUID PRIMARY KEY,
    first_name     TEXT NOT NULL,
    last_name      TEXT NOT NULL,
    license_number TEXT NOT NULL,
    deleted_at     TIMESTAMPTZ
);

-- +goose Down
DROP TABLE drivers;
