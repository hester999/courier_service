-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
create index if not exists couriers_available_order_count_idx on couriers(status,order_count);
create index if not exists couriers_status_idx on couriers(status);
create index if not exists delivery_courier_deadline_idx on delivery(courier_id, deadline);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
drop index if exists  couriers_available_order_count_idx;
drop index if exists  couriers_status_idx;
drop index if exists  delivery_courier_deadline_idx;
