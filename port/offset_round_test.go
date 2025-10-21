package clipper

import (
	"math"
	"testing"
)

// TestOffsetRoundSquareExpansion tests basic square expansion with Round joins
func TestOffsetRoundSquareExpansion(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, 10.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from round offset")
	}

	// Count total points
	totalPoints := 0
	for _, path := range result {
		totalPoints += len(path)
	}

	t.Logf("Square round expansion result: %d paths, %d total points", len(result), totalPoints)
}

// TestOffsetRoundSquareContraction tests square contraction with Round joins
func TestOffsetRoundSquareContraction(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, -10.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from round contraction")
	}

	t.Logf("Square round contraction result: %d paths", len(result))
}

// TestOffsetRoundCircleApproximation tests round joins on an octagon (circular approximation)
func TestOffsetRoundCircleApproximation(t *testing.T) {
	// Create an octagon (8-sided approximation of a circle)
	octagon := make(Path64, 8)
	radius := 50.0
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2.0 * math.Pi / 8.0
		octagon[i] = Point64{
			X: int64(radius * math.Cos(angle)),
			Y: int64(radius * math.Sin(angle)),
		}
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(Paths64{octagon}, 10.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from octagon offset")
	}

	t.Logf("Octagon round offset result: %d paths", len(result))
}

// TestOffsetRoundArcTolerance tests different arc tolerance values
func TestOffsetRoundArcTolerance(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	tolerances := []float64{0.1, 0.5, 1.0, 2.0}
	for _, tol := range tolerances {
		options := OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: tol,
		}

		result, err := InflatePaths64(square, 10.0, JoinRound, EndPolygon, options)
		if err != nil {
			t.Fatalf("InflatePaths64 failed with tolerance %f: %v", tol, err)
		}

		if len(result) == 0 {
			t.Fatal("Expected non-empty result")
		}

		totalPoints := 0
		for _, path := range result {
			totalPoints += len(path)
		}

		t.Logf("Arc tolerance %f: %d total points", tol, totalPoints)
	}
}

// TestOffsetRoundSharpAngles tests round joins on acute angles
func TestOffsetRoundSharpAngles(t *testing.T) {
	// Triangle with sharp angle at top
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 150}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(triangle, 10.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle offset")
	}

	totalPoints := 0
	for _, path := range result {
		totalPoints += len(path)
	}

	t.Logf("Triangle round offset result: %d paths, %d points", len(result), totalPoints)
}

// TestOffsetRoundTriangle tests offsetting a triangle with round joins
func TestOffsetRoundTriangle(t *testing.T) {
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(triangle, 10.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle offset")
	}

	t.Logf("Triangle offset result: %d paths", len(result))
}

// TestOffsetRoundVsOther compares round joins with other join types
func TestOffsetRoundVsOther(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	joinTypes := []JoinType{JoinBevel, JoinMiter, JoinSquare, JoinRound}
	joinNames := []string{"Bevel", "Miter", "Square", "Round"}

	for i, jt := range joinTypes {
		result, err := InflatePaths64(square, 10.0, jt, EndPolygon, options)
		if err != nil {
			t.Fatalf("InflatePaths64 failed for %s: %v", joinNames[i], err)
		}

		totalPoints := 0
		for _, path := range result {
			totalPoints += len(path)
		}

		t.Logf("%s join: %d paths, %d points", joinNames[i], len(result), totalPoints)
	}
}

// TestOffsetRoundMultiplePaths tests offsetting multiple paths at once
func TestOffsetRoundMultiplePaths(t *testing.T) {
	paths := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
		{{100, 100}, {150, 100}, {150, 150}, {100, 150}},
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(paths, 5.0, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) < 2 {
		t.Errorf("Expected at least 2 paths, got %d", len(result))
	}

	t.Logf("Multiple paths round offset result: %d paths", len(result))
}

// TestOffsetRoundSmallDelta tests offsetting with a very small delta
func TestOffsetRoundSmallDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	// Very small delta (< 0.5) should return paths nearly unchanged
	result, err := InflatePaths64(square, 0.1, JoinRound, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from small delta offset")
	}

	t.Logf("Small delta round offset result: %d paths", len(result))
}

// TestOffsetRoundAccessorMethods tests the ArcTolerance getter and setter
func TestOffsetRoundAccessorMethods(t *testing.T) {
	co := NewClipperOffset(2.0, 0.25)

	// Test initial value
	if co.ArcTolerance() != 0.25 {
		t.Errorf("Expected initial arc tolerance 0.25, got %f", co.ArcTolerance())
	}

	// Test setter
	co.SetArcTolerance(1.0)
	if co.ArcTolerance() != 1.0 {
		t.Errorf("Expected arc tolerance 1.0 after set, got %f", co.ArcTolerance())
	}

	// Test with zero/negative (should use default calculation)
	co.SetArcTolerance(0.0)
	if co.ArcTolerance() != 0.0 {
		t.Errorf("Expected arc tolerance 0.0, got %f", co.ArcTolerance())
	}

	co.SetArcTolerance(-0.5)
	if co.ArcTolerance() != -0.5 {
		t.Errorf("Expected arc tolerance -0.5, got %f", co.ArcTolerance())
	}
}
