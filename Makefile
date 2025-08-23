BIN_DIR=bin

.PHONY: all build clean

all: build

# Build app1
build-app1:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/app1 cmd/api/main.go

# Build all
build: build-app1 build-app2

# Clean all binaries
clean:
	rm -rf $(BIN_DIR)
