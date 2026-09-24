.PHONY: build run test test-cover vet lint tidy vuln docker-build docker-up docker-down

build:
	go build -trimpath -ldflags="-s -w" -o server .

run: build
	./server

test:
	go test -race -cover ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -5

vet:
	go vet ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

docker-build:
	docker build -t go-fiber-api:local .

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v
