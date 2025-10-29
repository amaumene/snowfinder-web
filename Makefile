.PHONY: build run docker-up docker-down clean

# Build the web server binary
build:
	go build -o bin/snowfinder-web ./cmd/server

# Run the web server
run:
	go run ./cmd/server

# Start PostgreSQL (optional if using scraper's DB)
docker-up:
	docker-compose up -d

# Stop PostgreSQL
docker-down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf bin/

# Default DATABASE_URL
DATABASE_URL ?= postgres://snowfinder:snowfinder@localhost:5432/snowfinder?sslmode=disable

# Default PORT
PORT ?= 8080
