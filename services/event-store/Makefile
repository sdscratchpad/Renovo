BIN=event-store
PORT?=8085
DB_PATH?=./data/events.db

.PHONY: build run up down test

build:
	go build -o bin/$(BIN) .

run: build
	PORT=$(PORT) DB_PATH=$(DB_PATH) ./bin/$(BIN)

up: build
	mkdir -p data
	PORT=$(PORT) DB_PATH=$(DB_PATH) ./bin/$(BIN) &
	@echo "✓ event-store up on :$(PORT)"

down:
	pkill -f bin/$(BIN) || true

## Wipe the database and restart clean.
reset: down
	rm -f data/events.db
	$(MAKE) up
	@echo "\u2713 event-store reset"

test:
	go test ./...
