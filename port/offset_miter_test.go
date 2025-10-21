package clipper

import (
	"testing"
)

// TestOffsetMiterSquareExpansion tests basic square expansion with Miter joins
func TestOffsetMiterSquareExpansion(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, 10.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from miter offset")
	}

	// With miter joins on a square, we should get clean sharp corners
	// Each corner should produce a single miter point instead of two bevel points
	totalPoints := 0
	for _, path := range result {
		totalPoints += len(path)
	}

	t.Logf("Square miter expansion result: %d paths, %d total points", len(result), totalPoints)
}

// TestOffsetMiterSquareContraction tests square contraction with Miter joins
func TestOffsetMiterSquareContraction(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(square, -10.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from miter contraction")
	}

	t.Logf("Square miter contraction result: %d paths", len(result))
}

// TestOffsetMiterSharpAngles tests miter joins on acute angles
func TestOffsetMiterSharpAngles(t *testing.T) {
	// Triangle with sharp angle at top
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 150}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(triangle, 10.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle offset")
	}

	t.Logf("Triangle miter offset result: %d paths", len(result))
}

// TestOffsetMiterLimitExceeded tests behavior when miter limit is exceeded
func TestOffsetMiterLimitExceeded(t *testing.T) {
	// Very sharp triangle (almost flat)
	sharpTriangle := Paths64{{{0, 0}, {100, 0}, {50, 5}}}

	// Use small miter limit to trigger fallback
	options := OffsetOptions{
		MiterLimit:   1.5,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(sharpTriangle, 10.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result even with miter limit exceeded")
	}

	t.Logf("Sharp triangle with miter limit result: %d paths", len(result))
}

// TestOffsetMiterStarPolygon tests miter joins on a star shape with many sharp corners
func TestOffsetMiterStarPolygon(t *testing.T) {
	// Simple 4-point star
	star := Paths64{
		{{50, 0}, {60, 40}, {100, 50}, {60, 60}, {50, 100}, {40, 60}, {0, 50}, {40, 40}},
	}

	options := OffsetOptions{
		MiterLimit:   3.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(star, 5.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from star offset")
	}

	t.Logf("Star miter offset result: %d paths", len(result))
}

// TestOffsetMiterVariousLimits tests different miter limit values
func TestOffsetMiterVariousLimits(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	limits := []float64{1.5, 2.0, 3.0, 5.0}

	for _, limit := range limits {
		options := OffsetOptions{
			MiterLimit:   limit,
			ArcTolerance: 0.25,
		}

		result, err := InflatePaths64(square, 10.0, JoinMiter, EndPolygon, options)
		if err != nil {
			t.Fatalf("InflatePaths64 with limit %v failed: %v", limit, err)
		}

		if len(result) == 0 {
			t.Fatalf("Expected non-empty result for miter limit %v", limit)
		}

		t.Logf("Miter limit %v result: %d paths", limit, len(result))
	}
}

// TestOffsetMiterAccessorMethods tests the MiterLimit getter/setter
func TestOffsetMiterAccessorMethods(t *testing.T) {
	co := NewClipperOffset(2.0, 0.25)

	// Test getter
	if co.MiterLimit() != 2.0 {
		t.Errorf("Expected miter limit 2.0, got %v", co.MiterLimit())
	}

	// Test setter with valid value
	co.SetMiterLimit(3.5)
	if co.MiterLimit() != 3.5 {
		t.Errorf("Expected miter limit 3.5, got %v", co.MiterLimit())
	}

	// Test setter with invalid value (should clamp to 2.0)
	co.SetMiterLimit(0.5)
	if co.MiterLimit() != 2.0 {
		t.Errorf("Expected miter limit to be clamped to 2.0, got %v", co.MiterLimit())
	}

	// Test ArcTolerance getter/setter
	if co.ArcTolerance() != 0.25 {
		t.Errorf("Expected arc tolerance 0.25, got %v", co.ArcTolerance())
	}

	co.SetArcTolerance(0.5)
	if co.ArcTolerance() != 0.5 {
		t.Errorf("Expected arc tolerance 0.5, got %v", co.ArcTolerance())
	}
}

// TestOffsetMiterConcavePolygon tests miter joins on concave polygons
func TestOffsetMiterConcavePolygon(t *testing.T) {
	// L-shaped polygon (concave)
	lShape := Paths64{{{0, 0}, {60, 0}, {60, 40}, {40, 40}, {40, 100}, {0, 100}}}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	result, err := InflatePaths64(lShape, 5.0, JoinMiter, EndPolygon, options)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from L-shape miter offset")
	}

	t.Logf("L-shape miter offset result: %d paths", len(result))
}

// TestOffsetMiterRectangle tests various rectangles with different aspect ratios
func TestOffsetMiterRectangle(t *testing.T) {
	testCases := []struct {
		name string
		rect Path64
	}{
		{"Wide rectangle", Path64{{0, 0}, {200, 0}, {200, 50}, {0, 50}}},
		{"Tall rectangle", Path64{{0, 0}, {50, 0}, {50, 200}, {0, 200}}},
		{"Small square", Path64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}},
	}

	options := OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(Paths64{tc.rect}, 10.0, JoinMiter, EndPolygon, options)
			if err != nil {
				t.Fatalf("InflatePaths64 failed for %s: %v", tc.name, err)
			}

			if len(result) == 0 {
				t.Fatalf("Expected non-empty result for %s", tc.name)
			}

			t.Logf("%s miter offset result: %d paths", tc.name, len(result))
		})
	}
}
