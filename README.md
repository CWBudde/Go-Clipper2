# Go Clipper2

A pure Go port of [Clipper2](https://github.com/AngusJohnson/Clipper2) polygon clipping and offsetting library.

## Features

- **Pure Go Implementation**: No C/C++ dependencies for the main library
- **CGO Oracle**: Optional CGO bindings for testing and validation against the original C++ implementation
- **Robust Integer Arithmetic**: Uses 64-bit integers for geometric robustness
- **Full Clipper2 API**: Boolean operations, polygon offsetting, and line clipping

## Operations

- **Boolean Operations**: Union, Intersection, Difference, XOR
- **Polygon Offsetting**: Inflate/deflate with various join types (round, miter, square)
- **Line Clipping**: Fast rectangular clipping

## Installation

```bash
go get github.com/you/clipper2-go
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/you/clipper2-go/port"
)

func main() {
    // Define two overlapping rectangles
    subject := clipper.Paths64{
        {{0, 0}, {10, 0}, {10, 10}, {0, 10}},
    }
    clip := clipper.Paths64{
        {{5, 5}, {15, 5}, {15, 15}, {5, 15}},
    }
    
    // Compute union
    result, err := clipper.Union64(subject, clip, clipper.NonZero)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Union result: %v\n", result)
}
```

## Testing with CGO Oracle

To run tests that validate against the original Clipper2 C++ library:

### macOS
```bash
brew install clipper2
go test ./... -tags=clipper_cgo
```

### Linux (Fedora)
```bash
sudo dnf install clipper2-devel
go test ./... -tags=clipper_cgo
```

### Linux (Ubuntu/Debian via vcpkg)
```bash
vcpkg install clipper2
# Set CGO_CXXFLAGS/CGO_LDFLAGS to point to vcpkg paths
go test ./... -tags=clipper_cgo
```

## License

This project is licensed under the Boost Software License 1.0, same as the original Clipper2 library.

## Acknowledgments

This is a Go port of the excellent [Clipper2](https://github.com/AngusJohnson/Clipper2) library by Angus Johnson.