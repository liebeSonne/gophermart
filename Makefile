export CGO_ENABLED=0

.PHONY: all
all: lint build generate generate-api tests

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build
build:
	go build -o cmd/gophermart/gophermart ./cmd/gophermart

.PHONY: generate
generate:
	go generate ./...

.PHONY: generate-api
generate-api:
	go tool oapi-codegen -config ./api/server/config.yaml ./api/server/public.openapi.yaml
	go tool oapi-codegen -config ./api/client/config.yaml ./api/client/accrual.openapi.yaml

.PHONY: clean-api
clean-api:
	rm -f api/server/public.openapi.gen.go
	rm -f api/client/accrual.openapi.gen.go

.PHONY: tests
tests:
	go test ./...

.PHONY: tests-v
tests-v:
	go test -v ./...

.PHONY: tests-make-coverage
tests-make-coverage:
	go test -coverprofile=coverage.out ./...
	grep -vE "mock|\.gen\.go" coverage.out > coverage.filtered.out

.PHONY: tests-cover
tests-cover: tests-make-coverage
	go tool cover -func=coverage.filtered.out

.PHONY: tests-cover-total
tests-cover-total: tests-make-coverage
	go tool cover -func=coverage.filtered.out | grep total

.PHONY: tests-cover-html
tests-cover-html: tests-make-coverage
	go tool cover -html=coverage.filtered.out

.PHONY: clean-bin
clean-bin:
	rm -f cmd/gophermart/gophermart

.PHONY: create-migration
create-migration:
	# example: make create-migration name=add_table
	migrate create -ext sql -dir ./migrations -format "20060102150405" $(name)

.PHONY: mocks
mocks:
	mockery