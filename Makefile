.PHONY: setup build test lint docker-up docker-down

GO=go
GOLINT=golangci-lint

setup:
	$(GO) work sync

build:
	$(GO) build ./...

test:
	$(GO) test -v -race ./...

lint:
	$(GOLINT) run ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

test-cover:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
