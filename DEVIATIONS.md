# Deviations from C++ Clipper2

This document catalogs all intentional deviations between the Go Clipper2 port
and the upstream
[C++ Clipper2 library](https://github.com/AngusJohnson/Clipper2). Understanding
these differences is essential when migrating code from C++ Clipper2 to this Go
implementation.

## Overview

Go Clipper2 aims for **functional equivalence** with C++ Clipper2 while
embracing Go idioms and conventions. The core clipping algorithms produce
identical results (validated via CGO oracle testing), but the API differs in
several intentional ways to provide a more natural Go experience.

## Major Deviations

### 1. No Floating-Point API (PathsD/PointD)

**C++ Clipper2 provides:**

- Integer precision types: `Point64`, `Path64`, `Paths64`
- Floating-point types: `PointD`, `PathD`, `PathsD`
- `ClipperD` class for direct floating-point operations with precision parameter
- Automatic scaling between integer and floating-point representations

**Go Clipper2 provides:**

- **Only** integer precision types: `Point64`, `Path64`, `Paths64`
- **No** floating-point equivalents

**Rationale:** The integer-based API eliminates floating-point precision errors
and provides deterministic results. All geometric operations use exact integer
arithmetic or robust 128-bit calculations for numerical stability.

**Migration Path:** Users needing floating-point coordinates must handle scaling
manually:

```go
// C++: automatic with ClipperD
ClipperD clipper(2); // 2 decimal places
clipper.AddSubject(floatingPointPaths);
clipper.Execute(ClipType::Union, fillRule, solution);

// Go: manual scaling required
scale := 100.0 // 2 decimal places = 10^2
scaled := make(clipper.Paths64, len(floatingPointPaths))
for i, path := range floatingPointPaths {
    scaled[i] = make(clipper.Path64, len(path))
    for j, pt := range path {
        scaled[i][j] = clipper.Point64{
            X: int64(pt.X * scale),
            Y: int64(pt.Y * scale),
        }
    }
}
result, _ := clipper.Union64(scaled, clips, clipper.NonZero)
// Unscale result back to float64 if needed
```

### 2. Simplified Functional API Only

**C++ Clipper2 provides:**

1. **Class-based API** (full control):
   ```cpp
   Clipper64 clipper;
   clipper.AddSubject(subjects);
   clipper.AddClip(clips);
   clipper.Execute(ClipType::Union, FillRule::NonZero, solution);
   ```
2. **Simplified functional API** (convenience):
   ```cpp
   solution = Union(subjects, clips, FillRule::NonZero);
   ```

**Go Clipper2 provides:**

- **Only simplified functional API**:
  ```go
  solution, err := clipper.Union64(subjects, clips, clipper.NonZero)
  ```

**Missing from Go:**

- Stateful `Clipper64` class
- Incremental path addition (`AddSubject`, `AddClip`, `AddOpenSubject`)
- Reusable clipper instances for multiple operations
- `ReuseableDataContainer64` for preprocessing paths

**Rationale:** The functional API covers 95% of use cases with a simpler, more
Go-idiomatic interface. The stateful class API adds complexity that most users
don't need.

**Migration Path:** Most C++ code using `Clipper64` class can be simplified:

```go
// C++: class-based approach
Clipper64 clipper;
for (const auto& path : manyPaths) {
    clipper.AddSubject(path);
}
clipper.AddClip(clipPath);
clipper.Execute(ClipType::Union, FillRule::NonZero, solution);

// Go: combine all subjects first
allSubjects := clipper.Paths64{}
for _, path := range manyPaths {
    allSubjects = append(allSubjects, path)
}
solution, err := clipper.Union64(allSubjects, clipPath, clipper.NonZero)
```

**Advanced users needing stateful operations**: Consider using CGO oracle mode
(`-tags=clipper_cgo`) which delegates to full C++ library, or file a feature
request for specific use cases.

### 3. Rectangle Clipping Interface Difference

**C++ Clipper2:**

```cpp
Rect64 rect(0, 0, 100, 100);
Paths64 result = RectClip(rect, paths);
Paths64 lines = RectClipLines(rect, openPaths); // Separate function for lines
```

**Go Clipper2:**

```go
rect := clipper.Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}
result, err := clipper.RectClip64(rect, paths)
lines, err := clipper.RectClipLines64(rect, openPaths) // ✅ Available
```

**Differences:**

1. Go takes `Path64` (4 points) instead of `Rect64` struct
2. Go validates exactly 4 points, returns `ErrInvalidRectangle` otherwise
3. Both `RectClip64` (closed paths) and `RectClipLines64` (open paths) are
   implemented
4. 32-bit variants also available: `RectClip32` and `RectClipLines32`

**Rationale:** Using `Path64` directly avoids introducing another struct type
and leverages existing path validation logic. The Cohen-Sutherland algorithm
provides efficient line clipping with O(1) per-segment complexity.

### 4. No Z-Coordinate Support

**C++ Clipper2:**

- Optional 3D support via `USINGZ` compilation flag
- `Point64::z` field when enabled
- `ZCallback64` for custom z-interpolation at intersections

**Go Clipper2:**

- 2D only: `Point64{X, Y int64}`
- No Z coordinate or callback support

**Rationale:** 2D covers the overwhelming majority of use cases. Adding Z
support would complicate the implementation and API surface without clear
demand.

**Future Enhancement:** If 3D use cases emerge, could add `Point64Z` types and
separate API functions (e.g., `Union64Z`).

## Minor Deviations

### 5. Function Naming Convention

**C++ Clipper2:**

```cpp
Paths64 Union(subjects, clips, fillrule);
Paths64 Intersect(subjects, clips, fillrule);
Paths64 Difference(subjects, clips, fillrule);
Paths64 Xor(subjects, clips, fillrule);
Paths64 InflatePaths(paths, delta, joinType, endType);
```

**Go Clipper2:**

```go
Union64(subjects, clips, fillrule) (Paths64, error)
Intersect64(subjects, clips, fillrule) (Paths64, error)
Difference64(subjects, clips, fillrule) (Paths64, error)
Xor64(subjects, clips, fillrule) (Paths64, error)
InflatePaths64(paths, delta, joinType, endType, ...opts) (Paths64, error)
```

**Differences:**

- Go adds `64` suffix to all precision-dependent functions
- Go returns `(result, error)` tuples per Go conventions

**Rationale:** The `64` suffix explicitly signals integer precision and leaves
room for future precision variants (if needed). Error returns are idiomatic Go.

### 6. Union Without Clips

**C++ Clipper2:**

```cpp
// Union subjects with themselves (dissolve overlaps)
Paths64 result = Union(subjects, FillRule::NonZero);
```

**Go Clipper2:**

```go
// Must provide empty clips explicitly
result, err := clipper.Union64(subjects, clipper.Paths64{}, clipper.NonZero)
```

**Rationale:** Go doesn't support function overloading. Requiring explicit empty
clips is more explicit and avoids hidden behavior.

### 7. Error Handling Philosophy

**C++ Clipper2:**

- Throws `Clipper2Exception` (when exceptions enabled)
- Returns error codes via mutable reference parameters
- Silent failure when exceptions disabled

**Go Clipper2:**

- Returns idiomatic `error` as last return value
- Specific error types for different conditions:
  - `ErrInvalidFillRule` - fill rule out of range (0-3)
  - `ErrInvalidClipType` - clip type out of range (0-3)
  - `ErrInvalidJoinType` - join type out of range (0-3)
  - `ErrInvalidEndType` - end type out of range (0-4)
  - `ErrInvalidParameter` - invalid numeric parameter (epsilon ≤ 0, etc.)
  - `ErrInvalidOptions` - invalid option values (miterLimit ≤ 0, etc.)
  - `ErrInvalidRectangle` - rectangle doesn't have exactly 4 points
  - `ErrEmptyPath` - nil or empty path where valid path required
  - `ErrDegeneratePolygon` - polygon with < 3 points
  - `ErrNotImplemented` - feature not yet implemented in pure Go

**Example:**

```go
result, err := clipper.Union64(subjects, clips, fillRule)
switch {
case errors.Is(err, clipper.ErrNotImplemented):
    log.Println("Use -tags=clipper_cgo for full functionality")
case errors.Is(err, clipper.ErrInvalidFillRule):
    log.Printf("Invalid fill rule: %v", fillRule)
case err != nil:
    log.Printf("Unexpected error: %v", err)
default:
    // Process result
}
```

### 8. Offset Options Configuration

**C++ Clipper2:**

```cpp
ClipperOffset co(
    /*miter_limit=*/2.0,
    /*arc_tolerance=*/0.25,
    /*preserve_collinear=*/false,
    /*reverse_solution=*/false
);
co.AddPaths(paths, JoinType::Round, EndType::Polygon);
co.Execute(delta, solution);

// Or with simplified API:
solution = InflatePaths(paths, delta, JoinType::Round, EndType::Polygon,
                        /*miter_limit=*/2.0, /*arc_tolerance=*/0.25);
```

**Go Clipper2:**

```go
// Simplified API with optional configuration
result, err := clipper.InflatePaths64(
    paths, delta,
    clipper.JoinRound, clipper.EndPolygon,
    clipper.OffsetOptions{
        MiterLimit:        2.0,
        ArcTolerance:      0.25,
        PreserveCollinear: false,
        ReverseSolution:   false,
    },
)

// Default options (variadic parameter can be omitted)
result, err := clipper.InflatePaths64(paths, delta, clipper.JoinRound, clipper.EndPolygon)
```

**Rationale:** Go's struct-based options provide named parameters and default
values without function overloading. Variadic parameters allow omitting the
struct entirely for defaults.

### 9. PolyTree Memory Management

**C++ Clipper2:**

```cpp
PolyTree64 polytree;
Clipper64 clipper;
clipper.Execute(ClipType::Union, FillRule::NonZero, polytree);
// std::unique_ptr manages child lifetimes
```

**Go Clipper2:**

```go
polytree, openPaths, err := clipper.Union64Tree(subjects, clips, clipper.NonZero)
// Garbage collector manages lifetimes
defer polytree.Clear() // Optional, for large trees
```

**Differences:**

- C++ uses `std::unique_ptr<PolyPath64>` for children
- Go uses native pointers with garbage collection
- Both provide identical tree structure and navigation API

**API Compatibility:** Navigation methods are nearly identical:

- `Level()`, `IsHole()`, `Parent()`, `Child(i)`, `Count()`, `Polygon()`

## Missing Helper Functions

The following C++ helper functions are not currently implemented in Go Clipper2.
Most are convenience functions that can be implemented in user code if needed.

### Geometric Utilities

| C++ Function                               | Purpose                  | Workaround in Go                                 |
| ------------------------------------------ | ------------------------ | ------------------------------------------------ |
| `TranslatePath(path, dx, dy)`              | Translate path by offset | Loop and add offset to each point                |
| `TranslatePaths(paths, dx, dy)`            | Translate multiple paths | Loop over paths and points                       |
| `Ellipse(center, radiusX, radiusY, steps)` | Generate ellipse path    | Implement using trigonometry                     |
| `Length(path, is_closed)`                  | Calculate path perimeter | Sum distances between points                     |
| `Distance(pt1, pt2)`                       | Euclidean distance       | `math.Sqrt(Sqr(pt1.X-pt2.X) + Sqr(pt1.Y-pt2.Y))` |

### Path Cleanup Utilities

| C++ Function                                     | Purpose                      | Workaround in Go              |
| ------------------------------------------------ | ---------------------------- | ----------------------------- |
| `TrimCollinear(path, is_open)`                   | Remove collinear points      | Implement cross product check |
| `StripNearEqual(path, max_dist_sqrd, is_closed)` | Remove near-duplicate points | Filter with distance check    |
| `StripDuplicates(path, is_closed)`               | Remove exact duplicates      | Loop and compare points       |

### Simplification Algorithms

| C++ Function                             | Purpose                               | Go Equivalent                              |
| ---------------------------------------- | ------------------------------------- | ------------------------------------------ |
| `SimplifyPath(path, epsilon, is_closed)` | Perpendicular distance simplification | `SimplifyPath64(path, epsilon, is_closed)` |
| `RamerDouglasPeucker(path, epsilon)`     | RDP simplification                    | Not implemented (use SimplifyPath64)       |

**Note:** Go's `SimplifyPath64` uses perpendicular distance algorithm
(Visvalingam-Whyatt inspired), while C++ provides both this and
Ramer-Douglas-Peucker.

### Analysis Functions

| C++ Function                                       | Purpose                     | Workaround in Go                      |
| -------------------------------------------------- | --------------------------- | ------------------------------------- |
| `Path2ContainsPath1(path1, path2)`                 | Containment test            | Use `PointInPolygon64` for all points |
| `NearCollinear(pt1, pt2, pt3, sin_sqrd_min_angle)` | Collinearity with tolerance | Implement cross product check         |

### Rectangle Clipping

| C++ Function                 | Purpose               | Status in Go                             |
| ---------------------------- | --------------------- | ---------------------------------------- |
| `RectClip(rect, paths)`      | Clip closed paths     | `RectClip64(rect Path64, paths)` ✅      |
| `RectClipLines(rect, lines)` | Clip open paths/lines | `RectClipLines64(rect Path64, lines)` ✅ |

## Go-Specific Enhancements

While this document focuses on deviations, Go Clipper2 includes several
enhancements over C++ Clipper2:

### 1. Dual Implementation Architecture

**Unique to Go:**

```bash
# Pure Go implementation (portable, no C++ dependencies)
go test ./port

# CGO oracle mode (validates against C++ library)
go test ./port -tags=clipper_cgo
```

- **Pure Go mode**: Zero C++ dependencies, fully portable
- **CGO oracle mode**: 100% C++ Clipper2 compatibility for validation
- Build tag system ensures identical API between modes

### 2. 32-bit Coordinate Support

**Go-Specific Enhancement** (not present in C++ Clipper2):

Go Clipper2 provides full API support for 32-bit integer coordinates alongside
the standard 64-bit types. This enables seamless integration with 32-bit
graphics APIs, game engines, and systems where memory efficiency is important.

**Available Types:**

```go
Point32  // 32-bit coordinate point
Path32   // 32-bit coordinate path
Paths32  // 32-bit coordinate paths
Rect32   // 32-bit coordinate rectangle
PolyTree32  // 32-bit hierarchical polygon tree
```

**API Coverage:** All public API functions have 32-bit variants with `32`
suffix:

- Boolean operations: `Union32`, `Intersect32`, `Difference32`, `Xor32`,
  `BooleanOp32`
- Tree operations: `Union32Tree`, `Intersect32Tree`, `Difference32Tree`,
  `Xor32Tree`
- Offsetting: `InflatePaths32`
- Utilities: `Area32`, `Bounds32`, `RectClip32`, `SimplifyPath32`, etc.
- Minkowski: `MinkowskiSum32`, `MinkowskiDiff32`

**Implementation:**

- **Internal processing**: All operations are performed in 64-bit for numerical
  stability
- **Automatic conversion**: 32-bit inputs are promoted to 64-bit at API
  boundaries
- **Overflow detection**: Results are validated to fit in int32 range before
  conversion back
- **Error handling**: `ErrInt32Overflow` or `ErrResultOverflow` returned when
  values exceed int32 limits

**Conversion utilities:**

```go
// Safe conversions with overflow detection
pt32, err := Point64ToPoint32(pt64)  // Returns ErrInt32Overflow if out of range
path32, err := Path64ToPath32(path64)
paths32, err := Paths64ToPaths32(paths64)

// Always-safe promotions
pt64 := Point32ToPoint64(pt32)  // No overflow possible
path64 := Path32ToPath64(path32)
paths64 := Paths32ToPaths64(paths32)
```

**Example Usage:**

```go
// Working with 32-bit coordinates (e.g., from a graphics API)
subject32 := Paths32{
    {{0, 0}, {1920, 0}, {1920, 1080}, {0, 1080}},  // Screen coordinates
}
clip32 := Paths32{
    {{100, 100}, {500, 100}, {500, 400}, {100, 400}},
}

// Perform operation - automatically handled in 64-bit internally
result32, err := Intersect32(subject32, clip32, NonZero)
if err == ErrResultOverflow {
    // Result doesn't fit in int32 range
    // Fall back to 64-bit API
    result64, _ := Intersect64(Paths32ToPaths64(subject32), Paths32ToPaths64(clip32), NonZero)
}
```

**Coordinate Ranges:**

- Valid range: `-2,147,483,648` to `2,147,483,647` (int32 limits)
- Overflow errors occur when conversion results exceed these bounds
- Consider using 64-bit API if your coordinates or results may exceed int32
  range

**Why 32-bit support?**

1. **Memory efficiency**: 50% reduction in coordinate storage for large datasets
2. **API compatibility**: Direct integration with 32-bit graphics APIs (OpenGL,
   DirectX, SDL, etc.)
3. **Integer constraints**: Applications that know coordinates fit in int32
   range can enforce this at type level
4. **Performance**: Smaller memory footprint can improve cache efficiency

**Note:** C++ Clipper2 does not provide 32-bit coordinate types. This is a
Go-specific enhancement for better ecosystem integration.

### 3. Comprehensive Input Validation

Go implementation validates inputs and returns specific errors:

```go
// Automatic degenerate path filtering
Union64([]Path64{{p1, p2}}, clips, NonZero) // < 3 points, automatically filtered

// Enum validation
Union64(subjects, clips, FillRule(99)) // Returns ErrInvalidFillRule

// Struct field validation
InflatePaths64(paths, delta, jt, et, OffsetOptions{
    MiterLimit: -1.0, // Returns ErrInvalidOptions
})
```

C++ generally assumes valid input or exhibits undefined behavior.

### 3. Type Safety

Go's type system prevents entire classes of errors:

```go
// Compile-time type safety
var fillRule FillRule = 2 // OK
var fillRule FillRule = -1 // Compile error

// C++ allows:
FillRule fillRule = static_cast<FillRule>(999); // Compiles, undefined behavior
```

## Migration Guide

### Quick Reference Table

| Task           | C++ Clipper2                                   | Go Clipper2                                          |
| -------------- | ---------------------------------------------- | ---------------------------------------------------- |
| Union          | `Union(subj, clip, fr)`                        | `Union64(subj, clip, fr)`                            |
| Class-based    | `Clipper64 c; c.AddSubject(s); c.Execute(...)` | Combine paths, use functional API                    |
| Float paths    | `ClipperD(precision); Execute(pathsD, ...)`    | Scale manually, use `Paths64`                        |
| Rect clip      | `RectClip(Rect64{l,t,r,b}, paths)`             | `RectClip64(Path64{{l,t},{r,t},{r,b},{l,b}}, paths)` |
| Error handling | `try { ... } catch (Clipper2Exception& e)`     | `result, err := ...; if err != nil { ... }`          |
| Z-coordinates  | `Point64{x, y, z}` (with `USINGZ`)             | Not supported                                        |

### Common Migration Patterns

#### Pattern 1: Simple Boolean Operation

```cpp
// C++
Paths64 result = Intersect(subjects, clips, FillRule::NonZero);
```

```go
// Go
result, err := clipper.Intersect64(subjects, clips, clipper.NonZero)
if err != nil {
    return fmt.Errorf("intersection failed: %w", err)
}
```

#### Pattern 2: Polygon Offsetting

```cpp
// C++
Paths64 expanded = InflatePaths(paths, 10.0, JoinType::Round, EndType::Polygon,
                                 2.0, 0.25);
```

```go
// Go
expanded, err := clipper.InflatePaths64(
    paths, 10.0,
    clipper.JoinRound, clipper.EndPolygon,
    clipper.OffsetOptions{
        MiterLimit:   2.0,
        ArcTolerance: 0.25,
    },
)
if err != nil {
    return fmt.Errorf("offset failed: %w", err)
}
```

#### Pattern 3: Hierarchical Output

```cpp
// C++
PolyTree64 polytree;
Clipper64 clipper;
clipper.AddSubject(subjects);
clipper.AddClip(clips);
clipper.Execute(ClipType::Union, FillRule::NonZero, polytree);
```

```go
// Go
polytree, openPaths, err := clipper.Union64Tree(subjects, clips, clipper.NonZero)
if err != nil {
    return fmt.Errorf("union tree failed: %w", err)
}
// Note: openPaths will be empty for closed polygons
```

#### Pattern 4: Floating-Point Coordinates

```cpp
// C++
PathsD pathsD = { {{0.5, 1.5}, {2.3, 4.7}, ...} };
PathsD result = Union(pathsD, clipsD, FillRule::NonZero, 2);
```

```go
// Go
// Scale to integers (2 decimal places = 10^2)
scale := 100.0
paths64 := make(clipper.Paths64, len(pathsD))
for i, path := range pathsD {
    paths64[i] = make(clipper.Path64, len(path))
    for j, pt := range path {
        paths64[i][j] = clipper.Point64{
            X: int64(pt.X * scale),
            Y: int64(pt.Y * scale),
        }
    }
}

result64, err := clipper.Union64(paths64, clips64, clipper.NonZero)

// Unscale back to float if needed
resultD := make([][]PointD, len(result64))
for i, path := range result64 {
    resultD[i] = make([]PointD, len(path))
    for j, pt := range path {
        resultD[i][j] = PointD{
            X: float64(pt.X) / scale,
            Y: float64(pt.Y) / scale,
        }
    }
}
```

## Future Compatibility

### Planned Enhancements

These C++ features may be added in future releases:

1. **Path utility functions** - TranslatePath, TrimCollinear, etc.
2. **Additional simplification algorithms** - RamerDouglasPeucker

### Not Planned

The following are unlikely to be implemented without significant demand:

1. **Float/double API (PathsD)** - Integer API is sufficient for most uses
2. **Z-coordinate support** - 2D covers 95%+ of use cases
3. **Stateful Clipper64 class** - Functional API is simpler and sufficient

## Reporting Issues

If a C++ Clipper2 operation produces different results than Go Clipper2:

1. **Test with CGO oracle**: Run tests with `-tags=clipper_cgo` to validate
   against official C++ library
2. **Check PLAN.md**: Feature may be in progress or marked as debugging
3. **File an issue**: Include minimal test case showing discrepancy

If you need a missing C++ feature:

1. **Check this document**: May have suggested workaround
2. **File a feature request**: Describe use case and why workarounds
   insufficient
3. **Contribute a PR**: Implementations with tests are always welcome!

## See Also

- [README.md](README.md) - Usage examples and getting started
- [PLAN.md](PLAN.md) - Implementation roadmap and status
- [Upstream Clipper2 Documentation](http://www.angusj.com/clipper2/Docs/Overview.htm)
