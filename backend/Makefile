.PHONY: build test run clean fmt vet

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	go test ./... -v -cover

# Run the server
run:
	go run cmd/server/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Format code
fmt:
	gofmt -w .

# Run go vet
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Install dependencies
deps:
	go mod download
	go mod tidy
