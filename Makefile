.PHONY: run build test tidy swag migrate-up migrate-down

SWAG ?= swag

run:
	go run ./cmd/app

build:
	go build -o bin/flysoft-flight-service ./cmd/app

test:
	go test ./...

tidy:
	go mod tidy

swag:
	$(SWAG) init -g cmd/app/main.go -o docs

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
