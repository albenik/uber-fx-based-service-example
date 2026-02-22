-- +goose Up
CREATE TABLE fleets (
    id              UUID PRIMARY KEY,
    legal_entity_id UUID NOT NULL REFERENCES legal_entities(id),
    name            TEXT NOT NULL,
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_fleets_legal_entity_id ON fleets(legal_entity_id);

-- +goose Down
DROP TABLE fleets;
