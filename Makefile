include .env
export

.PHONY: migrate-up migrate-down migrate-create run test

migrate-up:
	@migrate -database "$(DATABASE_URL)" -path migrations up

migrate-down:
	@migrate -database "$(DATABASE_URL)" -path migrations down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

run:
	go run ./cmd/api

test:
	go test ./...
