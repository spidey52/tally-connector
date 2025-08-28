PORT ?= 8084
BIN_DIR=bin

APP_NAME=api

.PHONY: all build clean dev run

all: build

# Build API for production
build-api:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) cmd/api/main.go

# Build all (if you have multiple apps later)
build: build-api

# Run development with live reload
dev:
	cd cmd/api && air -c ../../.air.toml

loader-dev:
	cd cmd/loader && air -c ../../.air.toml

# Run production binary
run: build
	$(BIN_DIR)/$(APP_NAME) --port=$(PORT)

# Clean all binaries
clean:
	rm -rf $(BIN_DIR)
	cd cmd/loader/reqlog && rm -r request/*  && rm -r response/*
