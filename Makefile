.PHONY: proto proto-clean help

# Generate protobuf and validation code
proto:
	@echo "Generating protobuf code..."
	./build_proto.sh
	@echo "Done!"

# Clean generated proto files
proto-clean:
	@echo "Cleaning generated protobuf files..."
	rm -f api/proto/v1/*.pb.go
	@echo "Done!"

# Install required tools
tools:
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	@echo "Done! Make sure $$GOPATH/bin is in your PATH"

# Run the application
run:
	go run ./cmd/api

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run tests
test:
	go test ./...

# Update dependencies
deps:
	go mod download
	go mod tidy

# Help command
help:
	@echo "Available commands:"
	@echo "  make proto       - Generate protobuf and validation code"
	@echo "  make proto-clean - Remove generated protobuf files"
	@echo "  make tools       - Install required protobuf tools"
	@echo "  make run         - Run the application"
	@echo "  make build       - Build the application"
	@echo "  make test        - Run tests"
	@echo "  make deps        - Update Go dependencies"