-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
alter table couriers add column if not exists order_count bigint not null default 0;
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table couriers drop column order_count;
-- +goose StatementEnd
