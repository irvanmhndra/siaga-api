.PHONY: build run dev migrate-up migrate-down migrate-create migrate-version docker-up docker-down clean

include .env
export

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
MIGRATE=migrate -path migrations -database "$(DATABASE_URL)"

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

dev:
	air

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-version:
	$(MIGRATE) version

migrate-force:
	$(MIGRATE) force $(version)

clean:
	rm -rf bin tmp
