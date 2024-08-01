lint:
	go fmt ./...

build:
	go build .

run:
	go run .

debug: build run

