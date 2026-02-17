build:
	@go build -o bin/catalog cmd/main.go

migration:
	@migrate create -ext sql -dir internal/infra/db/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

DB_URL=mysql://root:mysqlroot@tcp(localhost:3306)/catalog-db
MIGRATE_CMD=migrate -path internal/infra/db/migrate/migrations -database "$(DB_URL)"

migrate-up:
	@$(MIGRATE_CMD) up

migrate-down:
	@$(MIGRATE_CMD) down 1

migrate-force:
	@$(MIGRATE_CMD) force $(filter-out $@,$(MAKECMDGOALS))

test:
	@go test -v ./...

test-cover:
	@go test -count=1 -coverprofile coverage.out ./... && go tool cover -func=coverage.out

test-cover-html: test-cover
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

run: build
	@./bin/catalog

proto:
	@mkdir -p internal/adapters/input/grpc/pb
	@protoc --go_out=internal/adapters/input/grpc/pb --go_opt=paths=source_relative \
	--go-grpc_out=internal/adapters/input/grpc/pb --go-grpc_opt=paths=source_relative \
	--proto_path=internal/adapters/input/grpc/proto \
	catalog_service.proto