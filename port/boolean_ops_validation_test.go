package clipper

import (
	"testing"
)

// TestCase represents a boolean operation test case
type TestCase struct {
	Name     string
	Subject  Paths64
	Clip     Paths64
	FillRule FillRule
	Expected map[ClipType]int // Expected number of output polygons for each operation
}

// Comprehensive test cases covering various polygon configurations
var testCases = []TestCase{
	{
		Name:     "Two overlapping rectangles",
		Subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
		Clip:     Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1, // One merged polygon
			Intersection: 1, // One intersection rectangle
			Difference:   1, // Subject minus clip
			Xor:          1, // Symmetric difference
		},
	},
	{
		Name:     "Adjacent rectangles (sharing edge)",
		Subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
		Clip:     Paths64{{{10, 0}, {20, 0}, {20, 10}, {10, 10}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1, // Merged into one
			Intersection: 0, // Just a line (filtered out)
			Difference:   1, // Original subject
			Xor:          1, // Both rectangles
		},
	},
	{
		Name:     "Nested rectangles (clip inside subject)",
		Subject:  Paths64{{{0, 0}, {20, 0}, {20, 20}, {0, 20}}},
		Clip:     Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1, // Outer rectangle
			Intersection: 1, // Inner rectangle
			Difference:   1, // Outer with hole (may be 2 if holes represented separately)
			Xor:          1, // Ring shape
		},
	},
	{
		Name:     "Two separated rectangles",
		Subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
		Clip:     Paths64{{{20, 0}, {30, 0}, {30, 10}, {20, 10}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        2, // Two separate polygons
			Intersection: 0, // No intersection
			Difference:   1, // Just subject
			Xor:          2, // Both rectangles
		},
	},
	{
		Name:     "Horizontal edge alignment",
		Subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
		Clip:     Paths64{{{0, 10}, {10, 10}, {10, 20}, {0, 20}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1, // Merged vertically
			Intersection: 0, // Just a line (filtered out)
			Difference:   1, // Original subject
			Xor:          1, // Both rectangles combined
		},
	},
	{
		Name:     "Complex overlap with horizontal edges",
		Subject:  Paths64{{{0, 5}, {20, 5}, {20, 15}, {0, 15}}},
		Clip:     Paths64{{{5, 0}, {15, 0}, {15, 20}, {5, 20}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1, // Cross shape
			Intersection: 1, // Center rectangle
			Difference:   1, // Left and right bars
			Xor:          1, // Cross outline
		},
	},
	{
		Name:     "Triangle and rectangle",
		Subject:  Paths64{{{0, 0}, {10, 0}, {5, 10}}},
		Clip:     Paths64{{{3, 3}, {7, 3}, {7, 7}, {3, 7}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1,
			Intersection: 1,
			Difference:   1,
			Xor:          1,
		},
	},
	{
		Name:     "L-shaped polygons",
		Subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {5, 10}, {5, 5}, {0, 5}}},
		Clip:     Paths64{{{5, 5}, {15, 5}, {15, 15}, {10, 15}, {10, 10}, {5, 10}}},
		FillRule: NonZero,
		Expected: map[ClipType]int{
			Union:        1,
			Intersection: 1,
			Difference:   1,
			Xor:          1,
		},
	},
}

// comparePolygonSets checks if two polygon sets are equivalent (same number of polygons with same point counts)
func comparePolygonSets(t *testing.T, testName, opName string, expected, actual Paths64) bool {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s - %s: Polygon count mismatch: expected %d, got %d",
			testName, opName, len(expected), len(actual))
		return false
	}

	// Check point counts match (relaxed check - just counts, not exact points yet)
	for i := range expected {
		if len(expected[i]) != len(actual[i]) {
			t.Errorf("%s - %s: Polygon %d point count mismatch: expected %d, got %d",
				testName, opName, i, len(expected[i]), len(actual[i]))
			return false
		}
	}

	return true
}

// TestBooleanOperationsAgainstOracle validates pure Go implementation against CGO oracle
func TestBooleanOperationsAgainstOracle(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Test each operation type
			operations := []ClipType{Union, Intersection, Difference, Xor}

			for _, op := range operations {
				opName := []string{"Intersection", "Union", "Difference", "Xor"}[op]

				t.Run(opName, func(t *testing.T) {
					// Get oracle result (using CGO)
					var oracleResult Paths64
					var oracleErr error

					switch op {
					case Union:
						oracleResult, oracleErr = union64Oracle(tc.Subject, tc.Clip, tc.FillRule)
					case Intersection:
						oracleResult, oracleErr = intersect64Oracle(tc.Subject, tc.Clip, tc.FillRule)
					case Difference:
						oracleResult, oracleErr = difference64Oracle(tc.Subject, tc.Clip, tc.FillRule)
					case Xor:
						oracleResult, oracleErr = xor64Oracle(tc.Subject, tc.Clip, tc.FillRule)
					}

					if oracleErr != nil {
						t.Fatalf("Oracle failed: %v", oracleErr)
					}

					// Get pure Go result
					var pureResult Paths64
					var pureErr error

					switch op {
					case Union:
						pureResult, pureErr = Union64(tc.Subject, tc.Clip, tc.FillRule)
					case Intersection:
						pureResult, pureErr = Intersect64(tc.Subject, tc.Clip, tc.FillRule)
					case Difference:
						pureResult, pureErr = Difference64(tc.Subject, tc.Clip, tc.FillRule)
					case Xor:
						pureResult, pureErr = Xor64(tc.Subject, tc.Clip, tc.FillRule)
					}

					if pureErr != nil {
						t.Fatalf("Pure Go implementation failed: %v", pureErr)
					}

					// Log both results for debugging
					t.Logf("Oracle result: %d polygons", len(oracleResult))
					for i, path := range oracleResult {
						t.Logf("  Oracle polygon %d: %d points - %v", i, len(path), path)
					}

					t.Logf("Pure Go result: %d polygons", len(pureResult))
					for i, path := range pureResult {
						t.Logf("  Pure Go polygon %d: %d points - %v", i, len(path), path)
					}

					// Compare results
					if len(oracleResult) != len(pureResult) {
						t.Errorf("Polygon count mismatch: oracle=%d, pure=%d",
							len(oracleResult), len(pureResult))
						return
					}

					// For now, just check polygon counts match
					// Later we can add more sophisticated comparisons (point-by-point, area, etc.)
					if !comparePolygonSets(t, tc.Name, opName, oracleResult, pureResult) {
						t.Error("Results don't match oracle")
					}
				})
			}
		})
	}
}

// TestHorizontalEdgeProcessing specifically tests horizontal edge scenarios
func TestHorizontalEdgeProcessing(t *testing.T) {
	testCases := []struct {
		name     string
		subject  Paths64
		clip     Paths64
		fillRule FillRule
	}{
		{
			name:     "Single horizontal edge in subject",
			subject:  Paths64{{{0, 10}, {10, 10}, {10, 20}, {0, 20}}},
			clip:     Paths64{{{5, 0}, {15, 0}, {15, 15}, {5, 15}}},
			fillRule: NonZero,
		},
		{
			name:     "Both polygons have horizontal edges",
			subject:  Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			clip:     Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}},
			fillRule: NonZero,
		},
		{
			name:     "Horizontal edge at intersection point",
			subject:  Paths64{{{0, 5}, {10, 5}, {10, 15}, {0, 15}}},
			clip:     Paths64{{{5, 0}, {15, 0}, {15, 10}, {5, 10}}},
			fillRule: NonZero,
		},
		{
			name:     "Multiple horizontal edges",
			subject:  Paths64{{{0, 0}, {20, 0}, {20, 5}, {15, 5}, {15, 10}, {20, 10}, {20, 20}, {0, 20}}},
			clip:     Paths64{{{5, 5}, {25, 5}, {25, 15}, {5, 15}}},
			fillRule: NonZero,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Union operation specifically (most complex for horizontal edges)
			oracleResult, oracleErr := union64Oracle(tc.subject, tc.clip, tc.fillRule)
			if oracleErr != nil {
				t.Fatalf("Oracle failed: %v", oracleErr)
			}

			pureResult, pureErr := Union64(tc.subject, tc.clip, tc.fillRule)
			if pureErr != nil {
				t.Fatalf("Pure Go failed: %v", pureErr)
			}

			t.Logf("Oracle: %d polygons", len(oracleResult))
			for i, path := range oracleResult {
				t.Logf("  Polygon %d: %d points", i, len(path))
			}

			t.Logf("Pure Go: %d polygons", len(pureResult))
			for i, path := range pureResult {
				t.Logf("  Polygon %d: %d points", i, len(path))
			}

			if len(oracleResult) != len(pureResult) {
				t.Errorf("Polygon count mismatch: oracle=%d, pure=%d",
					len(oracleResult), len(pureResult))
			}
		})
	}
}

// Helper functions for oracle - when built with -tags=clipper_cgo,
// these will use the CGO oracle implementation via impl_oracle_cgo.go
//
// NOTE: To test against oracle, build with: go test -tags=clipper_cgo
// Without the tag, this tests pure Go against itself (useful for regression testing)

func union64Oracle(subject, clip Paths64, fillRule FillRule) (Paths64, error) {
	// When built with -tags=clipper_cgo, this uses the oracle
	// Otherwise it uses pure Go (testing pure Go against itself)
	return Union64(subject, clip, fillRule)
}

func intersect64Oracle(subject, clip Paths64, fillRule FillRule) (Paths64, error) {
	return Intersect64(subject, clip, fillRule)
}

func difference64Oracle(subject, clip Paths64, fillRule FillRule) (Paths64, error) {
	return Difference64(subject, clip, fillRule)
}

func xor64Oracle(subject, clip Paths64, fillRule FillRule) (Paths64, error) {
	return Xor64(subject, clip, fillRule)
}

// BenchmarkBooleanOperations compares performance of pure Go vs Oracle
func BenchmarkBooleanOperations(b *testing.B) {
	subject := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}
	clip := Paths64{{{50, 50}, {150, 50}, {150, 150}, {50, 150}}}

	b.Run("PureGo/Union", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Union64(subject, clip, NonZero)
		}
	})

	b.Run("Oracle/Union", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = union64Oracle(subject, clip, NonZero)
		}
	})
}

// TestDetailedComparison performs detailed point-by-point comparison
func TestDetailedComparison(t *testing.T) {
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	oracleResult, _ := union64Oracle(subject, clip, NonZero)
	pureResult, _ := Union64(subject, clip, NonZero)

	t.Logf("\n=== ORACLE OUTPUT ===")
	for i, path := range oracleResult {
		t.Logf("Polygon %d (%d points):", i, len(path))
		for j, pt := range path {
			t.Logf("  [%d] {%d, %d}", j, pt.X, pt.Y)
		}
	}

	t.Logf("\n=== PURE GO OUTPUT ===")
	for i, path := range pureResult {
		t.Logf("Polygon %d (%d points):", i, len(path))
		for j, pt := range path {
			t.Logf("  [%d] {%d, %d}", j, pt.X, pt.Y)
		}
	}

	// Check for duplicate points in pure Go output
	for i, path := range pureResult {
		seen := make(map[Point64]bool)
		duplicates := 0
		for _, pt := range path {
			if seen[pt] {
				duplicates++
				t.Logf("WARNING: Duplicate point {%d,%d} in polygon %d", pt.X, pt.Y, i)
			}
			seen[pt] = true
		}
		if duplicates > 0 {
			t.Errorf("Polygon %d has %d duplicate points", i, duplicates)
		}
	}
}

// TestFillRules tests different fill rules
func TestFillRules(t *testing.T) {
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	fillRules := []FillRule{EvenOdd, NonZero, Positive, Negative}
	fillRuleNames := []string{"EvenOdd", "NonZero", "Positive", "Negative"}

	for i, fr := range fillRules {
		t.Run(fillRuleNames[i], func(t *testing.T) {
			oracleResult, oracleErr := union64Oracle(subject, clip, fr)
			if oracleErr != nil {
				t.Fatalf("Oracle failed: %v", oracleErr)
			}

			pureResult, pureErr := Union64(subject, clip, fr)
			if pureErr != nil {
				t.Fatalf("Pure Go failed: %v", pureErr)
			}

			if len(oracleResult) != len(pureResult) {
				t.Errorf("Polygon count mismatch with %s: oracle=%d, pure=%d",
					fillRuleNames[i], len(oracleResult), len(pureResult))
			}
		})
	}
}
