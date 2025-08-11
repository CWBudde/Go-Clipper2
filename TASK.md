# Go Clipper2 Implementation Milestones

This document outlines the milestone-based implementation plan for porting Clipper2 to pure Go with validated results against the CGO oracle.

## Current Status
- ✅ Project structure and build system established
- ✅ CGO oracle bindings implemented for validation
- ✅ Basic utility functions (Area64, IsPositive64, Reverse64) 
- ✅ Rectangle clipping (RectClip64) fully implemented with comprehensive tests and fuzz validation
- ⏳ Core boolean operations are stub implementations returning `ErrNotImplemented`

---

## M0 — Polish the Foundation (Quick Wins) 🏗️
**Goal: Establish a solid, validated foundation with proper CI/CD**

### Tasks
- [x] Validate module path & imports (fix README placeholder paths in code snippets)
- [x] Define stable public API surface in `port/` (ensure all types, enums, error values are complete)
- [x] Ensure every API function has:
  - [x] Pure Go stub returning `ErrNotImplemented` in `port/impl_pure.go`
  - [x] Oracle implementation behind `//go:build clipper_cgo` in `port/impl_oracle_cgo.go`
  - [x] Table-driven tests that run in oracle mode and skip (not fail) in pure mode
- [x] Set up CI pipeline for Linux/macOS testing both build tags
- [x] Verify README examples compile with real import path

**Done When:** CI green on Linux/macOS for both build tags; README examples compile with the real import path; all API functions have proper stubs and oracle implementations.

---

## M1 — Pure Go Rectangle Clipping (Sutherland-Hodgman) 📐
**Goal: Implement first complete algorithm as proof-of-concept**

### Tasks
- [x] Research and implement Sutherland-Hodgman clipping algorithm for `RectClip64`
- [x] Handle both closed and open paths correctly
- [x] Add edge case handling (degenerate rectangles, empty inputs, etc.)
- [x] Implement property-based tests comparing pure vs. oracle across:
  - [x] Random rectangles and random paths
  - [x] Edge cases (points on boundaries, fully inside/outside)
- [x] Document any degenerate cases where results intentionally differ
- [x] Implement native Go fuzz test (Go 1.18+ `FuzzXxx` functions) to achieve ≥99% match rate between pure Go and CGO oracle implementations

**Done When:** Pure and oracle results match for ≥99% of fuzz cases; remaining cases are documented as acceptable degenerates; `RectClip64` is production-ready.

---

## M2 — Core Geometry Kernel & Robust Integer Math 🔧
**Goal: Build bulletproof mathematical foundation for all operations**

### Tasks
- [ ] Implement robust 128-bit intermediate math using `math/bits` for:
  - [ ] Cross products without overflow
  - [ ] Area accumulations
  - [ ] Distance calculations
- [ ] Add robust segment intersection detection:
  - [ ] Handle collinear overlaps correctly
  - [ ] Manage floating-point edge cases
  - [ ] Include parallel and near-parallel segments
- [ ] Implement winding number and point-in-polygon tests consistent with Clipper2's fill rules
- [ ] Create comprehensive geometry kernel test suite covering:
  - [ ] Parallel/collinear/degenerate cases
  - [ ] Numerical edge cases near overflow
  - [ ] All fill rule behaviors

**Done When:** Kernel unit tests cover all tricky parallel/collinear/degenerate cases; fuzz tests pass; numerical stability validated across range of inputs.

---

## M3 — Pure Go Boolean Operations (Scanline + Active Edge List) ✂️
**Goal: Implement complete boolean operations using proven algorithm structure**

### Tasks
- [ ] Port upstream algorithm structure:
  - [ ] Event queue (local minima detection)
  - [ ] Active edge list management
  - [ ] Intersection ordering and processing
  - [ ] Output polygon builder
- [ ] Implement all fill rules correctly:
  - [ ] EvenOdd (odd-numbered sub-regions)
  - [ ] NonZero (non-zero winding sub-regions)
  - [ ] Positive (positive winding sub-regions)  
  - [ ] Negative (negative winding sub-regions)
- [ ] Start with simple cases (disjoint/overlapping rectangles) then generalize to:
  - [ ] Complex polygons with holes
  - [ ] Self-intersecting polygons
  - [ ] Open paths (lines)
- [ ] Extensive validation against oracle implementation

**Done When:** Pure vs. oracle parity across large randomized test suite; boolean operations produce identical path counts and nearly identical point sets (allowing tiny tolerance on point ordering).

---

## M4 — Pure Go Offsetting (ClipperOffset) 📏
**Goal: Complete polygon offsetting with all join and end types**

### Tasks
- [ ] Implement core offsetting algorithm for both expansion and contraction
- [ ] Add all join types:
  - [ ] Round joins with configurable arc approximation
  - [ ] Miter joins with miter limit control
  - [ ] Square joins for sharp corners
- [ ] Support all end types for closed and open paths:
  - [ ] ClosedPolygon and ClosedLine
  - [ ] OpenSquare, OpenRound, OpenButt end caps
- [ ] Implement precision controls:
  - [ ] MiterLimit for controlling spike length
  - [ ] ArcTolerance for round approximation quality
- [ ] Validate against oracle across grid of:
  - [ ] Various deltas (positive and negative)
  - [ ] All join type combinations
  - [ ] Noisy inputs with self-intersections

**Done When:** Parity vs. oracle across comprehensive test matrix; miter spikes properly controlled; round arcs stable with configurable tolerance; all join/end type combinations working.

---

## M5 — Completeness Features & API Polish 🎯
**Goal: Complete feature set and prepare for production use**

### Tasks
- [ ] Implement missing utility functions:
  - [ ] `PointInPolygon` with all fill rule support
  - [ ] `Bounds/BoundingRect` calculation
  - [ ] `ReversePaths` for batch operations
  - [ ] `Simplify/CleanPolygon` (Douglas-Peucker or Clipper2-style)
- [ ] Add advanced operations:
  - [ ] `MinkowskiSum` and `MinkowskiDiff` operations
  - [ ] `PolyTree`/`PolyPath` hierarchy support if needed
- [ ] API improvements:
  - [ ] Error handling consistency
  - [ ] Input validation and sanitization
  - [ ] Memory-efficient path operations
- [ ] Document any intentional deviations from Clipper2 naming/behavior
- [ ] Create migration notes from other polygon libraries

**Done When:** Feature checklist matches Clipper2 documentation; examples mirror upstream behavior; API is clean and Go-idiomatic; migration guide available.

---

## M6 — Performance & Production Readiness 🚀
**Goal: Optimize performance and prepare for public release**

### Tasks
- [ ] Comprehensive benchmarking suite:
  - [ ] Boolean operations across varied vertex counts
  - [ ] Offsetting with different complexities
  - [ ] Nested holes and complex polygon handling
  - [ ] Memory allocation profiling
- [ ] Performance optimizations:
  - [ ] Memory tuning (pre-allocation, slice reuse)
  - [ ] Optional concurrency for large batch operations
  - [ ] Critical path optimization based on profiling
- [ ] Production readiness:
  - [ ] Complete API documentation with examples
  - [ ] "compat tests" folder reproducing subset of upstream test corpus
  - [ ] CI/CD pipeline with multiple Go versions and platforms
  - [ ] Performance regression testing
- [ ] Documentation and examples:
  - [ ] Usage guides and tutorials
  - [ ] Performance comparison documentation
  - [ ] Best practices guide

**Done When:** Pure Go performance within acceptable factor of C++ oracle for common workloads; comprehensive documentation ready; CI pipeline robust; ready for tagged release.

---

## Success Criteria

### Functional Requirements
- ✅ **Feature Parity**: All core Clipper2 operations implemented and validated
- ✅ **Result Accuracy**: Pure Go matches oracle within acceptable numerical tolerance
- ✅ **Fill Rule Support**: All four fill rules working correctly
- ✅ **Join/End Types**: Complete offsetting support

### Quality Requirements  
- ✅ **Validation Coverage**: Property-based tests comparing pure vs. oracle
- ✅ **Fuzz Testing**: Edge cases and numerical stability validated
- ✅ **Performance**: Within reasonable factor (2-3x) of C++ for typical workloads
- ✅ **Reliability**: Handles degenerate inputs gracefully

### Production Requirements
- ✅ **API Stability**: Clean, Go-idiomatic interface
- ✅ **Documentation**: Complete with examples and migration guides
- ✅ **CI/CD**: Multi-platform testing and performance monitoring
- ✅ **Release Ready**: Tagged releases with semantic versioning

---

## Implementation Strategy

### Development Approach
1. **Milestone-Driven**: Complete each milestone fully before proceeding
2. **Oracle Validation**: Continuously compare against CGO reference
3. **Test-First**: Write property-based tests before implementation
4. **Performance Aware**: Profile and optimize at each milestone

### Build System
- **Default Mode**: Pure Go implementation (`port/impl_pure.go`)
- **Oracle Mode**: CGO validation (`port/impl_oracle_cgo.go` with `-tags=clipper_cgo`)
- **Continuous Integration**: Both modes tested on every commit

### Testing Philosophy
- **Property-Based**: Random inputs validated against oracle
- **Fuzz Testing**: Edge cases and numerical stability
- **Acceptance Criteria**: 99%+ match rate with documented exceptions
- **Performance**: Benchmarked against reference implementation

### Numerical Precision
- **128-bit Intermediate Math**: Prevent overflow in calculations
- **Acceptable Tolerances**: Document precision limitations
- **Edge Case Handling**: Graceful degradation for degenerate inputs

---

**Estimated Timeline**: 6-8 months (assuming part-time development)
**Current Milestone**: M1 Complete ✅ - Ready for M2 (Core Geometry Kernel & Robust Integer Math)