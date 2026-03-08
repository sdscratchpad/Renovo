BIN=batch-worker

.PHONY: build run up down test

build:
	go build -o bin/$(BIN) .

run: build
	./bin/$(BIN)

up: build
	./bin/$(BIN) &
	@echo "✓ batch-worker up"

down:
	pkill -f bin/$(BIN) || true

test:
	go test ./...
