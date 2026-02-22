-- +goose Up
CREATE TABLE vehicles (
    id            UUID PRIMARY KEY,
    fleet_id      UUID NOT NULL REFERENCES fleets(id),
    make          TEXT NOT NULL,
    model         TEXT NOT NULL,
    year          INTEGER NOT NULL,
    license_plate TEXT NOT NULL,
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX idx_vehicles_fleet_id ON vehicles(fleet_id);

-- +goose Down
DROP TABLE vehicles;
