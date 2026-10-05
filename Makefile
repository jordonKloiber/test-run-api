include .env
export

.PHONY: migrate-up migrate-down migrate-create run test test-integration

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

# Requires DATABASE_URL to point at the dedicated integration-test database,
# never the real app's .env value — see the Phase 2 design doc.
test-integration:
	go test -tags=integration ./...
