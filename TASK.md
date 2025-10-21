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

### What's In Progress 🔧

**Pure Go Implementation - Boolean Operations**
- 🔧 **Vatti scanline algorithm implemented** (~600 lines in `vatti_engine.go`)
- 🔧 **Tests pass but produce WRONG results**:
  - Union: Returns separate polygons instead of merged shape
  - Intersection: Returns malformed 4-point polygon
  - Difference/XOR: Incorrect output
- 🔧 Algorithm structure exists, debugging needed to match oracle behavior

### What's Not Started ❌

- ❌ Pure Go polygon offsetting (InflatePaths64)
- ❌ Advanced operations (Minkowski sum/diff)
- ❌ Performance optimization
- ❌ Production documentation and examples

### Overall Progress: **~70% Complete**

- M0: Foundation ✅ **DONE**
- M1: Rectangle Clipping + CGO Oracle ✅ **DONE**
- M2: Geometry Kernel ✅ **DONE**
- M3: Boolean Operations 🔧 **70% - Debugging Needed**
- M4: Offsetting ❌ **Not Started**
- M5: Completeness ❌ **Not Started**
- M6: Production Polish ❌ **Not Started**

---

## Milestone Details

## M0 — Foundation ✅ **COMPLETE**

**Goal: Project structure, build system, CI/CD**

### Completed Tasks
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

**Goal: First pure Go algorithm + working CGO validation infrastructure**

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

**Goal: Bulletproof mathematical foundation**

### Completed Tasks
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

## M3 — Pure Go Boolean Operations 🔧 **70% COMPLETE - DEBUGGING**

**Goal: Implement and debug Vatti scanline algorithm for all boolean operations**

### Algorithm Implementation Status

**✅ Implemented (but buggy):**
- [x] Vatti scanline algorithm structure (~600 lines)
- [x] Event queue (local minima detection) - `vertex.go`
- [x] Active edge list management - `vatti_engine.go`
- [x] Intersection detection and processing
- [x] Output polygon builder
- [x] All fill rules: EvenOdd, NonZero, Positive, Negative
- [x] ClipType handling: Union, Intersection, Difference, XOR

**🔧 Debugging Needed:**

The pure Go implementation has all the algorithm structure but produces incorrect results. Tests pass without errors but output doesn't match oracle.

### Current Test Results

**Pure Go vs. Oracle Comparison:**

| Operation      | Pure Go Result                    | Oracle Result                     | Status   |
|----------------|-----------------------------------|-----------------------------------|----------|
| Union          | 2 separate polygons               | 1 merged polygon                  | ❌ Wrong |
| Intersection   | Malformed 4-point polygon         | Correct intersection square       | ❌ Wrong |
| Difference     | Returns subject unchanged         | Correct difference                | ❌ Wrong |
| XOR            | Returns all input polygons        | Correct symmetric difference      | ❌ Wrong |

### Debugging Tasks

**Priority 1: Fix Core Algorithm Bugs**
- [ ] Debug Union operation:
  - [ ] Trace why polygons aren't being merged
  - [ ] Check output polygon linking logic
  - [ ] Verify winding count calculations for union
- [ ] Debug Intersection operation:
  - [ ] Investigate malformed output polygon
  - [ ] Check intersection point calculation
  - [ ] Verify edge list management during intersection
- [ ] Debug Difference operation:
  - [ ] Check why clip paths aren't subtracting
  - [ ] Verify clip type handling in Vatti engine
- [ ] Debug XOR operation:
  - [ ] Check symmetric difference logic
  - [ ] Verify proper region extraction

**Priority 2: Systematic Testing**
- [ ] Start with simplest cases (disjoint rectangles)
- [ ] Add overlapping rectangles
- [ ] Test adjacent rectangles (shared edges)
- [ ] Complex polygons with holes
- [ ] Self-intersecting polygons
- [ ] Open paths (lines)

**Priority 3: Oracle Validation**
- [ ] Property-based testing comparing pure Go vs. oracle
- [ ] Fuzz testing for edge cases
- [ ] Achieve ≥99% match rate across random inputs

### Debugging Strategy

1. **Add detailed logging** to Vatti engine to trace:
   - Local minima detection
   - Edge insertion/removal from active list
   - Intersection calculations
   - Output point generation
   - Polygon linking

2. **Compare step-by-step** with reference implementation in `third_party/clipper2/`

3. **Test incrementally**:
   - Fix one operation at a time (start with Intersection)
   - Add test case, debug until it passes
   - Move to next operation

4. **Use oracle as ground truth**:
   - Every test case should pass with `-tags=clipper_cgo`
   - Pure Go should match oracle output exactly

**Done When:**
- All boolean operations produce identical results to oracle
- Tests pass in both pure Go and CGO modes
- ≥99% match rate on property-based/fuzz tests
- Complex polygons (holes, self-intersecting) handled correctly

---

## M4 — Pure Go Offsetting ❌ **NOT STARTED**

**Goal: Complete polygon offsetting with all join and end types**

### Tasks
- [ ] Implement core offsetting algorithm (expansion/contraction)
- [ ] Add all join types:
  - [ ] Round joins with arc approximation
  - [ ] Miter joins with miter limit
  - [ ] Square joins for sharp corners
- [ ] Support all end types:
  - [ ] ClosedPolygon and ClosedLine
  - [ ] OpenSquare, OpenRound, OpenButt end caps
- [ ] Implement precision controls:
  - [ ] MiterLimit for spike length control
  - [ ] ArcTolerance for round approximation quality
- [ ] Validate against oracle:
  - [ ] Various deltas (positive and negative)
  - [ ] All join/end type combinations
  - [ ] Self-intersecting inputs

**Prerequisites:**
- M3 must be complete (debugging fixed)
- Reference implementation in `third_party/clipper2/CPP/Clipper2Lib/src/clipper.offset.cpp`

**Done When:**
- Parity with oracle across comprehensive test matrix
- All join/end types working correctly
- Miter/arc controls functioning properly

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

### Current Focus: M3 Debugging (Next 2-4 weeks)

**Immediate priorities:**
1. Fix Intersection operation first (simplest case)
2. Add comprehensive logging to Vatti engine
3. Test with minimal test cases (2 overlapping rectangles)
4. Compare with reference implementation step-by-step
5. Once Intersection works, apply learnings to Union/Difference/XOR

### After M3 Complete:

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

```
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
