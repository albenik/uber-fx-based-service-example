-- +goose Up
CREATE TABLE vehicle_assignments (
    id          UUID PRIMARY KEY,
    driver_id   UUID NOT NULL REFERENCES drivers(id),
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_vehicle_assignments_contract_id ON vehicle_assignments(contract_id);
CREATE INDEX idx_vehicle_assignments_driver_id ON vehicle_assignments(driver_id);

-- +goose Down
DROP TABLE vehicle_assignments;
