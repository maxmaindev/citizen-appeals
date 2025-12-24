.PHONY: help run build test clean docker-up docker-down migrate-up migrate-down

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	go run cmd/api/main.go

build: ## Build the application
	go build -o bin/api cmd/api/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

migrate-up: ## Run database migrations up
	@echo "TODO: Implement migration tool"

migrate-down: ## Run database migrations down
	@echo "TODO: Implement migration tool"

deps: ## Download dependencies
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
