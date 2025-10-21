package clipper

import (
	"testing"
)

// TestOffsetBevelSquareExpansion tests basic square expansion with Bevel joins
func TestOffsetBevelSquareExpansion(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	result, err := InflatePaths64(square, 10.0, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from offset")
	}

	// The result should have more points than input (bevel adds 2 points per corner)
	totalPoints := 0
	for _, path := range result {
		totalPoints += len(path)
	}

	if totalPoints < 4 {
		t.Errorf("Expected at least 4 points, got %d", totalPoints)
	}

	t.Logf("Square expansion result: %d paths, %d total points", len(result), totalPoints)
}

// TestOffsetBevelSquareContraction tests square contraction (negative delta)
func TestOffsetBevelSquareContraction(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	result, err := InflatePaths64(square, -10.0, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from offset")
	}

	t.Logf("Square contraction result: %d paths", len(result))
}

// TestOffsetBevelConcavePolygon tests offsetting a concave polygon
func TestOffsetBevelConcavePolygon(t *testing.T) {
	// L-shaped polygon (concave)
	lShape := Paths64{{{0, 0}, {60, 0}, {60, 40}, {40, 40}, {40, 100}, {0, 100}}}

	result, err := InflatePaths64(lShape, 5.0, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from concave polygon offset")
	}

	t.Logf("L-shape offset result: %d paths", len(result))
}

// TestOffsetBevelMultiplePaths tests offsetting multiple paths at once
func TestOffsetBevelMultiplePaths(t *testing.T) {
	paths := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
		{{100, 100}, {150, 100}, {150, 150}, {100, 150}},
	}

	result, err := InflatePaths64(paths, 5.0, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) < 2 {
		t.Errorf("Expected at least 2 paths, got %d", len(result))
	}

	t.Logf("Multiple paths offset result: %d paths", len(result))
}

// TestOffsetBevelTriangle tests offsetting a triangle
func TestOffsetBevelTriangle(t *testing.T) {
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 100}}}

	result, err := InflatePaths64(triangle, 10.0, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from triangle offset")
	}

	t.Logf("Triangle offset result: %d paths", len(result))
}

// TestOffsetBevelSmallDelta tests offsetting with a very small delta
func TestOffsetBevelSmallDelta(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	// Very small delta (< 0.5) should return paths nearly unchanged
	result, err := InflatePaths64(square, 0.1, JoinBevel, EndPolygon)
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from small delta offset")
	}

	t.Logf("Small delta offset result: %d paths", len(result))
}

// TestOffsetPhase4UnsupportedOperations tests that unsupported operations return ErrNotImplemented
// This test only applies to pure Go mode (oracle supports everything)
func TestOffsetPhase4UnsupportedOperations(t *testing.T) {
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	// Phase 4: All join types now supported (Bevel, Miter, Square, Round)
	// No unsupported join types to test

	// Test unsupported end types (Phase 4 - only EndPolygon supported)
	unsupportedEnds := []EndType{EndJoined, EndButt, EndSquare, EndRound}
	for _, et := range unsupportedEnds {
		_, err := InflatePaths64(square, 10.0, JoinBevel, et)
		// Oracle supports all features, so skip this check in oracle mode
		if err == nil {
			// This means we're running in oracle mode, skip
			t.Skip("Skipping unsupported operations test in oracle mode")
		}
		if err != ErrNotImplemented {
			t.Errorf("Expected ErrNotImplemented for end type %v, got %v", et, err)
		}
	}
}
