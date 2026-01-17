.PHONY: build clean fmt test tidy

BINARY_NAME := mcd-cn
BIN_DIR := bin

ifeq ($(OS),Windows_NT)
	BINARY_EXT := .exe
else
	BINARY_EXT :=
endif

BINARY_PATH := $(BIN_DIR)/$(BINARY_NAME)$(BINARY_EXT)

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY_PATH) ./cmd/mcd-cn

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
