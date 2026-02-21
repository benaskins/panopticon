# Panopticon — Apple Silicon Cockpit

default:
    @just --list

# Build the binary
build:
    go build -o pan ./cmd/pan/

# Run the dashboard
run: build
    ./pan

# Run all tests
test:
    go test ./...

# Run tests verbose
test-v:
    go test -v ./...

# Format code
fmt:
    go fmt ./...

# Vet code
vet:
    go vet ./...

# Format + vet
check: fmt vet

# Clean build artifacts
clean:
    rm -f pan
