# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mcp-adapter
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_mac -v

# Docker
docker-build:
	docker build -t mcp-adapter .

docker-run:
	docker run -p 8080:8080 mcp-adapter

# Development
dev:
	air

# Database
migrate-up:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME) migrate up

migrate-down:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME) migrate down

seed:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME) seed

# Swagger
swagger:
	swag init -g main.go

# Linting
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Security check
security:
	gosec ./...

# Generate mocks
mocks:
	mockgen -source=services/interfaces.go -destination=mocks/services.go

# Install tools
install-tools:
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u github.com/cosmtrek/air
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec
	$(GOGET) -u github.com/golang/mock/mockgen

.PHONY: all build test clean run deps build-linux build-windows build-mac docker-build docker-run dev migrate-up migrate-down seed swagger lint fmt security mocks install-tools