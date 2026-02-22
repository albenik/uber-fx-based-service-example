-- +goose Up
CREATE TABLE legal_entities (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    tax_id     TEXT NOT NULL,
    deleted_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE legal_entities;
