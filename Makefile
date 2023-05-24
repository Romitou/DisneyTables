.PHONY: all build run test vet lint clean docker-build docker-run help

BINARY_NAME=disneytables
BUILD_DIR=builds

all: vet test build

build:
	@echo "==> Building binary..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

run:
	@echo "==> Running application..."
	go run main.go

test:
	@echo "==> Running tests..."
	go test -v -race ./...

vet:
	@echo "==> Running go vet..."
	go vet ./...

docker-build:
	@echo "==> Building Docker image..."
	docker build -t disneytables:latest .

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
