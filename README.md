# Go Clipper2

A high-performance pure Go port of [Clipper2](https://github.com/AngusJohnson/Clipper2), the industry-standard polygon clipping and offsetting library. Go Clipper2 provides robust geometric operations with 64-bit integer precision, eliminating floating-point numerical errors common in computational geometry.

## 🌟 Features

- **🚀 Pure Go Implementation**: Zero C/C++ dependencies for production use
- **🔧 Dual Architecture**: CGO oracle for development validation and testing
- **⚡ Robust Arithmetic**: 64-bit integer coordinates prevent numerical instability
- **🎯 Complete API**: All Clipper2 operations including boolean ops, offsetting, and clipping
- **🧪 Comprehensive Testing**: Property-based testing with fuzzing and golden reference validation
- **📦 Easy Integration**: Simple Go module with clean, idiomatic API

## 🛠️ Development Setup

### Prerequisites
- Go 1.22 or later
- Git with submodule support
- C++ compiler (for CGO oracle testing only)

### Quick Start
```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/go-clipper/clipper2
cd clipper2

# Test pure Go implementation (most operations will skip until implemented)
go test ./port -v

# Build and verify
go build ./...
go vet ./...
```

### CGO Oracle Setup (Development & Testing)

The CGO oracle validates pure Go implementations against the original C++ library.

#### macOS
```bash
brew install clipper2
go test ./... -tags=clipper_cgo -v
```

#### Linux (Fedora/CentOS)
```bash
sudo dnf install clipper2-devel
go test ./... -tags=clipper_cgo -v
```

#### Linux (Ubuntu/Debian)
```bash
# Option 1: Build from source
cd third_party/clipper2
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release -DCLIPPER2_EXAMPLES=OFF -DCLIPPER2_TESTS=OFF
make -j$(nproc)
sudo make install
sudo ldconfig

# Option 2: Use vcpkg
vcpkg install clipper2
export CGO_CXXFLAGS="-I$(vcpkg list clipper2 | grep include)"
export CGO_LDFLAGS="-L$(vcpkg list clipper2 | grep lib) -lClipper2"
go test ./... -tags=clipper_cgo -v
```

#### Windows
```bash
# Using MSYS2/MinGW
pacman -S mingw-w64-x86_64-cmake mingw-w64-x86_64-toolchain
cd third_party/clipper2
mkdir build && cd build
cmake .. -G "MSYS Makefiles" -DCMAKE_BUILD_TYPE=Release
make && make install

# Or build the Visual Studio solution in DLL/CPP_DLL/
```

## 📖 Usage Examples

### Basic Boolean Operations

```go
package main

import (
    "fmt"
    "github.com/go-clipper/clipper2/port"
)

func main() {
    // Define two overlapping rectangles
    subject := clipper.Paths64{
        {{0, 0}, {100, 0}, {100, 100}, {0, 100}},
    }
    clip := clipper.Paths64{
        {{50, 50}, {150, 50}, {150, 150}, {50, 150}},
    }
    
    // Union: combine both shapes
    union, err := clipper.Union64(subject, clip, clipper.NonZero)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Union area: %.0f\n", clipper.Area64(union[0]))
    
    // Intersection: overlapping area only
    intersection, err := clipper.Intersect64(subject, clip, clipper.NonZero)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Intersection area: %.0f\n", clipper.Area64(intersection[0]))
    
    // Difference: subject minus clip
    difference, err := clipper.Difference64(subject, clip, clipper.NonZero)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Difference paths: %d\n", len(difference))
    
    // XOR: symmetric difference
    xor, err := clipper.Xor64(subject, clip, clipper.NonZero)
    if err != nil {
        panic(err)
    }
    fmt.Printf("XOR paths: %d\n", len(xor))
}
```

### Working with Complex Polygons

```go
// Polygon with hole
outer := clipper.Path64{{0, 0}, {200, 0}, {200, 200}, {0, 200}}
hole := clipper.Path64{{50, 50}, {50, 150}, {150, 150}, {150, 50}}

// Ensure correct orientation (outer CCW, hole CW)
if !clipper.IsPositive64(outer) {
    outer = clipper.Reverse64(outer)
}
if clipper.IsPositive64(hole) {
    hole = clipper.Reverse64(hole)
}

subject := clipper.Paths64{outer, hole}
clip := clipper.Paths64{{{100, 100}, {300, 100}, {300, 300}, {100, 300}}}

result, err := clipper.Union64(subject, clip, clipper.NonZero)
if err != nil {
    panic(err)
}
```

### Polygon Offsetting (Expansion/Contraction)

```go
// Original shape
shape := clipper.Paths64{
    {{50, 50}, {150, 50}, {150, 150}, {50, 150}},
}

// Expand by 10 units with rounded corners
expanded, err := clipper.InflatePaths64(shape, 10.0, clipper.Round, clipper.ClosedPolygon)
if err != nil {
    panic(err)
}

// Contract by 5 units with square corners
contracted, err := clipper.InflatePaths64(shape, -5.0, clipper.Square, clipper.ClosedPolygon,
    clipper.OffsetOptions{
        MiterLimit:   2.0,
        ArcTolerance: 0.25,
    })
if err != nil {
    panic(err)
}
```

### Error Handling Patterns

```go
import "errors"

result, err := clipper.Union64(subject, clip, clipper.NonZero)
switch {
case errors.Is(err, clipper.ErrNotImplemented):
    log.Println("Feature not yet implemented in pure Go")
case errors.Is(err, clipper.ErrInvalidInput):
    log.Printf("Invalid input parameters: %v", err)
case err != nil:
    log.Printf("Unexpected error: %v", err)
default:
    // Process result
    fmt.Printf("Success: got %d paths\n", len(result))
}
```

## 📚 API Reference

### Core Types

```go
type Point64 struct {
    X, Y int64  // 64-bit integer coordinates for precision
}

type Path64 []Point64    // Sequence of points forming a path
type Paths64 []Path64    // Collection of paths (polygons with holes)
```

### Boolean Operations

```go
// Primary operations (simplified interface)
func Union64(subjects, clips Paths64, fillRule FillRule) (Paths64, error)
func Intersect64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) 
func Difference64(subjects, clips Paths64, fillRule FillRule) (Paths64, error)
func Xor64(subjects, clips Paths64, fillRule FillRule) (Paths64, error)

// Advanced operation (full control)
func BooleanOp64(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution, solutionOpen Paths64, err error)
```

### Fill Rules

Controls how polygon interiors are determined:

- `EvenOdd`: Odd-numbered regions are filled (simple toggle)
- `NonZero`: Non-zero winding regions are filled (default for most cases)
- `Positive`: Only positive winding regions are filled
- `Negative`: Only negative winding regions are filled

### Offsetting Operations

```go
func InflatePaths64(paths Paths64, delta float64, joinType JoinType, endType EndType, opts ...OffsetOptions) (Paths64, error)

// Join types for connecting segments
const (
    Square JoinType = iota  // Sharp corners
    Round                   // Rounded corners  
    Miter                   // Mitered corners (with limit)
)

// End types for open paths
const (
    ClosedPolygon EndType = iota  // Closed polygon paths
    ClosedLine                    // Closed line paths
    OpenSquare                    // Square end caps
    OpenRound                     // Round end caps
    OpenButt                      // Flat end caps
)
```

### Utility Functions

```go
func Area64(path Path64) float64              // Signed area (positive = CCW)
func IsPositive64(path Path64) bool           // True if counter-clockwise
func Reverse64(path Path64) Path64            // Reverse point order
func RectClip64(rect Path64, paths Paths64) (Paths64, error)  // Fast rectangular clipping
```

## 📊 Implementation Status

| Feature | Pure Go | CGO Oracle | Status |
|---------|---------|------------|--------|
| Boolean Operations | ❌ | ✅ | In Development |
| Union64 | ❌ | ✅ | Planned |
| Intersect64 | ❌ | ✅ | Planned |
| Difference64 | ❌ | ✅ | Planned |
| Xor64 | ❌ | ✅ | Planned |
| Polygon Offsetting | ❌ | ❌ | Planned |
| Rectangle Clipping | ❌ | ❌ | Planned |
| Area Calculation | ✅ | ✅ | Complete |
| Orientation Detection | ✅ | ✅ | Complete |
| Path Reversal | ✅ | ✅ | Complete |
| Minkowski Operations | ❌ | ❌ | Future |

**Legend**: ✅ Implemented, ❌ Not implemented, 🚧 In progress

## 🤝 Contributing & Development

### Development Workflow

1. **Write Tests First**: Add test cases to `port/clipper_test.go`
2. **Validate with Oracle**: Ensure tests pass with `-tags=clipper_cgo`
3. **Implement Pure Go**: Replace stubs in `port/impl_pure.go`
4. **Validate Implementation**: Tests should pass in both modes

### Running Tests

```bash
# Test pure Go implementation
go test ./port -v

# Test specific function
go test ./port -run TestArea64 -v

# Test with CGO oracle (validation)
go test ./... -tags=clipper_cgo -v

# Run specific oracle test
go test ./capi -run TestUnionTiny -tags=clipper_cgo -v

# Benchmark when available
go test -bench=. ./port
```

### Code Architecture

The project uses Go build tags for dual implementation:

- `port/impl_pure.go` (no build tag): Pure Go implementations
- `port/impl_oracle_cgo.go` (`//go:build clipper_cgo`): Delegates to CGO oracle  
- `capi/` (all files have `//go:build clipper_cgo`): CGO bindings

### Adding New Operations

1. Add function signature to `port/clipper.go`
2. Add stub returning `ErrNotImplemented` to `port/impl_pure.go`
3. Add CGO delegation to `port/impl_oracle_cgo.go` 
4. Add CGO binding to `capi/clipper_cgo.go` if needed
5. Add comprehensive tests to `port/clipper_test.go`

## 🐛 Troubleshooting

### CGO Oracle Build Issues

**"clipper2/clipper.export.h: No such file"**
```bash
# Ensure Clipper2 is installed with headers
brew install clipper2  # macOS
dnf install clipper2-devel  # Fedora
```

**"undefined reference to BooleanOp64"**  
```bash
# Check library installation
pkg-config --cflags --libs Clipper2

# On Linux, may need:
sudo ldconfig
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
```

**Windows CGO Issues**
```bash
# Use MSYS2 environment
set CGO_ENABLED=1
set CC=gcc

# Or use the pre-built DLL approach (see docs)
```

### Common Errors

**`ErrNotImplemented`**: Feature not yet ported to pure Go. Use CGO oracle for testing or wait for implementation.

**`ErrInvalidInput`**: Check polygon orientation, ensure paths are not self-intersecting for basic operations.

**Memory Issues**: When using CGO oracle, ensure proper cleanup. The library handles this automatically.

### Performance Tips

- Use integer coordinates when possible (more robust than float64)
- For simple rectangular clipping, use `RectClip64` instead of boolean operations
- Pre-simplify complex polygons before operations
- Consider polygon orientation for optimal performance

## 🔗 Related Projects

- [Clipper2](https://github.com/AngusJohnson/Clipper2) - Original C++ library
- [geos](https://github.com/twpayne/go-geos) - Go bindings for GEOS
- [orb](https://github.com/paulmach/orb) - 2D geometry types and utilities

## 📄 License

This project is licensed under the **Boost Software License 1.0**, the same as the original Clipper2 library. See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

This project is a port of the excellent [Clipper2](https://github.com/AngusJohnson/Clipper2) library by Angus Johnson. Special thanks to the Clipper2 community for creating such a robust and well-designed computational geometry library.