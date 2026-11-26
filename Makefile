include .env
export

APP_NAME=server
OUTPUT_PATH=build
MIGRATION_PATH=./migration
TEST_DSN="host=localhost user=test password=test dbname=test port=9999 sslmode=disable"
.PHONY: build run clean test migrate-up migrate-down migrate-status

all: build run

build:
	mkdir -p $(OUTPUT_PATH)
	go build -o $(OUTPUT_PATH)/$(APP_NAME) cmd/server/server.go

run: build
	./$(OUTPUT_PATH)/$(APP_NAME)

test:
	go test -count 1 -race -parallel 20 -p 20 ./...

test-cover:
	mkdir report
	GOTOOLCHAIN=go1.25.0+auto go test -parallel 20 -p 20 -cover ./...
	GOTOOLCHAIN=go1.25.0+auto go test -parallel 20 -p 20 ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html || echo "coverage.html готов "
	mv coverage.html coverage.out report


integration-test-courier:
	@echo "Starting test Postgres container..."
	docker compose up -d postgres-test

	@echo "Waiting for Postgres to be ready..."
	sleep 3

	@echo "Running migrations on test database..."
	goose -dir $(MIGRATION_PATH) postgres $(TEST_DSN) up

	@echo "Running COURIER integration tests..."
	GOTOOLCHAIN=go1.25.0+auto go test -tags=integration ./internal/repo/courier -v


integration-test-delivery:
	@echo "Starting test Postgres container..."
	docker compose up -d postgres-test

	@echo "Waiting for Postgres to be ready..."
	sleep 3

	@echo "Running migrations on test database..."
	goose -dir $(MIGRATION_PATH) postgres $(TEST_DSN) up

	@echo "Running DELIVERY integration tests..."
	GOTOOLCHAIN=go1.25.0+auto go test -tags=integration ./internal/repo/delivery -v

proto-gen:
	 protoc --go_out=. --go-grpc_out=. internal/gateway/order/order.proto

clean:
	rm -rf $(OUTPUT_PATH)
	rm -rf report/
migrate-up:
	goose -dir $(MIGRATION_PATH) postgres "host=localhost user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) port=$(POSTGRES_PORT) sslmode=disable" up

migrate-down:
	goose -dir $(MIGRATION_PATH) postgres "host=localhost user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) port=$(POSTGRES_PORT) sslmode=disable" down

migrate-status:
	goose -dir $(MIGRATION_PATH) postgres "host=localhost user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) port=$(POSTGRES_PORT) sslmode=disable" status

linter-install:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0
	golangci-lint version

style-check:
	golangci-lint run