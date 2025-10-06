.PHONY: test test-verbose test-coverage test-unit test-integration clean

# Run all tests
test:
	@echo "Running all tests..."
	@go test -timeout=30s ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -timeout=30s ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -timeout=30s -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run only unit tests (excluding integration tests if any)
test-unit:
	@echo "Running unit tests..."
	@go test -v -timeout=30s -short ./...

# Run specific package tests
test-domain:
	@go test -v -timeout=10s ./internal/domain

test-repository:
	@go test -v -timeout=10s ./internal/repositories/sqlite

test-services:
	@go test -v -timeout=10s ./internal/services/...

test-handlers:
	@go test -v -timeout=10s ./internal/api/handlers

test-utils:
	@go test -v -timeout=10s ./internal/util/...

# Clean test cache and coverage files
clean:
	@echo "Cleaning test cache and coverage files..."
	@go clean -testcache
	@rm -f coverage.out coverage.html

# Run tests and show which tests are passing/failing
test-list:
	@echo "Listing all tests..."
	@go test -v -timeout=30s ./... 2>&1 | grep -E "^(PASS|FAIL|---)"

# Run tests in parallel (faster)
test-parallel:
	@echo "Running tests in parallel..."
	@go test -timeout=30s -parallel=4 ./...

# Help target
help:
	@echo "Available test commands:"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-unit     - Run only unit tests"
	@echo "  make test-domain   - Run domain layer tests"
	@echo "  make test-repository - Run repository tests"
	@echo "  make test-services - Run service layer tests"
	@echo "  make test-handlers - Run handler tests"
	@echo "  make test-utils    - Run utility tests"
	@echo "  make clean         - Clean test cache and coverage files"
	@echo "  make test-list     - List all test results"
	@echo "  make test-parallel - Run tests in parallel (faster)"
