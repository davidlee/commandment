# Default recipe
default: deps test demo

lint:
    golangci-lint run ./pkg/... ./examples/...

# Run the basic demo
demo:
    @echo "🚀 Running Commandment Demo..."
    @cd examples/basic && go run main.go

# Run tests
test:
    @echo "🧪 Running tests..."
    @go test ./pkg/... ./examples/...

# Build the demo binary
build:
    @echo "🔨 Building demo..."
    @cd examples/basic && go build -o ../../bin/demo main.go
    @echo "✅ Built: bin/demo"

# Run with verbose logging
demo-verbose:
    @echo "🚀 Running Commandment Demo (verbose)..."
    @cd examples/basic && LOG_LEVEL=debug go run main.go

# Clean build artifacts
clean:
    @echo "🧹 Cleaning..."
    @rm -rf bin/

# Initialize/update dependencies
deps:
    @echo "📦 Updating dependencies..."
    @go mod tidy
    @go mod download

# Show help
help:
    @echo "Available commands:"
    @echo "  demo          - Run the basic demo"
    @echo "  test          - Run tests"
    @echo "  build         - Build demo binary"
    @echo "  demo-verbose  - Run demo with verbose logging"
    @echo "  clean         - Clean build artifacts"
    @echo "  deps          - Update dependencies"
    @echo "  help          - Show this help"
