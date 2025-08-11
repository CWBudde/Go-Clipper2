package clipper

import (
	"math"
	"sort"
	"testing"
	"unsafe"
)

// FuzzRectClip64 performs comprehensive fuzz testing of RectClip64 by comparing
// pure Go implementation results against the CGO oracle implementation.
// This test aims to achieve ≥99% match rate between implementations as required for M1.
//
// Usage:
//
//	Pure Go mode (self-consistency):  go test -fuzz=FuzzRectClip64 -fuzztime=30s
//	Oracle mode (full validation):    go test -tags=clipper_cgo -fuzz=FuzzRectClip64 -fuzztime=30s
//
// The test generates random rectangles and paths, clips them using both implementations,
// and compares results. It tracks match rates and provides detailed failure analysis.
func FuzzRectClip64(f *testing.F) {
	// Seed corpus with representative test cases
	seedTestCases := []struct {
		name  string
		rect  [4][2]int64 // Rectangle as [4][2]int64 for fuzz compatibility
		paths [][]Point64 // Paths for fuzzing
	}{
		{
			"BasicOverlap",
			[4][2]int64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			[][]Point64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}},
		},
		{
			"CompletelyInside",
			[4][2]int64{{0, 0}, {20, 0}, {20, 20}, {0, 20}},
			[][]Point64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}},
		},
		{
			"CompletelyOutside",
			[4][2]int64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			[][]Point64{{{20, 20}, {30, 20}, {30, 30}, {20, 30}}},
		},
		{
			"OnBoundary",
			[4][2]int64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			[][]Point64{{{0, 5}, {5, 0}, {10, 5}, {5, 10}}},
		},
		{
			"NegativeCoords",
			[4][2]int64{{-10, -10}, {10, -10}, {10, 10}, {-10, 10}},
			[][]Point64{{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}},
		},
		{
			"EmptyPaths",
			[4][2]int64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			[][]Point64{},
		},
		{
			"SinglePoint",
			[4][2]int64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			[][]Point64{{{5, 5}}},
		},
		{
			"DegenerateRect",
			[4][2]int64{{5, 5}, {5, 5}, {5, 10}, {5, 10}}, // Zero width
			[][]Point64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
		},
	}

	// Add seed corpus
	for _, tc := range seedTestCases {
		// Flatten paths into a format suitable for fuzzing
		var flatPaths []byte

		// Encode number of paths
		numPaths := len(tc.paths)
		flatPaths = append(flatPaths, byte(numPaths))

		for _, path := range tc.paths {
			// Encode number of points in path
			numPoints := len(path)
			flatPaths = append(flatPaths, byte(numPoints))

			// Encode each point (16 bytes: 8 for X, 8 for Y)
			for _, pt := range path {
				xBytes := (*[8]byte)(unsafe.Pointer(&pt.X))[:]
				yBytes := (*[8]byte)(unsafe.Pointer(&pt.Y))[:]
				flatPaths = append(flatPaths, xBytes...)
				flatPaths = append(flatPaths, yBytes...)
			}
		}

		// Add to fuzz corpus
		rectBytes := (*[32]byte)(unsafe.Pointer(&tc.rect))[:]
		f.Add(rectBytes[:], flatPaths)
	}

	var totalTests int
	var matches int
	var mismatches int
	var pureErrors int
	var oracleErrors int

	f.Fuzz(func(t *testing.T, rectData []byte, pathsData []byte) {
		totalTests++

		// Skip if input data is too small
		if len(rectData) < 32 { // 4 points × 2 coords × 8 bytes
			return
		}
		if len(pathsData) < 1 {
			return
		}

		// Parse rectangle from bytes
		rect := parseRectFromBytes(rectData)

		// Parse paths from bytes
		paths := parsePathsFromBytes(pathsData)

		// Skip extreme cases that would cause overflow
		if !isValidInput(rect, paths) {
			return
		}

		// Test pure Go implementation
		pureResult, pureErr := RectClip64(rect, paths)

		// Test CGO oracle implementation (only when CGO build tag is active)
		oracleResult, oracleErr := rectClipOracle(rect, paths)

		// Handle error cases
		if pureErr != nil && oracleErr != nil {
			// Both failed - this is acceptable for some invalid inputs
			if pureErr.Error() == oracleErr.Error() {
				matches++
			} else {
				mismatches++
				t.Logf("Error mismatch - Pure: %v, Oracle: %v", pureErr, oracleErr)
			}
			return
		}

		if pureErr != nil {
			pureErrors++
			t.Logf("Pure implementation error: %v (Input: rect=%v, paths=%v)", pureErr, rect, paths)
			return
		}

		if oracleErr != nil {
			oracleErrors++
			t.Logf("Oracle implementation error: %v (Input: rect=%v, paths=%v)", oracleErr, rect, paths)
			return
		}

		// Compare results
		if pathsetsEqual(pureResult, oracleResult) {
			matches++
		} else {
			mismatches++

			// Log detailed mismatch for analysis (but only first few to avoid spam)
			if mismatches <= 10 {
				t.Logf("MISMATCH #%d:", mismatches)
				t.Logf("  Input rect: %v", rect)
				t.Logf("  Input paths: %v", paths)
				t.Logf("  Pure result (%d paths): %v", len(pureResult), pureResult)
				t.Logf("  Oracle result (%d paths): %v", len(oracleResult), oracleResult)
			}

			// Only fail the test if mismatch rate is too high AND we have real oracle (CGO mode)
			if totalTests > 100 && isRealOracleMode() {
				matchRate := float64(matches) / float64(totalTests) * 100
				if matchRate < 99.0 {
					t.Errorf("Match rate too low: %.2f%% (%d matches / %d total)", matchRate, matches, totalTests)
				}
			}
		}

		// Report progress periodically
		if totalTests%1000 == 0 {
			matchRate := float64(matches) / float64(totalTests) * 100
			mode := "pure-go-consistency"
			if isRealOracleMode() {
				mode = "pure-vs-cgo-oracle"
			}
			t.Logf("Progress: %d tests (%s), %.2f%% match rate, %d mismatches, %d pure errors, %d oracle errors",
				totalTests, mode, matchRate, mismatches, pureErrors, oracleErrors)
		}

		// Final report
		if totalTests > 0 && (totalTests%10000 == 0 || totalTests == 1) {
			matchRate := float64(matches) / float64(totalTests) * 100
			mode := "Pure Go consistency check"
			if isRealOracleMode() {
				mode = "Pure Go vs CGO oracle validation"
			}
			t.Logf("=== FUZZ TEST SUMMARY ===")
			t.Logf("Mode: %s", mode)
			t.Logf("Total tests: %d", totalTests)
			t.Logf("Match rate: %.2f%%", matchRate)
			t.Logf("Matches: %d, Mismatches: %d", matches, mismatches)
			t.Logf("Pure errors: %d, Oracle errors: %d", pureErrors, oracleErrors)
			if isRealOracleMode() {
				if matchRate >= 99.0 {
					t.Logf("✓ M1 SUCCESS: Match rate ≥99%% achieved!")
				} else {
					t.Logf("✗ M1 FAILURE: Match rate <99%% (need ≥99%% for M1 completion)")
				}
			} else {
				t.Logf("ℹ Running in pure Go mode - use -tags=clipper_cgo for oracle validation")
			}
		}
	})
}

// isRealOracleMode returns true if we're running with CGO oracle, false if pure Go mode
func isRealOracleMode() bool {
	return isRealOracleModeImpl()
}

// parseRectFromBytes converts byte data to a valid rectangle Path64
func parseRectFromBytes(data []byte) Path64 {
	// Extract 4 points from byte data
	rect := make(Path64, 4)

	for i := 0; i < 4 && (i*16+15) < len(data); i++ {
		// Extract X coordinate (8 bytes)
		x := *(*int64)(unsafe.Pointer(&data[i*16]))
		// Extract Y coordinate (8 bytes)
		y := *(*int64)(unsafe.Pointer(&data[i*16+8]))

		// Constrain to reasonable range to avoid overflow
		x = constrainCoordinate(x)
		y = constrainCoordinate(y)

		rect[i] = Point64{X: x, Y: y}
	}

	return rect
}

// parsePathsFromBytes converts byte data to Paths64
func parsePathsFromBytes(data []byte) Paths64 {
	if len(data) == 0 {
		return Paths64{}
	}

	paths := Paths64{}
	offset := 0

	// Read number of paths (limited to reasonable count)
	numPaths := int(data[0]) % 21 // Limit to 0-20 paths
	offset++

	for p := 0; p < numPaths && offset < len(data); p++ {
		// Read number of points in this path
		if offset >= len(data) {
			break
		}
		numPoints := int(data[offset]) % 51 // Limit to 0-50 points per path
		offset++

		if numPoints < 2 { // Skip paths with fewer than 2 points
			continue
		}

		path := make(Path64, 0, numPoints)

		for i := 0; i < numPoints && offset+15 < len(data); i++ {
			// Read X coordinate (8 bytes)
			x := *(*int64)(unsafe.Pointer(&data[offset]))
			offset += 8

			// Read Y coordinate (8 bytes)
			y := *(*int64)(unsafe.Pointer(&data[offset]))
			offset += 8

			// Constrain coordinates
			x = constrainCoordinate(x)
			y = constrainCoordinate(y)

			path = append(path, Point64{X: x, Y: y})
		}

		if len(path) >= 2 {
			paths = append(paths, path)
		}
	}

	return paths
}

// constrainCoordinate ensures coordinates stay within a reasonable range
func constrainCoordinate(coord int64) int64 {
	const maxCoord = 1e9
	const minCoord = -1e9

	if coord > maxCoord {
		return maxCoord
	}
	if coord < minCoord {
		return minCoord
	}
	return coord
}

// isValidInput performs basic validation to skip problematic inputs
func isValidInput(rect Path64, paths Paths64) bool {
	// Check for NaN-like values that could cause issues
	for _, pt := range rect {
		if pt.X == math.MinInt64 || pt.Y == math.MinInt64 {
			return false
		}
	}

	for _, path := range paths {
		for _, pt := range path {
			if pt.X == math.MinInt64 || pt.Y == math.MinInt64 {
				return false
			}
		}
	}

	return true
}

// pathsetsEqual compares two Paths64 results allowing for acceptable differences
func pathsetsEqual(a, b Paths64) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort both pathsets for comparison (since order may vary)
	aSorted := sortPathsForComparison(a)
	bSorted := sortPathsForComparison(b)

	for i := range aSorted {
		if !pathsEqual(aSorted[i], bSorted[i]) {
			return false
		}
	}

	return true
}

// pathsEqual compares two individual paths allowing for point reordering and minor numerical differences
func pathsEqual(a, b Path64) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	// Try to find matching starting point (paths may start at different indices)
	for offset := 0; offset < len(a); offset++ {
		if pathsEqualWithOffset(a, b, offset) {
			return true
		}
	}

	// Try reverse order as well
	bReversed := make(Path64, len(b))
	for i, j := 0, len(b)-1; i < len(b); i, j = i+1, j-1 {
		bReversed[i] = b[j]
	}

	for offset := 0; offset < len(a); offset++ {
		if pathsEqualWithOffset(a, bReversed, offset) {
			return true
		}
	}

	return false
}

// pathsEqualWithOffset compares paths with a given starting offset
func pathsEqualWithOffset(a, b Path64, offset int) bool {
	for i := 0; i < len(a); i++ {
		bIdx := (i + offset) % len(b)
		if !pointsEqual(a[i], b[bIdx]) {
			return false
		}
	}
	return true
}

// pointsEqual compares two points with small tolerance for floating-point precision differences
func pointsEqual(a, b Point64) bool {
	return a.X == b.X && a.Y == b.Y
}

// sortPathsForComparison sorts paths by their first point for consistent comparison
func sortPathsForComparison(paths Paths64) Paths64 {
	if len(paths) <= 1 {
		return paths
	}

	sorted := make(Paths64, len(paths))
	copy(sorted, paths)

	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i]) == 0 && len(sorted[j]) == 0 {
			return false
		}
		if len(sorted[i]) == 0 {
			return true
		}
		if len(sorted[j]) == 0 {
			return false
		}

		// Compare first points
		if sorted[i][0].X != sorted[j][0].X {
			return sorted[i][0].X < sorted[j][0].X
		}
		return sorted[i][0].Y < sorted[j][0].Y
	})

	return sorted
}

// rectClipOracle provides oracle implementation when CGO is available
func rectClipOracle(rect Path64, paths Paths64) (Paths64, error) {
	// This function will be conditionally compiled based on build tags
	// When CGO is not available, it will return the same as pure implementation
	// When CGO is available, it will call the CAPI oracle
	return rectClipOracleImpl(rect, paths)
}
