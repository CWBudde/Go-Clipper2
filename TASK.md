# Go Clipper2 Implementation Roadmap

A milestone-based implementation plan for porting Clipper2 to pure Go with validation against the CGO oracle.

---

## 🎯 Project Goal

Create a **production-ready pure Go port** of the Clipper2 polygon clipping library with:

- Zero C/C++ dependencies for production deployments
- Identical results to the original C++ implementation
- Clean, idiomatic Go API
- Comprehensive test coverage with property-based testing

---

## 📊 Current Status Summary

### What's Working ✅

**CGO Oracle (Development & Validation Tool)**

- ✅ **100% Functional** - All 11/11 tests passing
- ✅ Complete C bridge to vendored Clipper2 C++ source
- ✅ All boolean operations: Union, Intersection, Difference, XOR
- ✅ Path offsetting: InflatePaths64 with all join/end types
- ✅ Rectangle clipping: RectClip64
- ✅ Proper memory management and error handling
- ✅ **Production-ready** if pure Go is not required

**Pure Go Implementation - Foundations**

- ✅ Core utilities: Area64, IsPositive64, Reverse64
- ✅ Rectangle clipping: RectClip64 (fully tested, fuzz validated)
- ✅ Robust 128-bit integer math (CrossProduct128, Area128, DistanceSquared128)
- ✅ Geometry kernel: segment intersection, winding numbers, point-in-polygon
- ✅ All fill rules implemented: EvenOdd, NonZero, Positive, Negative

### What's Complete ✅

**Pure Go Implementation - Boolean Operations**

- ✅ **Vatti scanline algorithm fully working** (~800 lines in `vatti_engine.go`)
- ✅ **100% match with C++ oracle on all test cases**:
  - Union: Correctly merges overlapping polygons
  - Intersection: Returns exact intersection regions
  - Difference: Properly subtracts clip from subject
  - XOR: Correctly computes exclusive-or regions
- ✅ All 4 fill rules: EvenOdd, NonZero, Positive, Negative
- ✅ All edge cases handled: nested, separated, adjacent, L-shaped polygons

### What's Not Started ❌

- ❌ Pure Go polygon offsetting (InflatePaths64)
- ❌ Advanced operations (Minkowski sum/diff)
- ❌ Performance optimization
- ❌ Production documentation and examples

### Overall Progress: **~75% Complete**

- M0: Foundation ✅ **DONE**
- M1: Rectangle Clipping + CGO Oracle ✅ **DONE**
- M2: Geometry Kernel ✅ **DONE**
- M3: Boolean Operations ✅ **DONE** (2025-10-21)
- M4: Offsetting ❌ **Not Started**
- M5: Completeness ❌ **Not Started**
- M6: Production Polish ❌ **Not Started**

---

## Milestone Details

## M0 — Foundation ✅ **COMPLETE**

- [x] Module path and imports configured
- [x] Public API surface defined in `port/`
- [x] Build tag system working:
  - `port/impl_pure.go` (pure Go implementations)
  - `port/impl_oracle_cgo.go` (CGO delegations with `-tags=clipper_cgo`)
  - `capi/` (CGO bindings, all files require `clipper_cgo` tag)
- [x] CI pipeline for Linux/macOS testing both modes
- [x] Development commands via `justfile`

---

## M1 — Rectangle Clipping + CGO Oracle ✅ **COMPLETE**

### Rectangle Clipping (Pure Go)

- [x] Sutherland-Hodgman clipping algorithm implemented
- [x] Handles closed and open paths
- [x] Edge case handling (degenerate rectangles, boundaries, etc.)
- [x] Property-based tests comparing pure Go vs. oracle
- [x] Fuzz testing achieving ≥99% match rate
- [x] All tests passing

### CGO Oracle Infrastructure ✅

- [x] C++ bridge library (`capi/clipper_bridge.cc` and `clipper_bridge.h`)
- [x] Vendored Clipper2 source code in `third_party/clipper2/`
- [x] Core wrapper functions:
  - `clipper2c_boolean64` - All boolean operations
  - `clipper2c_offset64` - Path offsetting
  - `clipper2c_rectclip64` - Rectangle clipping
  - `clipper2c_free_paths` - Memory management
- [x] Go↔C data conversion (pack/unpack functions)
- [x] All CGO tests passing (11/11)
- [x] **Oracle is production-ready and fully functional**

---

## M2 — Core Geometry Kernel ✅ **COMPLETE**

- [x] Robust 128-bit integer math:
  - [x] Cross products without overflow
  - [x] Area accumulations
  - [x] Distance calculations
- [x] Robust segment intersection:
  - [x] Collinear overlap handling
  - [x] Parallel and near-parallel segments
  - [x] Endpoint and crossing detection
- [x] Winding number and point-in-polygon tests
- [x] All fill rules: EvenOdd, NonZero, Positive, Negative
- [x] Comprehensive test suite:
  - [x] Parallel/collinear/degenerate cases
  - [x] Numerical edge cases near overflow
  - [x] All fill rule behaviors
  - [x] 900+ lines of geometry tests

**Status:** Kernel is production-ready and fully tested.

---

## M3 — Pure Go Boolean Operations ✅ **COMPLETE** (2025-10-21)

- [x] Vatti scanline algorithm structure (~800 lines in `vatti_engine.go`)
- [x] Event queue and local minima detection (`vertex.go`)
- [x] Active edge list (AEL) management with proper ordering
- [x] **Critical fix: C++ merge-sort intersection detection**
  - [x] `BuildIntersectList` - finds ALL edge crossings (not just adjacent)
  - [x] `ProcessIntersectList` - processes intersections bottom-up
  - [x] `AddNewIntersectNode` - precise intersection point calculation
- [x] Output polygon builder (circular linked lists)
- [x] All 4 operations: Union, Intersection, Difference, XOR
- [x] All 4 fill rules: EvenOdd, NonZero, Positive, Negative
- [x] Horizontal edge processing subsystem (~400 lines)
  - [x] `DoHorizontal` - handles horizontal segments
  - [x] Queue management with `PushHorz`/`PopHorz`
  - [x] Edge advancement and direction tracking
- [x] Comprehensive test suite (32 tests, 100% oracle match)
  - [x] 8 test cases × 4 operations
  - [x] Nested, separated, adjacent, L-shaped polygons
  - [x] Point-by-point validation against oracle

**Status:** All boolean operations production-ready and fully tested.

---

## M4 — Pure Go Offsetting ❌ **NOT STARTED**

**Goal: Complete polygon offsetting with all join and end types**

**Strategy: Incremental implementation by join/end type complexity for risk mitigation**

Reference: `third_party/clipper2/CPP/Clipper2Lib/src/clipper.offset.cpp` (~662 lines)

### Phase 1: Infrastructure + Bevel Joins + Closed Polygons

**Goal:** Core infrastructure and simplest join type working

**Tasks:**

- [ ] Create `port/offset.go` with ClipperOffset type
- [ ] Create `port/offset_internal.go` for helper functions
- [ ] Add JoinType and EndType enums to `port/types.go`
- [ ] Implement core infrastructure:
  - [ ] offsetGroup type
  - [ ] AddPath/AddPaths methods
  - [ ] BuildNormals - calculate perpendicular unit vectors
  - [ ] Helper functions: GetUnitNormal, Hypot, GetPerpendic
  - [ ] GetLowestClosedPathInfo - orientation detection
- [ ] Implement DoBevel (simplest join - just 2 offset points)
- [ ] Implement OffsetPoint orchestrator (Phase 1: only Bevel support)
- [ ] Implement OffsetPolygon (closed path offsetting)
- [ ] Implement ExecuteInternal with Union cleanup
- [ ] Implement public Execute method
- [ ] Add oracle tests: simple polygon expansion/contraction with Bevel joins
- [ ] Validate against oracle (target: 100% match)

**Validation Focus:**

- Simple square/rectangle offsetting
- Positive deltas (expansion)
- Negative deltas (contraction)
- Concave and convex polygons

**Done When:**

- Bevel joins working perfectly on closed polygons
- Oracle validation passing at 100%
- ~150-200 lines implemented

### Phase 2: Add Miter Joins

**Goal:** Add miter joins with miter limit control

**Tasks:**

- [ ] Add miter_limit field to ClipperOffset
- [ ] Add temp_lim calculation in ExecuteInternal
- [ ] Implement DoMiter:
  - [ ] Calculate miter point from averaged normals
  - [ ] Apply group_delta scaling
- [ ] Update OffsetPoint to handle JoinType::Miter
- [ ] Add miter limit fallback logic (falls back to Square when exceeded)
- [ ] Add MiterLimit accessor methods
- [ ] Add oracle tests: sharp corners with various miter limits
- [ ] Validate against oracle

**Validation Focus:**

- Acute angles (< 45°)
- Sharp spikes with miter limits
- Miter limit exceeded → fallback behavior
- Star polygons and other sharp-cornered shapes

**Done When:**

- Miter joins working correctly
- MiterLimit parameter controlling spike length
- Oracle validation passing at 100%
- ~50-80 lines added

### Phase 3: Add Square Joins

**Goal:** Add square joins with intersection calculations

**Tasks:**

- [ ] Implement DoSquare:
  - [ ] Calculate average unit vector
  - [ ] Offset original vertex along unit vector
  - [ ] Calculate perpendicular vertices
  - [ ] Find segment intersection points
  - [ ] Handle reflection for symmetry
- [ ] Add GetAvgUnitVector helper
- [ ] Add segment intersection helper (may already exist in geometry.go)
- [ ] Update OffsetPoint to handle JoinType::Square
- [ ] Add oracle tests: square corner behaviors
- [ ] Validate against oracle

**Validation Focus:**

- 90-degree corners
- Various angles
- Edge cases where square joins create long extensions

**Done When:**

- Square joins working correctly
- Oracle validation passing at 100%
- ~80-100 lines added

### Phase 4: Add Round Joins

**Goal:** Add round joins with arc approximation and arc tolerance

**Tasks:**

- [ ] Add arc_tolerance field to ClipperOffset
- [ ] Add steps_per_rad, step_sin, step_cos fields
- [ ] Implement DoRound:
  - [ ] Calculate steps needed for arc approximation
  - [ ] Generate arc points using rotation matrix
  - [ ] Handle dynamic arc tolerance calculation
- [ ] Add arc calculation logic to ExecuteInternal
- [ ] Add ArcTolerance accessor methods
- [ ] Update OffsetPoint to handle JoinType::Round
- [ ] Add oracle tests: smooth curves with various arc tolerances
- [ ] Validate against oracle

**Validation Focus:**

- Smooth rounded corners
- Arc tolerance effects (coarse vs fine approximation)
- Large vs small offset deltas
- Circles and ellipses

**Done When:**

- Round joins working correctly
- ArcTolerance parameter controlling curve quality
- Oracle validation passing at 100%
- ~80-100 lines added

### Phase 5: Add Open Path Support

**Goal:** Support open paths with all end cap types

**Tasks:**

- [ ] Implement OffsetOpenPath:
  - [ ] Start cap handling (Butt, Square, Round)
  - [ ] Forward edge offsetting
  - [ ] Reverse normals for return path
  - [ ] End cap handling
  - [ ] Backward edge offsetting
- [ ] Implement OffsetOpenJoined:
  - [ ] Offset as polygon
  - [ ] Reverse path
  - [ ] Rebuild and negate normals
  - [ ] Offset reversed path
- [ ] Update DoGroupOffset to route to correct handler based on EndType
- [ ] Handle single-point paths (circle/square generation)
- [ ] Handle two-point paths with Joined end type
- [ ] Add oracle tests: line offsetting with all end cap combinations
- [ ] Validate against oracle

**Validation Focus:**

- Open line segments
- All end cap types: Butt, Square, Round
- Joined open paths
- Single-point and two-point edge cases

**Done When:**

- All end types working correctly
- Open path offsetting matches oracle
- Single/two-point edge cases handled
- Oracle validation passing at 100%
- ~120-150 lines added

### Phase 6: Polish and Edge Cases

**Goal:** Handle remaining edge cases and optimize

**Tasks:**

- [ ] Add preserve_collinear support
- [ ] Add reverse_solution support
- [ ] Handle degenerate inputs gracefully
- [ ] Add comprehensive test suite:
  - [ ] All join type × end type combinations
  - [ ] Self-intersecting inputs
  - [ ] Very large and very small deltas
  - [ ] Edge cases from upstream test suite
- [ ] Performance profiling and optimization
- [ ] Documentation and examples

**Done When:**

- Feature parity with C++ Clipper2 offset implementation
- All oracle tests passing at 100%
- Edge cases handled robustly
- Ready for production use

### Overall M4 Completion Criteria

- ✅ All 4 join types working: Bevel, Miter, Square, Round
- ✅ All 5 end types working: Polygon, Joined, Butt, Square, Round
- ✅ Precision controls: MiterLimit, ArcTolerance
- ✅ 100% parity with oracle across comprehensive test matrix
- ✅ ~450-600 lines of well-tested Go code

---

## M5 — Completeness Features ❌ **NOT STARTED**

**Goal: Complete feature set and production API**

### Tasks

- [ ] Implement missing utility functions:
  - [ ] `PointInPolygon` (currently in geometry.go, need to expose as public API)
  - [ ] `Bounds64`/`BoundingRect` calculation
  - [ ] `ReversePaths64` for batch operations
  - [ ] `SimplifyPath64`/`CleanPolygon64`
- [ ] Add advanced operations:
  - [ ] `MinkowskiSum64` and `MinkowskiDiff64`
  - [ ] `PolyTree`/`PolyPath` hierarchy if needed
- [ ] API polish:
  - [ ] Consistent error handling across all functions
  - [ ] Input validation and sanitization
  - [ ] Memory-efficient path operations
- [ ] Documentation:
  - [ ] Document deviations from C++ Clipper2 (if any)
  - [ ] Migration guide from other polygon libraries
  - [ ] Best practices guide

**Done When:**

- Feature parity with Clipper2 C++ library
- API is clean and Go-idiomatic
- Examples mirror upstream behavior
- Migration documentation complete

---

## M6 — Performance & Production Readiness ❌ **NOT STARTED**

**Goal: Optimize and prepare for public release**

### Tasks

- [ ] Comprehensive benchmarking:
  - [ ] Boolean operations across varied vertex counts
  - [ ] Offsetting with different complexities
  - [ ] Nested holes and complex polygons
  - [ ] Memory allocation profiling
- [ ] Performance optimizations:
  - [ ] Memory tuning (pre-allocation, slice reuse)
  - [ ] Optional concurrency for batch operations
  - [ ] Critical path optimization based on profiling
- [ ] Production readiness:
  - [ ] Complete API documentation with examples
  - [ ] Compat test suite reproducing upstream test cases
  - [ ] Multi-platform CI/CD (Linux, macOS, Windows)
  - [ ] Performance regression testing
- [ ] Documentation:
  - [ ] Usage guides and tutorials
  - [ ] Performance comparison with C++ version
  - [ ] Architecture documentation

**Done When:**

- Pure Go performance within 2-3x of C++ oracle
- Comprehensive documentation ready
- CI pipeline robust across platforms
- Ready for v1.0.0 tagged release

---

## Development Workflow

### Build and Test Commands

```bash
# Pure Go mode (default)
just build          # Build all packages
just test           # Run all tests
just test-port      # Test only port package
just dev            # Format + lint + test

# CGO Oracle mode (validation)
just build-oracle   # Build with CGO
just test-oracle    # Test with oracle validation
just test-capi      # Test only CGO bindings
just dev-oracle     # Full dev workflow with oracle

# Specific tests
just test-run TestIntersect64Basic           # Pure Go
just test-run-oracle TestIntersect64Basic    # With oracle

# Code quality
just fmt            # Format code
just lint           # Run linters
just lint-fix       # Auto-fix lint issues
```

### Testing Philosophy

1. **Test with Oracle First**
   - Write test case
   - Verify it passes with `-tags=clipper_cgo`
   - Oracle provides ground truth

2. **Implement Pure Go**
   - Implement algorithm in `port/impl_pure.go`
   - Test should pass without `-tags=clipper_cgo`
   - Results should match oracle exactly

3. **Property-Based Testing**
   - Random inputs validated against oracle
   - Fuzz testing for edge cases
   - Target: ≥99% match rate

4. **Debug Discrepancies**
   - Use oracle output as expected result
   - Add detailed logging to pure Go implementation
   - Compare step-by-step with reference C++ code

---

## Implementation Strategy

**M4: Offsetting (4-6 weeks)**

- Study reference implementation in `clipper.offset.cpp`
- Start with simple expansion/contraction
- Add join types incrementally
- Validate each step against oracle

**M5: Completeness (2-3 weeks)**

- Straightforward utility functions
- API cleanup and documentation
- No complex algorithms

**M6: Production Polish (3-4 weeks)**

- Performance profiling and optimization
- Comprehensive documentation
- CI/CD hardening
- Release preparation

**Estimated Timeline:** 3-5 months to v1.0.0 (assuming part-time development)

---

## Success Criteria

### Functional Requirements ✅/🔧/❌

- ✅ **CGO Oracle**: Complete and production-ready
- 🔧 **Boolean Operations**: Implemented but debugging needed
- ✅ **Fill Rules**: All four working correctly
- ❌ **Offsetting**: Not yet implemented
- ✅ **Rectangle Clipping**: Complete and tested
- ✅ **Geometry Kernel**: Production-ready

### Quality Requirements

- ✅ **Oracle Validation**: Infrastructure complete
- ✅ **Fuzz Testing**: Framework in place (used for RectClip)
- 🔧 **Result Accuracy**: Needs debugging to match oracle
- ❌ **Performance**: Not yet optimized

### Production Requirements

- 🔧 **API Stability**: Mostly stable, needs polish in M5
- 🔧 **Documentation**: Basic docs exist, needs expansion
- ✅ **CI/CD**: Working for both build modes
- ❌ **Release**: Not ready (M3 must complete first)

---

## Notes for Developers

### CGO Oracle is Production-Ready ✅

If you need Clipper2 functionality in Go **right now** and don't mind the C++ dependency:

- Use `-tags=clipper_cgo` when building
- All operations work correctly (11/11 tests passing)
- Performance is excellent (native C++ speed)
- Fully tested and validated

The pure Go implementation is still in development primarily for:

- Zero-dependency deployments
- Cross-compilation simplicity
- WebAssembly support
- Educational/research purposes

### Debugging Pure Go Boolean Operations

The Vatti algorithm implementation exists but has bugs. To help debug:

1. **Enable logging** in `vatti_engine.go` (add print statements)
2. **Start simple** - use 2 overlapping rectangles as test case
3. **Compare with reference** - `third_party/clipper2/CPP/Clipper2Lib/src/clipper.engine.cpp`
4. **Check intermediate steps**:
   - Are local minima detected correctly?
   - Is the active edge list managed properly?
   - Are intersections calculated correctly?
   - Are output points generated correctly?
   - Is polygon linking working?

5. **Oracle is always right** - if pure Go differs from oracle, pure Go is wrong

### Code Organization

```plain
port/
├── clipper.go              # Public API
├── types.go                # Type definitions
├── errors.go               # Error constants
├── impl_pure.go            # Pure Go implementations (entry points)
├── impl_oracle_cgo.go      # CGO delegations (build tag: clipper_cgo)
├── vatti_engine.go         # Vatti scanline algorithm (🔧 DEBUGGING)
├── vertex.go               # Vertex chain and local minima
├── boolean_simple.go       # Legacy simple implementations (unused)
├── rectangle_clipping.go   # RectClip64 (✅ complete)
├── geometry.go             # Geometry utilities (✅ complete)
├── math128.go              # 128-bit integer math (✅ complete)
└── *_test.go               # Tests

capi/
├── clipper_cgo.go          # CGO bindings (✅ complete)
├── clipper_cgo_test.go     # CGO tests (✅ all passing)
└── clipper_bridge.*        # C++ wrapper (✅ complete)
```

### File Sizes (Lines of Code)

- `vatti_engine.go`: 588 lines (core algorithm)
- `vertex.go`: 214 lines (vertex chain handling)
- `geometry.go`: 272 lines (utility functions)
- `math128.go`: 235 lines (robust arithmetic)
- Total implementation: ~5,200 lines of Go code

---

**Last Updated:** 2025-10-21
**Current Milestone:** M3 (Boolean Operations - Debugging)
**Next Release Target:** v0.3.0 (when M3 complete)
