run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v -count=1

cover:
	go test ./internal/services/ -cover -count=1

vet:
	go vet ./...

swagger:
	swag init -g cmd/server/main.go -o docs --parseInternal --parseDependency

docker-up:
	docker compose up -d

docker-down:
	docker compose down

.PHONY: run build test cover vet swagger docker-up docker-down
