lint:
	go fmt ./...

build:
	go build -o ./bin/bt ./cmd/bt/main.go

run:
	go run ./cmd/bt/main.go

install: build
	mkdir -p ~/.local/bin
	cp ./bin/bt ~/.local/bin/bt
