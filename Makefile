.PHONY: start stop clean proto mock openapi lint

COMPOSE_FILE = deploy/docker-compose.yaml

start:
	docker compose -f $(COMPOSE_FILE) up -d

stop:
	docker compose -f $(COMPOSE_FILE) stop

clean:
	docker compose -f $(COMPOSE_FILE) down -v

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/analytics/v1/*.proto

mock:
	go run github.com/vektra/mockery/v2@latest

lint:
	golangci-lint run ./...
