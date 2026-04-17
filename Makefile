.PHONY: run build test test-int lint migrate-up migrate-down docker-up docker-down

APP        := server
BIN_DIR    := bin
PKG        := ./...
DB_URL     ?= $(shell grep -E '^DATABASE_URL=' .env | cut -d= -f2-)

run:
	go run ./cmd/server

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP) ./cmd/server

test:
	go test -race -count=1 -short -coverprofile=coverage.out $(PKG)

test-int:
	go test -race -count=1 -tags=integration ./...

lint:
	golangci-lint run --timeout=3m

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v
