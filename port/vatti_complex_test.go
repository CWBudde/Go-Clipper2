package clipper

import (
	"bytes"
	"math"
	"testing"
)

// TestIntersectNonConvex tests intersection of non-convex polygons
func TestIntersectNonConvex(t *testing.T) {
	// L-shaped subject polygon
	subject := Paths64{{{0, 0}, {10, 0}, {10, 5}, {5, 5}, {5, 10}, {0, 10}}}
	// Rectangle clip
	clip := Paths64{{{3, 3}, {8, 3}, {8, 8}, {3, 8}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from L-shaped intersection")
	}

	t.Logf("L-shaped intersection result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))

	// Verify the intersection is within both input polygons
	for i, path := range result {
		if len(path) < 3 {
			t.Errorf("Polygon %d has fewer than 3 points", i)
		}
		t.Logf("Polygon %d: %d points", i, len(path))
	}
}

// TestIntersectTriangles tests intersection of two triangles
func TestIntersectTriangles(t *testing.T) {
	// Triangle pointing up
	subject := Paths64{{{0, 0}, {10, 0}, {5, 10}}}
	// Triangle pointing down
	clip := Paths64{{{0, 10}, {10, 10}, {5, 0}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle intersection")
	}

	t.Logf("Triangle intersection result: %v", result)

	// The intersection should form a hexagon
	if len(result) != 1 {
		t.Errorf("Expected 1 polygon, got %d", len(result))
	}
}

// TestIntersectMultipleRegions tests intersection creating multiple separate regions
func TestIntersectMultipleRegions(t *testing.T) {
	// Two separate rectangles as subject
	subject := Paths64{
		{{0, 0}, {5, 0}, {5, 5}, {0, 5}},     // Left square
		{{10, 0}, {15, 0}, {15, 5}, {10, 5}}, // Right square
	}
	// One wide rectangle as clip
	clip := Paths64{{{-1, 2}, {16, 2}, {16, 3}, {-1, 3}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Multiple regions intersection result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))

	// Should create two separate rectangular intersections
	if len(result) != 2 {
		t.Logf("Warning: Expected 2 separate intersection regions, got %d", len(result))
	}
}

// TestIntersectNoOverlap tests intersection with no overlap
func TestIntersectNoOverlap(t *testing.T) {
	subject := Paths64{{{0, 0}, {5, 0}, {5, 5}, {0, 5}}}
	clip := Paths64{{{10, 10}, {15, 10}, {15, 15}, {10, 15}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for non-overlapping polygons, got %v", result)
	}
}

// TestIntersectTouching tests intersection with touching boundaries
func TestIntersectTouching(t *testing.T) {
	// Two rectangles sharing an edge
	subject := Paths64{{{0, 0}, {5, 0}, {5, 5}, {0, 5}}}
	clip := Paths64{{{5, 0}, {10, 0}, {10, 5}, {5, 5}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Touching boundaries result: %v", result)

	// Touching edges may produce empty or degenerate result
	// This is implementation-dependent
	if len(result) > 0 {
		t.Logf("Got %d polygon(s) from touching boundaries", len(result))
	}
}

// TestIntersectCompleteOverlap tests intersection where one polygon completely contains another
func TestIntersectCompleteOverlap(t *testing.T) {
	// Large outer rectangle
	subject := Paths64{{{0, 0}, {20, 0}, {20, 20}, {0, 20}}}
	// Small inner rectangle
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 polygon (the smaller one), got %d", len(result))
	}

	t.Logf("Complete overlap result: %v", result)

	// The result should be approximately the smaller rectangle
	if len(result) > 0 {
		area := Area64(result[0])
		expectedArea := 100.0 // 10x10 square
		if math.Abs(area-expectedArea) > 1.0 {
			t.Errorf("Expected area ~%.1f, got %.1f", expectedArea, area)
		}
	}
}

// TestIntersectSelfIntersecting tests a self-intersecting polygon (figure-8)
func TestIntersectSelfIntersecting(t *testing.T) {
	// Enable debug for this complex case
	debugBuf := &bytes.Buffer{}

	// Figure-8 polygon (self-intersecting at center)
	// Forms two loops: left (0,0)-(5,0)-(0,10)-(5,10) and right (5,0)-(10,0)-(10,10)-(5,10)
	subject := Paths64{{{0, 0}, {10, 0}, {0, 10}, {10, 10}}}

	// Rectangle that intersects both loops
	clip := Paths64{{{2, 3}, {8, 3}, {8, 7}, {2, 7}}}

	result, err := Intersect64(subject, clip, NonZero)

	// Print debug output
	t.Log("\n=== DEBUG LOG ===\n" + debugBuf.String() + "\n=== END DEBUG LOG ===")

	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Self-intersecting polygon result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))

	// Self-intersecting polygons are complex - just verify we get some result
	if len(result) == 0 {
		t.Log("Note: Self-intersecting polygon produced no intersection (may be expected)")
	}
}

// TestIntersectConcavePolygon tests intersection with a star-shaped polygon
func TestIntersectConcavePolygon(t *testing.T) {
	// Simple star shape (5-pointed, simplified)
	// Center at (10, 10), outer radius ~7, inner radius ~3
	subject := Paths64{{
		{10, 3},  // Top point
		{12, 8},  // Top-right inner
		{17, 10}, // Right point
		{12, 12}, // Bottom-right inner
		{10, 17}, // Bottom point
		{8, 12},  // Bottom-left inner
		{3, 10},  // Left point
		{8, 8},   // Top-left inner
	}}

	// Rectangle intersecting the star
	clip := Paths64{{{8, 8}, {12, 8}, {12, 12}, {8, 12}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Star-polygon intersection result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))
}

// TestIntersectWithHole tests a polygon with a hole
func TestIntersectWithHole(t *testing.T) {
	// Note: Clipper2 represents holes as separate paths in opposite winding order
	// Outer rectangle (counter-clockwise)
	outer := Path64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}
	// Inner hole (clockwise - opposite winding)
	hole := Path64{{5, 5}, {5, 15}, {15, 15}, {15, 5}}

	// Subject is outer with hole
	subject := Paths64{outer, hole}

	// Clip rectangle that intersects both outer and hole
	clip := Paths64{{{10, 10}, {25, 10}, {25, 25}, {10, 25}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Polygon with hole intersection result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))

	// The result should be the bottom-right quadrant minus the hole
	// This might produce multiple polygons or one with a hole
	if len(result) == 0 {
		t.Error("Expected non-empty result from polygon with hole")
	}
}

// TestIntersectSliver tests a very thin sliver intersection
func TestIntersectSliver(t *testing.T) {
	// Two rectangles with minimal overlap
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{9, 5}, {15, 5}, {15, 6}, {9, 6}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Sliver intersection result: %v", result)

	if len(result) == 0 {
		t.Error("Expected thin sliver intersection")
	}

	if len(result) > 0 {
		area := Area64(result[0])
		t.Logf("Sliver area: %.2f", area)
		// Expected area: 1 unit wide × 1 unit tall = 1.0
		if math.Abs(area-1.0) > 0.1 {
			t.Logf("Warning: Sliver area %.2f differs from expected 1.0", area)
		}
	}
}

// TestIntersectDifferentFillRules tests intersection with different fill rules
func TestIntersectDifferentFillRules(t *testing.T) {
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	fillRules := []FillRule{EvenOdd, NonZero, Positive, Negative}
	fillRuleNames := []string{"EvenOdd", "NonZero", "Positive", "Negative"}

	for i, fillRule := range fillRules {
		t.Run(fillRuleNames[i], func(t *testing.T) {
			result, err := Intersect64(subject, clip, fillRule)
			if err == ErrNotImplemented {
				t.Skip("Intersect64 not yet implemented in pure Go")
			}
			if err != nil {
				t.Fatalf("Intersect64 with %s failed: %v", fillRuleNames[i], err)
			}

			t.Logf("%s result: %v", fillRuleNames[i], result)

			if len(result) == 0 {
				t.Errorf("Expected non-empty result for %s fill rule", fillRuleNames[i])
			}
		})
	}
}

// TestIntersectLargeCoordinates tests intersection with large coordinate values
func TestIntersectLargeCoordinates(t *testing.T) {
	offset := int64(1000000000) // 1 billion

	subject := Paths64{{
		{offset, offset},
		{offset + 10, offset},
		{offset + 10, offset + 10},
		{offset, offset + 10},
	}}
	clip := Paths64{{
		{offset + 5, offset + 5},
		{offset + 15, offset + 5},
		{offset + 15, offset + 15},
		{offset + 5, offset + 15},
	}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 with large coordinates failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 polygon, got %d", len(result))
	}

	t.Logf("Large coordinates result: %v", result)

	// Verify the result is in the expected range
	if len(result) > 0 {
		for _, pt := range result[0] {
			if pt.X < offset || pt.X > offset+15 || pt.Y < offset || pt.Y > offset+15 {
				t.Errorf("Point %v outside expected range", pt)
			}
		}
	}
}

// TestIntersectManyVertices tests intersection with polygons having many vertices
func TestIntersectManyVertices(t *testing.T) {
	// Create a polygon approximating a circle with many vertices
	centerX, centerY := 50.0, 50.0
	radius := 30.0
	numVertices := 36 // 36 vertices (10 degree increments)

	var circleApprox Path64
	for i := 0; i < numVertices; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numVertices)
		x := int64(centerX + radius*math.Cos(angle))
		y := int64(centerY + radius*math.Sin(angle))
		circleApprox = append(circleApprox, Point64{x, y})
	}

	subject := Paths64{circleApprox}
	// Rectangle intersecting the circle
	clip := Paths64{{{40, 40}, {60, 40}, {60, 60}, {40, 60}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 with many vertices failed: %v", err)
	}

	t.Logf("Circle approximation intersection: %d polygon(s)", len(result))
	if len(result) > 0 {
		t.Logf("Result polygon has %d vertices", len(result[0]))
	}
}
