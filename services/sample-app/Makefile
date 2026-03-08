BIN=sample-app
PORT?=8080

.PHONY: build run up down test

build:
	go build -o bin/$(BIN) .

run: build
	PORT=$(PORT) ./bin/$(BIN)

up: build
	PORT=$(PORT) ./bin/$(BIN) &
	@echo "✓ sample-app up on :$(PORT)"

down:
	pkill -f bin/$(BIN) || true

test:
	go test ./...
