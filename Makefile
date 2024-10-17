BINARY_NAME := bt
CMD_DIR := cmd/$(BINARY_NAME)
BIN_DIR := bin
INSTALL_PATH := ~/.local/bin

all: build

lint:
	@go fmt ./...

build: $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

run:
	@$(BIN_DIR)/$(BINARY_NAME)

install: build
	@mkdir -p $(INSTALL_PATH)
	@cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

clean:
	@rm -rf $(BIN_DIR)

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

.PHONY: lint build run install clean
