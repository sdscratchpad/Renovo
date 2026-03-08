BIN=diagnosis
PORT?=8083

.PHONY: build run up down test

build:
	go build -o bin/$(BIN) .

run: build
	PORT=$(PORT) ./bin/$(BIN)

up: build
	@test -n "$$GITHUB_TOKEN" || (echo "ERROR: GITHUB_TOKEN is not set" && exit 1)
	PORT=$(PORT) GITHUB_TOKEN=$$GITHUB_TOKEN ./bin/$(BIN) &
	@echo "✓ diagnosis up on :$(PORT)"

down:
	pkill -f bin/$(BIN) || true

test:
	go test ./...
