.PHONY: help build run test clean install dev docker

# Default target
help:
	@echo "CloudGenie Backend Service - Makefile Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  make install    - Download and install dependencies"
	@echo "  make build      - Build the backend binary"
	@echo "  make run        - Run the backend service"
	@echo "  make dev        - Run in development mode with auto-reload"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make setup-env  - Create .env file from template"
	@echo ""

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed!"

# Build the binary
build:
	@echo "Building cloudgenie-backend..."
	go build -o cloudgenie-backend .
	@echo "Build complete! Binary: ./cloudgenie-backend"

# Build for production (optimized)
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o cloudgenie-backend .
	@echo "Production build complete!"

# Run the service
run: build
	@echo "Starting CloudGenie Backend Service..."
	./cloudgenie-backend

# Run in development mode
dev:
	@echo "Starting in development mode..."
	go run main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f cloudgenie-backend
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Setup environment file
setup-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from template..."; \
		cp .env.example .env; \
		echo ".env file created! Please edit it with your actual values."; \
	else \
		echo ".env file already exists!"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Initialize the project (first-time setup)
init: setup-env install
	@echo ""
	@echo "Project initialized!"
	@echo "Next steps:"
	@echo "  1. Edit .env file with your configuration"
	@echo "  2. Build the MCP server (see README.md)"
	@echo "  3. Run 'make build' to build the backend"
	@echo "  4. Run 'make run' to start the service"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t cloudgenie-backend:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8081:8081 --env-file .env cloudgenie-backend:latest
