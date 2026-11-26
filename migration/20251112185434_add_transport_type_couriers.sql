-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE  couriers  ADD COLUMN IF NOT EXISTS transport_type TEXT NOT NULL DEFAULT 'on_foot';
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE couriers
    DROP COLUMN IF EXISTS transport_type;
-- +goose StatementEnd
