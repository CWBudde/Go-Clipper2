# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status and Milestones

This project follows a milestone-driven implementation plan documented in `TASK.md`. Check the task list to understand current implementation status and check mark completed items (✅). Key points:

- Most boolean operations return `ErrNotImplemented` in pure Go mode (until completed)
- Always validate new implementations against the CGO oracle
- Please always check the reference implementation in `third_party/clipper2/` if algorithms are complex or not well understood
- Completed tasks must be tested carefully and checked off in `TASK.md`

## Development Commands

Use the `just` command runner for all development tasks:

```bash
# Build and test
just build                # Build pure Go implementation
just build-oracle         # Build with CGO oracle (requires system Clipper2)
just test                 # Test pure Go (most will skip with ErrNotImplemented)
just test-oracle           # Test with CGO oracle validation
just test-port            # Test only port package (pure Go)
just test-capi            # Test only capi package (CGO)

# Development workflow
just dev                  # Format, lint, and test (pure Go)
just dev-oracle           # Format, lint, and test with oracle
just check                # Full validation (build + test + lint)
just check-oracle         # Full validation with oracle

# Code quality
just fmt                  # Format Go code
just lint                 # Run linting (includes vet and fmt)
just vet                  # Run go vet

# Specific tests
just test-run TestName    # Run specific test (pure Go)
just test-run-oracle TestName  # Run specific test with oracle

# Performance and coverage
just bench                # Benchmark tests
just coverage            # Generate coverage report
just fuzz                # Run fuzz tests
```

## Architecture Overview

This is a dual-implementation project with a unique build tag architecture:

### Core Structure

- **`port/`** - Pure Go implementation (target for production)
- **`capi/`** - CGO bindings to original Clipper2 C++ library (oracle for validation)
- **`third_party/clipper2/`** - Original Clipper2 C++ library source code (submodule, source of truth for algorithms)
- **Build Tags System**:
  - `port/impl_pure.go` (no tag) - Pure Go stubs/implementations
  - `port/impl_oracle_cgo.go` (`//go:build clipper_cgo`) - Delegates to CGO oracle
  - All `capi/` files require `clipper_cgo` build tag

### Development Workflow

1. **Write tests first** in `port/clipper_test.go`
2. **Validate with oracle**: Tests must pass with `-tags=clipper_cgo`
3. **Implement pure Go**: Replace `ErrNotImplemented` stubs in `port/impl_pure.go`
4. **Validate implementation**: Tests pass in both pure Go and oracle modes

### Key Implementation Files

- `port/clipper.go` - Public API and type definitions
- `port/errors.go` - Error constants (`ErrNotImplemented`, `ErrInvalidInput`, etc.)
- `port/impl_pure.go` - Pure Go implementations (many still return `ErrNotImplemented`)
- `port/impl_oracle_cgo.go` - CGO oracle delegations for validation

## API Design Patterns

### Error Handling

Always check for specific error types:

```go
result, err := clipper.Union64(subject, clip, clipper.NonZero)
switch {
case errors.Is(err, clipper.ErrNotImplemented):
    // Feature not yet implemented in pure Go
case errors.Is(err, clipper.ErrInvalidInput):
    // Invalid parameters
case err != nil:
    // Other errors
}
```

### Core Types

- `Point64` - 64-bit integer coordinates for numerical stability
- `Path64` - Sequence of points forming a path
- `Paths64` - Collection of paths (polygons with holes)
- Fill rules: `EvenOdd`, `NonZero`, `Positive`, `Negative`
- Boolean operations: `Union`, `Intersection`, `Difference`, `Xor`

## Implementation Status

Most boolean operations and polygon offsetting return `ErrNotImplemented` in pure Go mode. Only basic utilities like `Area64`, `IsPositive64`, and `Reverse64` are fully implemented. The CGO oracle provides complete functionality for validation and testing.

## Testing Strategy

- Test with oracle first: `just test-oracle`
- Pure Go tests will mostly skip until implemented
- Use property-based testing patterns for geometric operations
- Validate numerical precision with integer coordinates
- Test edge cases: degenerate polygons, boundary conditions, large coordinates

## Code Quality and Linting

Always run `just lint` before committing code. The project uses golangci-lint with strict rules including gocritic for style enforcement.

### Common Lint Issues to Avoid:

**Function Signatures:**
- Use combined return types when identical: `func name() (a, b Type, err error)` instead of `func name() (a Type, b Type, err error)`

**Control Flow:**
- Prefer switch statements over long if-else chains for better readability
- Eliminate duplicate branch bodies - consolidate identical logic

**Modern Go Syntax:**
- Use new octal literal format `0o12` instead of deprecated `012`
- Follow modern Go conventions for clarity

**Unused Fields:**
- During development, prefix unused struct fields with `_` to indicate intentional non-use
- Remove or implement unused fields before final implementation

### Pre-commit Workflow:

```bash
just dev-oracle    # Full development workflow with validation
just lint          # Quick lint check
just lint-fix      # Attempt automatic fixes where possible
```

## Adding New Operations

1. Add function signature to `port/clipper.go`
2. Add `ErrNotImplemented` stub to `port/impl_pure.go`
3. Add CGO delegation to `port/impl_oracle_cgo.go`
4. Add CGO binding to `capi/clipper_cgo.go` if needed
5. Write comprehensive tests in `port/clipper_test.go`
6. Validate with oracle before implementing pure Go version