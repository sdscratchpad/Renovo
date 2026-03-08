PORT?=3000

.PHONY: install up down build test

install:
	npm install

up: install
	PORT=$(PORT) npm start &
	@echo "✓ ui up on :$(PORT)"

down:
	pkill -f "react-scripts start" || pkill -f "vite" || true

build:
	npm run build

test:
	npm test -- --watchAll=false
