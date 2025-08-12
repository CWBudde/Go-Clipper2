# Go Clipper2 Development Commands
# Run 'just --list' to see all available commands

# Default recipe - shows help
default:
    @just --list

# Build all packages (pure Go mode)
build:
    go build ./...

# Build with CGO oracle (requires Clipper2 system installation)
build-oracle:
    go build -tags=clipper_cgo ./...

# Run all tests (pure Go mode - most will skip with ErrNotImplemented)
test:
    go test ./... -v

# Run tests with CGO oracle (requires Clipper2 system installation)
test-oracle:
    go test ./... -tags=clipper_cgo -v

# Run only port package tests (pure Go)
test-port:
    go test ./port -v

# Run only capi package tests (CGO oracle)
test-capi:
    go test ./capi -tags=clipper_cgo -v

# Run a specific test by name
test-run name:
    go test ./port -run {{name}} -v

# Run a specific test with CGO oracle
test-run-oracle name:
    go test ./... -run {{name}} -tags=clipper_cgo -v

# Run go vet on all packages
vet:
    go vet ./...

# Format all code using treefmt
fmt:
    treefmt

# Run linting checks
lint:
    golangci-lint run --config .golangci.toml

# Fix linting issues automatically
lint-fix:
    golangci-lint run --config .golangci.toml --fix
    treefmt

# Clean go module cache and build artifacts
clean:
    go clean -cache -modcache -testcache

# Download and tidy dependencies
deps:
    go mod download
    go mod tidy

# Run all checks (build, test, lint)
check: build test lint
    @echo "All checks passed!"

# Run all checks with CGO oracle
check-oracle: build-oracle test-oracle lint
    @echo "All oracle checks passed!"

# Setup development environment (install just if needed)
setup:
    @echo "Installing just (command runner)..."
    @echo "Visit https://github.com/casey/just#installation for installation instructions"

# Install Clipper2 system dependencies for CGO oracle mode
install-clipper2-macos:
    brew install clipper2

install-clipper2-fedora:
    sudo dnf install clipper2-devel

# Build Clipper2 from source (for systems without packages)
install-clipper2-source:
    cd third_party/clipper2 && \
    mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make && sudo make install && \
    sudo ldconfig

# Development workflow commands
dev: fmt lint test
    @echo "Development checks complete!"

dev-oracle: fmt lint test-oracle
    @echo "Development checks with oracle complete!"

# Benchmark commands
bench:
    go test -bench=. ./...

bench-oracle:
    go test -bench=. -tags=clipper_cgo ./...

# Coverage reporting
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

coverage-oracle:
    go test -coverprofile=coverage.out -tags=clipper_cgo ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report with oracle generated: coverage.html"

# Fuzz testing
fuzz:
    go test -fuzz=. ./port

# Quick validation (fastest checks)
quick: build test-port
    @echo "Quick validation complete!"