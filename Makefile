.PHONY: run build migrate-up migrate-down tidy

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

migrate-up:
	migrate -path db/migrations -database "${DATABASE_URL}" up

migrate-down:
	migrate -path db/migrations -database "${DATABASE_URL}" down

# Install golang-migrate: brew install golang-migrate
# Install deps: make tidy
