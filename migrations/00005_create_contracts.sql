-- +goose Up
CREATE TABLE contracts (
    id              UUID PRIMARY KEY,
    driver_id       UUID NOT NULL REFERENCES drivers(id),
    legal_entity_id UUID NOT NULL REFERENCES legal_entities(id),
    fleet_id        UUID NOT NULL REFERENCES fleets(id),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    terminated_at   TIMESTAMPTZ,
    terminated_by   TEXT NOT NULL DEFAULT '',
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_contracts_driver_id ON contracts(driver_id);
CREATE INDEX idx_contracts_overlap ON contracts(driver_id, legal_entity_id, fleet_id);

-- +goose Down
DROP TABLE contracts;
