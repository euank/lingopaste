.PHONY: build run test clean deps

build:
	go build -o server ./cmd/server

run:
	go run ./cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -f server
	go clean

deps:
	go mod download
	go mod tidy
