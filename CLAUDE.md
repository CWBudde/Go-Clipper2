# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Project Status and Milestones

This project is **~70% complete** and follows a milestone-driven implementation
plan documented in `PLAN.md`.

### Current Status (as of 2025-10-21)

**✅ Complete and Working:**

- M0: Foundation (project structure, build system, CI/CD)
- M1: Rectangle clipping + CGO oracle infrastructure
- M2: Core geometry kernel (128-bit math, segment intersection, winding numbers)
- CGO Oracle: **100% functional** - all 11/11 tests passing, production-ready

**🔧 In Progress (M3 - Boolean Operations):**

- Vatti scanline algorithm **implemented** (~600 lines in `vatti_engine.go`)
- Tests pass but produce **incorrect results** (needs debugging)
- Algorithm structure exists, debugging needed to match oracle behavior

**❌ Not Started:**

- M4: Polygon offsetting (pure Go)
- M5: Completeness features
- M6: Production polish and optimization

### Key Implementation Notes

- **CGO oracle is production-ready**: Use `-tags=clipper_cgo` for fully
  functional Clipper2 in Go
- **Pure Go boolean ops exist but are buggy**: Algorithm implemented but
  debugging needed
- **Only `InflatePaths64` returns `ErrNotImplemented`**: All other APIs have
  implementations (though some are incorrect)
- Always validate against CGO oracle when implementing/debugging
- Reference C++ implementation is in `third_party/clipper2/CPP/Clipper2Lib/src/`

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

- **`port/`** - Pure Go implementation (~70% complete)
  - `clipper.go` - Public API and type definitions
  - `impl_pure.go` - Pure Go implementation entry points
  - `vatti_engine.go` - Vatti scanline algorithm (🔧 debugging needed)
  - `vertex.go` - Vertex chain and local minima detection
  - `geometry.go` - Geometry utilities (✅ complete)
  - `math128.go` - Robust 128-bit integer math (✅ complete)
  - `rectangle_clipping.go` - RectClip64 (✅ complete)
- **`capi/`** - CGO bindings (✅ 100% functional, production-ready)
  - `clipper_cgo.go` - Go↔C bridge functions
  - `clipper_bridge.{h,cc}` - C wrapper for C++ Clipper2 API
- **`third_party/clipper2/`** - Vendored Clipper2 C++ source (reference
  implementation)

### Build Tags System

- **Default (pure Go)**: `go build ./...`
  - Uses `port/impl_pure.go` implementations
  - Boolean operations work but produce incorrect results
  - Rectangle clipping and utilities fully functional

- **CGO Oracle Mode**: `go build -tags=clipper_cgo ./...`
  - Uses `port/impl_oracle_cgo.go` which delegates to C++ library
  - **100% functional** - all operations work correctly
  - Recommended for production use until pure Go debugging complete

### Development Workflow

1. **Write tests first** in `port/clipper_test.go`
2. **Validate with oracle**: Tests must pass with `-tags=clipper_cgo` (oracle is
   ground truth)
3. **Implement pure Go**: Write/debug algorithm in `port/impl_pure.go` or
   specialized files
4. **Validate implementation**: Compare pure Go results against oracle
5. **Debug discrepancies**: Add logging, compare with C++ reference in
   `third_party/clipper2/`

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

**Pure Go Mode (default build):**

- ✅ **Fully Working**: Area64, IsPositive64, Reverse64, RectClip64
- ✅ **Complete Infrastructure**: 128-bit math, geometry kernel, all fill rules
- 🔧 **Debugging Needed**: Union64, Intersect64, Difference64, Xor64 (algorithm
  exists but produces wrong results)
- ❌ **Not Implemented**: InflatePaths64 (returns `ErrNotImplemented`)

**CGO Oracle Mode (`-tags=clipper_cgo`):**

- ✅ **100% Functional**: All operations work correctly
- ✅ **Production-Ready**: All 11/11 tests passing
- ✅ **Full Feature Set**: Boolean ops, offsetting, rectangle clipping, all
  utilities

## Testing Strategy

- Test with oracle first: `just test-oracle`
- Pure Go tests will mostly skip until implemented
- Use property-based testing patterns for geometric operations
- Validate numerical precision with integer coordinates
- Test edge cases: degenerate polygons, boundary conditions, large coordinates

## Code Quality and Linting

Always run `just lint` before committing code. The project uses golangci-lint
with strict rules including gocritic for style enforcement.

### Common Lint Issues to Avoid

**Function Signatures:**

- Use combined return types when identical: `func name() (a, b Type, err error)`
  instead of `func name() (a Type, b Type, err error)`

**Control Flow:**

- Prefer switch statements over long if-else chains for better readability
- Eliminate duplicate branch bodies - consolidate identical logic

**Modern Go Syntax:**

- Use new octal literal format `0o12` instead of deprecated `012`
- Follow modern Go conventions for clarity

**Unused Fields:**

- During development, prefix unused struct fields with `_` to indicate
  intentional non-use
- Remove or implement unused fields before final implementation

### Pre-commit Workflow

```bash
just dev-oracle    # Full development workflow with validation
just lint          # Quick lint check
just lint-fix      # Attempt automatic fixes where possible
```

## Current Debugging Focus (M3)

The Vatti scanline algorithm in `port/vatti_engine.go` is implemented but
produces incorrect results. When working on debugging:

1. **Add detailed logging** to trace algorithm execution
2. **Start with simple test cases** (2 overlapping rectangles)
3. **Compare with reference**:
   `third_party/clipper2/CPP/Clipper2Lib/src/clipper.engine.cpp`
4. **Focus on one operation** at a time (start with Intersection)
5. **Oracle is always right** - if pure Go differs from oracle, pure Go is wrong

Key areas to investigate:

- Local minima detection (vertex.go)
- Active edge list management (vatti_engine.go)
- Intersection point calculation
- Output polygon building and linking
- Winding count calculations for different ClipTypes

## Adding New Operations

1. Add function signature to `port/clipper.go`
2. Add `ErrNotImplemented` stub to `port/impl_pure.go`
3. Add CGO delegation to `port/impl_oracle_cgo.go`
4. Add CGO binding to `capi/clipper_cgo.go` if needed
5. Write comprehensive tests in `port/clipper_test.go`
6. Validate with oracle before implementing pure Go version

## Production Use Recommendations

**If you need Clipper2 in Go right now:**

- Use CGO oracle mode (`-tags=clipper_cgo`) - it's production-ready and fully
  functional
- All operations work correctly with native C++ performance
- Fully tested and validated (11/11 tests passing)

**If you need zero C++ dependencies:**

- Rectangle clipping (`RectClip64`) works perfectly in pure Go
- Basic utilities (Area64, IsPositive64, Reverse64) fully functional
- Boolean operations are implemented but need debugging - contributions welcome!
- Polygon offsetting (`InflatePaths64`) not yet implemented in pure Go

See `PLAN.md` for detailed implementation roadmap and current debugging status.
