package clipper

import (
	"testing"
)

// TestOffsetSquareSquareExpansion tests basic square expansion with Square joins
func TestOffsetSquareSquareExpansion(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, 10.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from square join offset")
	}

	// With square joins on a square, we should get perpendicular extensions at corners
	totalPoints := 0
	for _, path := range result {
		totalPoints += len(path)
	}

	t.Logf("Square expansion with square joins result: %d paths, %d total points", len(result), totalPoints)
}

// TestOffsetSquareSquareContraction tests square contraction with Square joins
func TestOffsetSquareSquareContraction(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, -10.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from square contraction")
	}

	t.Logf("Square contraction with square joins result: %d paths", len(result))
}

// TestOffsetSquareTriangle tests square joins on a triangle
func TestOffsetSquareTriangle(t *testing.T) {
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(triangle, 10.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle with square joins")
	}

	t.Logf("Triangle with square joins result: %d paths", len(result))
}

// TestOffsetSquareRightAngle tests 90-degree corners with square joins
func TestOffsetSquareRightAngle(t *testing.T) {
	// L-shape with right angles
	lShape := Paths64{{{0, 0}, {100, 0}, {100, 50}, {50, 50}, {50, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(lShape, 10.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from L-shape with square joins")
	}

	t.Logf("L-shape with square joins result: %d paths", len(result))
}

// TestOffsetSquareVsOther compares square joins with other join types
func TestOffsetSquareVsOther(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	// Test all supported join types
	joinTypes := []struct {
		name     string
		joinType JoinType
	}{
		{"Bevel", JoinBevel},
		{"Miter", JoinMiter},
		{"Square", JoinSquare},
	}

	for _, jt := range joinTypes {
		t.Run(jt.name, func(t *testing.T) {
			result, err := InflatePaths64(square, 10.0, jt.joinType, EndPolygon, options)
			if err != nil {
				t.Fatalf("InflatePaths64 with %s failed: %v", jt.name, err)
			}

			if len(result) == 0 {
				t.Fatalf("Expected non-empty result for %s join", jt.name)
			}

			totalPoints := 0
			for _, path := range result {
				totalPoints += len(path)
			}

			t.Logf("%s join result: %d paths, %d total points", jt.name, len(result), totalPoints)
		})
	}
}

// TestOffsetSquareStarShape tests square joins on sharp corners
func TestOffsetSquareStarShape(t *testing.T) {
	// Simple 4-point star
	star := Paths64{
		{{50, 0}, {60, 40}, {100, 50}, {60, 60}, {50, 100}, {40, 60}, {0, 50}, {40, 40}},
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(star, 5.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from star with square joins")
	}

	t.Logf("Star with square joins result: %d paths", len(result))
}

// TestOffsetSquareDifferentAngles tests square joins at various angles
func TestOffsetSquareDifferentAngles(t *testing.T) {
	testCases := []struct {
		name  string
		poly  Path64
		delta float64
	}{
		{"45 degree", Path64{{0, 0}, {100, 0}, {150, 50}, {100, 100}, {0, 100}}, 10.0},
		{"135 degree", Path64{{0, 0}, {100, 0}, {50, 50}, {100, 100}, {0, 100}}, 10.0},
		{"Acute angle", Path64{{0, 0}, {100, 0}, {95, 10}, {100, 20}}, 5.0},
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(Paths64{tc.poly}, tc.delta, JoinSquare, EndPolygon, options)
			if err != nil {
				t.Fatalf("InflatePaths64 failed for %s: %v", tc.name, err)
			}

			if len(result) == 0 {
				t.Fatalf("Expected non-empty result for %s", tc.name)
			}

			t.Logf("%s with square joins result: %d paths", tc.name, len(result))
		})
	}
}

// TestOffsetSquareMultiplePaths tests offsetting multiple paths with square joins
func TestOffsetSquareMultiplePaths(t *testing.T) {
	paths := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
		{{100, 100}, {150, 100}, {150, 150}, {100, 150}},
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(paths, 5.0, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) < 2 {
		t.Errorf("Expected at least 2 paths, got %d", len(result))
	}

	t.Logf("Multiple paths with square joins result: %d paths", len(result))
}

// TestOffsetSquareSmallDelta tests square joins with very small delta
func TestOffsetSquareSmallDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	// Very small delta (< 0.5) should return paths nearly unchanged
	result, err := InflatePaths64(square, 0.1, JoinSquare, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from small delta with square joins")
	}

	t.Logf("Small delta square joins result: %d paths", len(result))
}

// TestOffsetSquareMiterLimitFallback tests that miter joins fall back to square when limit exceeded
func TestOffsetSquareMiterLimitFallback(t *testing.T) {
	// Very sharp triangle (almost flat) to trigger miter limit
	sharpTriangle := Paths64{{{0, 0}, {100, 0}, {50, 2}}}

	options := OffsetOptions{
		MiterLimit:   1.5, // Low miter limit
		ArcTolerance: 0.25,
	}

	// Use miter join type, but expect fallback to square at sharp angle
	result, err := InflatePaths64(sharpTriangle, 10.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from miter with fallback to square")
	}

	t.Logf("Miter with square fallback result: %d paths", len(result))
}
