.PHONY: build run test lint docker-up docker-down

build:
	go build -ldflags="-s -w" -o server .

run: build
	./server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v
