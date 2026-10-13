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
	go tool oapi-codegen -generate types,chi-server -include-tags Trips,Service -package api -o api/generated/api.gen.go contracts/openapi/trip-service.openapi.yaml

test:
	go test -race ./...

#lint:
