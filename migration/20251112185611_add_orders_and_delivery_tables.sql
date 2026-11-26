-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS delivery (
                                        id BIGSERIAL PRIMARY KEY,
                                        courier_id BIGINT NOT NULL REFERENCES couriers(id),
                                        order_id UUID NOT NULL UNIQUE,
                                        assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                        deadline TIMESTAMP NOT NULL
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS delivery;
-- +goose StatementEnd
