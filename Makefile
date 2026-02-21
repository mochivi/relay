BINARY_NAME := relay
DOCKER_IMAGE := relay
CONFIG_DIR := configs
PORT ?= 8080

.PHONY: build run test clean docker-build docker-run compose-up compose-down help

# Build the relay binary
build:
	go build -o $(BINARY_NAME) ./cmd/relay

# Build and run locally (uses configs/example.yaml)
run: build
	./$(BINARY_NAME)

# Run without building (uses go run)
run-fast:
	go run ./cmd/relay

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Remove built binary and coverage artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run relay in Docker (mounts configs, exposes port)
docker-run: docker-build
	docker run --rm -p $(PORT):$(PORT) -v $(PWD)/$(CONFIG_DIR):/app/$(CONFIG_DIR):ro $(DOCKER_IMAGE)

# Start relay + stub APIs with Docker Compose (builds images, brings up everything)
docker-up:
	docker compose up --build

# Stop and remove Compose stack
docker-down:
	docker compose down

# Tidy and verify modules
deps:
	go mod tidy
	go mod verify

help:
	@echo "Targets:"
	@echo "  build         - build the relay binary"
	@echo "  run           - build and run locally"
	@echo "  run-fast      - run with go run (no build)"
	@echo "  test          - run tests"
	@echo "  test-coverage - run tests and generate coverage report"
	@echo "  clean         - remove binary and coverage artifacts"
	@echo "  docker-build  - build Docker image"
	@echo "  docker-run    - build image and run container (PORT=$(PORT))"
	@echo "  compose-up    - docker compose up --build (relay + APIs)"
	@echo "  compose-down  - docker compose down"
	@echo "  deps          - tidy and verify go modules"
	@echo "  help          - show this help"
