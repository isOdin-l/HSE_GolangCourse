.PHONY: migrate run test lint generate

run:


GOOSE_DRIVER=postgres
GOOSE_DBSTRING="postgres://user:pass@localhost/trip"
migrate:
	goose ${GOOSE_DRIVER} ${GOOSE_DBSTRING} -dir migrations up

generate:
	oapi-codegen -generate types,chi-server -package api -o internal/generated/api.gen.go contracts/openapi/trip-service.openapi.yaml

test:
	go test ./...

lint:
