# .env создаётся tripgoctl environment start; значения экспортируем в рецепты.
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

run:
	go run ./cmd/trip-service/main.go

migrate:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is not set. Run 'tripgoctl environment start' first."; exit 1)
	go tool goose -dir migrations postgres "$(DATABASE_URL)" up

generate:
	go tool oapi-codegen -generate types,strict-server,chi-server -include-tags Trips,Service -package api -o api/generated/api.gen.go contracts/openapi/trip-service.openapi.yaml

lint:
	go tool golangci-lint run

test-unit-cover:
	go tool cover -func=coverage.out | tail -1
test-unit:
	go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
test-integration:
	go test -race -count=1 -tags integration -coverpkg=./internal/... -coverprofile=coverage.integration.out ./internal/repository/... ./internal/queries/...
test:
	make test-unit
	make test-integration
	make test-unit-cover
