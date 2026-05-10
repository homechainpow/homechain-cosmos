# HomeChain V10 Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=homechaind
BUILD_DIR=./build

# Version info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build tags
BUILD_TAGS=netgo,ledger
LDFLAGS=-ldflags "-X github.com/homechain/homechain/version.Version=$(VERSION) \
	-X github.com/homechain/homechain/version.Commit=$(COMMIT) \
	-X github.com/homechain/homechain/version.BuildTime=$(BUILD_TIME) \
	-s -w"

# Docker
DOCKER_IMAGE=homechain/homechain
DOCKER_TAG=$(VERSION)

# All targets
.PHONY: all build clean test coverage deps proto-gen buf-gen build-argon2 help

all: build

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -mod=mod -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/homechaind

# Build for Linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -mod=mod -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/homechaind

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run tests
test:
	$(GOTEST) -mod=mod -v ./...

# Run tests with coverage
coverage:
	$(GOTEST) -mod=mod -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@docker run --rm -v $(CURDIR):/workspace -w /workspace \
		bufbuild/buf:1.28.1 generate

# Generate with buf (alternative)
buf-gen:
	@echo "Generating with buf..."
	@buf generate

# Build Argon2 C library for determinism
build-argon2:
	@echo "Building Argon2 C library for cross-language determinism..."
	@mkdir -p cdeps/argon2
	@if [ ! -d "cdeps/argon2/src" ]; then \
		cd cdeps && git clone https://github.com/P-H-C/phc-winner-argon2.git argon2; \
	fi
	@cd cdeps/argon2 && make && make check
	@mkdir -p lib include
	@cp cdeps/argon2/libargon2.a lib/
	@cp cdeps/argon2/include/argon2.h include/
	@echo "Argon2 C library built successfully"

# Install binary
install: build
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# Start local node
start: build
	$(BUILD_DIR)/$(BINARY_NAME) start

# Initialize node
init:
	$(BUILD_DIR)/$(BINARY_NAME) init homechain-node --chain-id homechain_9000-1

# Add key
add-key:
	$(BUILD_DIR)/$(BINARY_NAME) keys add validator

# Run local network
localnet-start:
	@if ! [ -f $(BUILD_DIR)/node0/homechain/config/genesis.json ]; then docker-compose up; fi
	docker-compose up

# Stop local network
localnet-stop:
	docker-compose down

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run simulation
simulate:
	@echo "Running simulation..."
	$(GOTEST) -mod=mod -tags "$(BUILD_TAGS)" -run TestFullAppSimulation ./app/ -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=1 -Period=5 -v -timeout 24h

# Docker build
docker-build:
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

# Docker push
docker-push:
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-linux   - Build for Linux"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage"
	@echo "  deps          - Download dependencies"
	@echo "  proto-gen     - Generate protobuf files"
	@echo "  buf-gen       - Generate with buf"
	@echo "  build-argon2  - Build Argon2 C library"
	@echo "  install       - Install binary"
	@echo "  start         - Start local node"
	@echo "  init          - Initialize node"
	@echo "  add-key       - Add validator key"
	@echo "  localnet-start- Start local network"
	@echo "  localnet-stop - Stop local network"
	@echo "  lint          - Lint code"
	@echo "  fmt           - Format code"
	@echo "  simulate      - Run simulation"
	@echo "  help          - Show this help"
