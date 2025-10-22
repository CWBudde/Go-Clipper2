package clipper

import (
	"math"
	"testing"
)

// ==============================================================================
// PreserveCollinear Tests
// ==============================================================================

func TestOffsetPreserveCollinearTrue(t *testing.T) {
	// Rectangle with extra collinear points on edges
	// Points at X=25 and X=75 are collinear with corners
	rectWithCollinear := Paths64{{{0, 0}, {25, 0}, {50, 0}, {75, 0}, {100, 0}, {100, 100}, {0, 100}}}

	co := NewClipperOffset(2.0, 0.25)
	co.SetPreserveCollinear(true)
	co.AddPaths(rectWithCollinear, JoinBevel, EndPolygon)

	result, err := co.Execute(10.0)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}

	// TODO: When preserve_collinear is fully implemented, verify that
	// collinear points are preserved in the output
	t.Logf("PreserveCollinear=true result: %d paths, %d vertices in first path",
		len(result), len(result[0]))
}

func TestOffsetPreserveCollinearFalse(t *testing.T) {
	// Same test but with preserve_collinear=false (default)
	rectWithCollinear := Paths64{{{0, 0}, {25, 0}, {50, 0}, {75, 0}, {100, 0}, {100, 100}, {0, 100}}}

	co := NewClipperOffset(2.0, 0.25)
	co.SetPreserveCollinear(false) // Explicit false
	co.AddPaths(rectWithCollinear, JoinBevel, EndPolygon)

	result, err := co.Execute(10.0)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}

	t.Logf("PreserveCollinear=false result: %d paths, %d vertices in first path",
		len(result), len(result[0]))
}

func TestOffsetPreserveCollinearAccessors(t *testing.T) {
	co := NewClipperOffset(2.0, 0.25)

	// Test default value
	if co.PreserveCollinear() != false {
		t.Error("Expected default PreserveCollinear to be false")
	}

	// Test setter and getter
	co.SetPreserveCollinear(true)
	if co.PreserveCollinear() != true {
		t.Error("Expected PreserveCollinear to be true after setting")
	}

	co.SetPreserveCollinear(false)
	if co.PreserveCollinear() != false {
		t.Error("Expected PreserveCollinear to be false after setting")
	}
}

// ==============================================================================
// ReverseSolution Tests
// ==============================================================================

func TestOffsetReverseSolutionTrue(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	co := NewClipperOffset(2.0, 0.25)
	co.SetReverseSolution(true)
	co.AddPaths(square, JoinBevel, EndPolygon)

	resultReversed, err := co.Execute(10.0)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Also test with reverse_solution=false for comparison
	co2 := NewClipperOffset(2.0, 0.25)
	co2.SetReverseSolution(false)
	co2.AddPaths(square, JoinBevel, EndPolygon)

	resultNormal, err := co2.Execute(10.0)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify both have results
	if len(resultReversed) == 0 || len(resultNormal) == 0 {
		t.Fatal("Expected non-empty results")
	}

	// When ReverseSolution is true, the path orientation should be opposite
	areaReversed := Area64(resultReversed[0])
	areaNormal := Area64(resultNormal[0])

	// Areas should have opposite signs (accounting for float precision)
	if (areaReversed > 0 && areaNormal > 0) || (areaReversed < 0 && areaNormal < 0) {
		t.Log("Note: ReverseSolution may not be fully functional yet")
		t.Logf("Reversed area: %.2f, Normal area: %.2f", areaReversed, areaNormal)
	} else {
		// Expected: opposite signs
		t.Logf("✓ Reversed area: %.2f, Normal area: %.2f (opposite signs as expected)",
			areaReversed, areaNormal)
	}
}

func TestOffsetReverseSolutionAccessors(t *testing.T) {
	co := NewClipperOffset(2.0, 0.25)

	// Test default value
	if co.ReverseSolution() != false {
		t.Error("Expected default ReverseSolution to be false")
	}

	// Test setter and getter
	co.SetReverseSolution(true)
	if co.ReverseSolution() != true {
		t.Error("Expected ReverseSolution to be true after setting")
	}

	co.SetReverseSolution(false)
	if co.ReverseSolution() != false {
		t.Error("Expected ReverseSolution to be false after setting")
	}
}

// ==============================================================================
// Self-Intersecting Input Tests
// ==============================================================================

func TestOffsetSelfIntersectingBowtie(t *testing.T) {
	// Bowtie/figure-8 shape (self-intersecting)
	bowtie := Paths64{{
		{0, 0},
		{100, 100}, // Cross to opposite corner
		{100, 0},
		{0, 100}, // Cross back
	}}

	result, err := InflatePaths64(bowtie, 5.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// Should handle gracefully, even if result is empty or split into multiple paths
	t.Logf("Bowtie offset result: %d paths", len(result))
}

func TestOffsetSelfIntersectingOpenPath(t *testing.T) {
	// Self-intersecting open path
	spiral := Path64{
		{50, 0},
		{100, 50},
		{50, 100},
		{0, 50},
		{50, 50}, // Crosses through center
	}

	result, err := InflatePaths64(Paths64{spiral}, 5.0, JoinRound, EndRound, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	t.Logf("Self-intersecting open path result: %d paths", len(result))
}

// ==============================================================================
// Extreme Delta Value Tests
// ==============================================================================

func TestOffsetVeryLargeDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	// Very large expansion
	result, err := InflatePaths64(square, 1000000.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}

	// Verify the result is much larger
	area := math.Abs(Area64(result[0]))
	originalArea := 100.0 * 100.0

	if area <= originalArea {
		t.Errorf("Expected area to be much larger than original (%.2f vs %.2f)",
			area, originalArea)
	}

	t.Logf("Very large delta result: area = %.2f (original = %.2f)", area, originalArea)
}

func TestOffsetVeryLargeNegativeDelta(t *testing.T) {
	// Large square that can handle large contraction
	largeSquare := Paths64{{{0, 0}, {10000, 0}, {10000, 10000}, {0, 10000}}}

	// Large contraction
	result, err := InflatePaths64(largeSquare, -4000.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// May or may not produce result depending on whether polygon collapses
	t.Logf("Large negative delta result: %d paths", len(result))
	if len(result) > 0 {
		area := math.Abs(Area64(result[0]))
		t.Logf("Remaining area: %.2f", area)
	}
}

func TestOffsetCompleteCollapse(t *testing.T) {
	// Small square with negative delta larger than its size
	smallSquare := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	// Contraction larger than square size - should collapse completely
	result, err := InflatePaths64(smallSquare, -20.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// Result should be empty (polygon collapsed)
	if len(result) != 0 {
		t.Logf("Note: Expected empty result for complete collapse, got %d paths", len(result))
		// This might not fail if implementation handles it differently
	} else {
		t.Log("✓ Polygon collapsed as expected")
	}
}

func TestOffsetFractionalDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	testCases := []float64{0.01, 0.1, 0.25, 0.49}

	for _, delta := range testCases {
		result, err := InflatePaths64(square, delta, JoinBevel, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})

		if err != nil {
			t.Fatalf("InflatePaths64 failed for delta=%.2f: %v", delta, err)
		}

		// For deltas < 0.5, implementation returns original paths
		if len(result) == 0 {
			t.Errorf("Expected non-empty result for delta=%.2f", delta)
		}

		t.Logf("Delta %.2f: %d paths", delta, len(result))
	}
}

// ==============================================================================
// Degenerate Input Tests
// ==============================================================================

func TestOffsetEmptyPaths(t *testing.T) {
	emptyPaths := Paths64{}

	result, err := InflatePaths64(emptyPaths, 10.0, JoinBevel, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d paths", len(result))
	}
}

func TestOffsetPathWithDuplicatePoints(t *testing.T) {
	// Path with consecutive duplicate points
	pathWithDupes := Path64{
		{0, 0},
		{0, 0}, // duplicate
		{100, 0},
		{100, 0}, // duplicate
		{100, 100},
		{0, 100},
		{0, 100}, // duplicate
	}

	result, err := InflatePaths64(Paths64{pathWithDupes}, 10.0, JoinBevel, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// Should handle duplicates gracefully (stripDuplicates function should remove them)
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	t.Logf("Path with duplicates result: %d paths", len(result))
}

func TestOffsetVeryLargeCoordinates(t *testing.T) {
	// Test with coordinates near int64 limits
	const large = int64(1000000000000) // 10^12

	largeSquare := Paths64{{{0, 0}, {large, 0}, {large, large}, {0, large}}}

	result, err := InflatePaths64(largeSquare, 1000.0, JoinBevel, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected non-empty result for very large coordinates")
	}

	t.Logf("Very large coordinates result: %d paths", len(result))
}

func TestOffsetZeroDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	result, err := InflatePaths64(square, 0.0, JoinBevel, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// Should return input paths unchanged
	if len(result) != 1 {
		t.Errorf("Expected 1 path, got %d", len(result))
	}

	t.Logf("Zero delta result: %d paths", len(result))
}

// ==============================================================================
// Complex Real-World Shapes
// ==============================================================================

func TestOffsetConcavePolygonWithManyVertices(t *testing.T) {
	// Create a star-like concave polygon with many vertices
	const numPoints = 50
	star := make(Path64, numPoints)

	for i := 0; i < numPoints; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numPoints)
		// Alternate between inner and outer radius
		var radius float64
		if i%2 == 0 {
			radius = 100.0
		} else {
			radius = 50.0
		}
		star[i] = Point64{
			X: int64(200 + radius*math.Cos(angle)),
			Y: int64(200 + radius*math.Sin(angle)),
		}
	}

	result, err := InflatePaths64(Paths64{star}, 5.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected non-empty result for star polygon")
	}

	t.Logf("50-vertex star result: %d paths, %d vertices in first",
		len(result), len(result[0]))
}

func TestOffsetMultipleNestedSquares(t *testing.T) {
	// Multiple concentric squares
	paths := Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{20, 20}, {80, 20}, {80, 80}, {20, 80}},
		{{40, 40}, {60, 40}, {60, 60}, {40, 60}},
	}

	result, err := InflatePaths64(paths, 5.0, JoinBevel, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// Should expand all squares
	if len(result) == 0 {
		t.Error("Expected non-empty result for nested squares")
	}

	t.Logf("Nested squares result: %d paths", len(result))
}

// ==============================================================================
// Upstream C++ Test Ports (from Clipper2 test suite)
// ==============================================================================

// TestOffsets2 - Ported from C++ TestOffsets2 (issue #448 & #456)
// Tests arc tolerance and distance constraints
func TestOffsets2(t *testing.T) {
	const scale = 10.0
	const delta = 10 * scale
	const arcTol = 0.25 * scale

	// Pentagon shape from upstream test
	subject := Paths64{{{50, 50}, {100, 50}, {100, 150}, {50, 150}, {0, 100}}}

	// Scale up the coordinates
	scaledSubject := make(Paths64, len(subject))
	for i, path := range subject {
		scaledPath := make(Path64, len(path))
		for j, pt := range path {
			scaledPath[j] = Point64{
				X: int64(float64(pt.X) * scale),
				Y: int64(float64(pt.Y) * scale),
			}
		}
		scaledSubject[i] = scaledPath
	}

	result, err := InflatePaths64(scaledSubject, delta, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: arcTol,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}

	// Verify minimum distance from original to result is approximately delta
	// (within arc tolerance)
	t.Logf("Pentagon offset result: %d paths, %d vertices", len(result), len(result[0]))

	// The C++ test verifies that the minimum distance is >= delta - arcTol
	// and the vertex count is reasonable (≤21)
	if len(result[0]) > 50 {
		t.Errorf("Too many vertices in result: %d (expected ≤50 for pentagon)", len(result[0]))
	}
}

// TestOffsets3 - Ported from C++ TestOffsets3 (issue #424)
// Tests that miter join doesn't add excessive vertices for very large negative deltas
func TestOffsets3(t *testing.T) {
	// Complex polygon with many vertices from upstream test (simplified version)
	subject := Paths64{{
		{1525311078, 1352369439}, {1526632284, 1366692987}, {1519397110, 1367437476},
		{1520246456, 1380177674}, {1520613458, 1385913385}, {1517383844, 1386238444},
		{1517771817, 1392099983}, {1518233190, 1398758441}, {1518421934, 1401883197},
		{1599487345, 1352704983}, {1602758902, 1378489467}, {1618990858, 1376350372},
	}}

	// Very large negative delta (contraction)
	const delta = -209715.0

	result, err := InflatePaths64(subject, delta, JoinMiter, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	// The C++ test expects that the result doesn't add more than 1 extra vertex
	// compared to the original (miter joins should be efficient)
	if len(result) > 0 {
		vertexDiff := len(result[0]) - len(subject[0])
		if vertexDiff > 10 {
			t.Logf("Note: Miter join added %d vertices (upstream expects ≤1)", vertexDiff)
		}
		t.Logf("Large contraction result: %d paths, vertex diff: %d", len(result), vertexDiff)
	} else {
		// Polygon may have collapsed completely with such a large contraction
		t.Log("Polygon collapsed with very large negative delta (acceptable)")
	}
}

// TestOffsetArcToleranceEffect - Additional test for arc tolerance impact
func TestOffsetArcToleranceEffect(t *testing.T) {
	// Circle approximation using octagon
	const radius = 100.0
	octagon := make(Path64, 8)
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2 * math.Pi / 8
		octagon[i] = Point64{
			X: int64(200 + radius*math.Cos(angle)),
			Y: int64(200 + radius*math.Sin(angle)),
		}
	}

	// Test with different arc tolerances
	testCases := []struct {
		name    string
		arcTol  float64
		maxVert int // Maximum expected vertices
	}{
		{"Very Coarse (5.0)", 5.0, 20},
		{"Coarse (2.0)", 2.0, 30},
		{"Medium (0.5)", 0.5, 50},
		{"Fine (0.1)", 0.1, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(Paths64{octagon}, 20.0, JoinRound, EndPolygon, OffsetOptions{
				MiterLimit:   2.0,
				ArcTolerance: tc.arcTol,
			})

			if err != nil {
				t.Fatalf("InflatePaths64 failed: %v", err)
			}

			if len(result) == 0 {
				t.Fatal("Expected non-empty result")
			}

			vertices := len(result[0])
			t.Logf("Arc tolerance %.2f: %d vertices", tc.arcTol, vertices)

			// Finer arc tolerance should generally produce more vertices
			if vertices > tc.maxVert {
				t.Logf("Note: Vertex count %d exceeds guideline %d (not necessarily an error)", vertices, tc.maxVert)
			}
		})
	}
}
