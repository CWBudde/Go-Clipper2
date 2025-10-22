package clipper

import (
	"testing"
)

// TestRectClipLinesBasicInside tests a line completely inside the rectangle
func TestRectClipLinesBasicInside(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line completely inside rectangle
	line := Path64{
		{X: 150, Y: 150},
		{X: 350, Y: 250},
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return the line unchanged
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// Check points match
	if result[0][0].X != 150 || result[0][0].Y != 150 {
		t.Errorf("First point mismatch: got (%d,%d), want (150,150)", result[0][0].X, result[0][0].Y)
	}
	if result[0][1].X != 350 || result[0][1].Y != 250 {
		t.Errorf("Second point mismatch: got (%d,%d), want (350,250)", result[0][1].X, result[0][1].Y)
	}
}

// TestRectClipLinesBasicOutside tests a line completely outside the rectangle
func TestRectClipLinesBasicOutside(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line completely outside rectangle (to the left)
	line := Path64{
		{X: 0, Y: 150},
		{X: 50, Y: 250},
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return empty result
	if len(result) != 0 {
		t.Fatalf("Expected 0 paths (line outside rect), got %d", len(result))
	}
}

// TestRectClipLinesSingleCrossing tests a line crossing one boundary
func TestRectClipLinesSingleCrossing(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line crossing left boundary (entering from outside)
	line := Path64{
		{X: 50, Y: 200},  // Outside
		{X: 300, Y: 200}, // Inside
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return clipped segment starting at left boundary
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// First point should be on left boundary (X=100, Y=200)
	if result[0][0].X != 100 {
		t.Errorf("Expected clipped point at X=100, got X=%d", result[0][0].X)
	}
	// Y should be approximately 200 (might have small rounding)
	if result[0][0].Y < 199 || result[0][0].Y > 201 {
		t.Errorf("Expected clipped point at Y≈200, got Y=%d", result[0][0].Y)
	}

	// Second point should be the interior point
	if result[0][1].X != 300 || result[0][1].Y != 200 {
		t.Errorf("Second point mismatch: got (%d,%d), want (300,200)", result[0][1].X, result[0][1].Y)
	}
}

// TestRectClipLinesDoubleCrossing tests a line crossing two boundaries
func TestRectClipLinesDoubleCrossing(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line passing completely through rectangle (left to right)
	line := Path64{
		{X: 0, Y: 200},   // Outside left
		{X: 500, Y: 200}, // Outside right
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return one segment clipped to rectangle
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// First point should be on left boundary
	if result[0][0].X != 100 {
		t.Errorf("Expected first point at X=100, got X=%d", result[0][0].X)
	}

	// Second point should be on right boundary
	if result[0][1].X != 400 {
		t.Errorf("Expected second point at X=400, got X=%d", result[0][1].X)
	}

	// Both Y coordinates should be approximately 200
	if result[0][0].Y < 199 || result[0][0].Y > 201 {
		t.Errorf("Expected first point at Y≈200, got Y=%d", result[0][0].Y)
	}
	if result[0][1].Y < 199 || result[0][1].Y > 201 {
		t.Errorf("Expected second point at Y≈200, got Y=%d", result[0][1].Y)
	}
}

// TestRectClipLinesMultipleSegments tests a polyline that creates multiple clipped segments
func TestRectClipLinesMultipleSegments(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Zigzag line that enters and exits rectangle multiple times
	line := Path64{
		{X: 50, Y: 150},  // Outside
		{X: 200, Y: 200}, // Inside
		{X: 50, Y: 250},  // Outside
		{X: 300, Y: 200}, // Inside
		{X: 450, Y: 150}, // Outside
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return multiple segments (exact number depends on algorithm)
	// At minimum, we should have some clipped segments
	if len(result) == 0 {
		t.Fatal("Expected at least 1 clipped segment, got 0")
	}

	// All points should be inside or on the rectangle boundary
	for i, path := range result {
		for j, pt := range path {
			if pt.X < 100 || pt.X > 400 || pt.Y < 100 || pt.Y > 300 {
				t.Errorf("Point [%d][%d] (%d,%d) is outside rectangle bounds", i, j, pt.X, pt.Y)
			}
		}
	}
}

// TestRectClipLinesVerticalLine tests vertical line clipping
func TestRectClipLinesVerticalLine(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Vertical line passing through rectangle
	line := Path64{
		{X: 250, Y: 50},  // Above rectangle
		{X: 250, Y: 350}, // Below rectangle
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return one segment clipped to rectangle
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// Both points should have X=250
	if result[0][0].X != 250 || result[0][1].X != 250 {
		t.Errorf("Expected vertical line at X=250, got X=(%d,%d)", result[0][0].X, result[0][1].X)
	}

	// First point should be at top boundary, second at bottom
	if result[0][0].Y != 100 {
		t.Errorf("Expected first point at Y=100, got Y=%d", result[0][0].Y)
	}
	if result[0][1].Y != 300 {
		t.Errorf("Expected second point at Y=300, got Y=%d", result[0][1].Y)
	}
}

// TestRectClipLinesHorizontalLine tests horizontal line clipping
func TestRectClipLinesHorizontalLine(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Horizontal line passing through rectangle
	line := Path64{
		{X: 50, Y: 200},  // Left of rectangle
		{X: 450, Y: 200}, // Right of rectangle
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return one segment clipped to rectangle
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// Both points should have Y=200
	if result[0][0].Y != 200 || result[0][1].Y != 200 {
		t.Errorf("Expected horizontal line at Y=200, got Y=(%d,%d)", result[0][0].Y, result[0][1].Y)
	}

	// First point should be at left boundary, second at right
	if result[0][0].X != 100 {
		t.Errorf("Expected first point at X=100, got X=%d", result[0][0].X)
	}
	if result[0][1].X != 400 {
		t.Errorf("Expected second point at X=400, got X=%d", result[0][1].X)
	}
}

// TestRectClipLinesOnBoundary tests a line that lies on rectangle boundary
func TestRectClipLinesOnBoundary(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line on left boundary
	line := Path64{
		{X: 100, Y: 150},
		{X: 100, Y: 250},
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Line on boundary should be included
	if len(result) != 1 {
		t.Fatalf("Expected 1 path (line on boundary), got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}
}

// TestRectClipLinesEmptyInput tests empty input handling
func TestRectClipLinesEmptyInput(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Empty paths
	paths := Paths64{}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return empty result
	if len(result) != 0 {
		t.Fatalf("Expected 0 paths (empty input), got %d", len(result))
	}
}

// TestRectClipLinesNilInput tests nil input handling
func TestRectClipLinesNilInput(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	result, err := RectClipLines64(rect, nil)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return empty result
	if len(result) != 0 {
		t.Fatalf("Expected 0 paths (nil input), got %d", len(result))
	}
}

// TestRectClipLinesInvalidRectangle tests error handling for invalid rectangle
func TestRectClipLinesInvalidRectangle(t *testing.T) {
	// Rectangle with only 3 points (invalid)
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
	}

	line := Path64{
		{X: 150, Y: 150},
		{X: 350, Y: 250},
	}
	paths := Paths64{line}

	_, err := RectClipLines64(rect, paths)
	if err != ErrInvalidRectangle {
		t.Fatalf("Expected ErrInvalidRectangle, got %v", err)
	}
}

// TestRectClipLinesDiagonalCrossing tests diagonal lines crossing corners
func TestRectClipLinesDiagonalCrossing(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Diagonal line from bottom-left corner area to top-right corner area
	line := Path64{
		{X: 50, Y: 350}, // Outside bottom-left
		{X: 450, Y: 50}, // Outside top-right
	}
	paths := Paths64{line}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return clipped segment
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// Points should be on rectangle boundaries
	for i, pt := range result[0] {
		if pt.X < 100 || pt.X > 400 || pt.Y < 100 || pt.Y > 300 {
			t.Errorf("Point [%d] (%d,%d) is outside rectangle bounds", i, pt.X, pt.Y)
		}
	}
}

// TestRectClipLinesMultiplePaths tests clipping multiple independent paths
func TestRectClipLinesMultiplePaths(t *testing.T) {
	rect := Path64{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	paths := Paths64{
		// Path 1: completely inside
		{{X: 150, Y: 150}, {X: 200, Y: 200}},
		// Path 2: completely outside
		{{X: 0, Y: 0}, {X: 50, Y: 50}},
		// Path 3: crossing boundary
		{{X: 50, Y: 200}, {X: 300, Y: 200}},
	}

	result, err := RectClipLines64(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines64 failed: %v", err)
	}

	// Should return at least 2 paths (path 1 unchanged, path 3 clipped)
	// Path 2 should be excluded
	if len(result) < 2 {
		t.Fatalf("Expected at least 2 paths, got %d", len(result))
	}

	// All points should be inside or on the rectangle boundary
	for i, path := range result {
		for j, pt := range path {
			if pt.X < 100 || pt.X > 400 || pt.Y < 100 || pt.Y > 300 {
				t.Errorf("Point [%d][%d] (%d,%d) is outside rectangle bounds", i, j, pt.X, pt.Y)
			}
		}
	}
}

// ==============================================================================
// 32-bit Coordinate Tests
// ==============================================================================

// TestRectClipLines32BasicInside tests 32-bit coordinate version
func TestRectClipLines32BasicInside(t *testing.T) {
	rect := Path32{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line completely inside rectangle
	line := Path32{
		{X: 150, Y: 150},
		{X: 350, Y: 250},
	}
	paths := Paths32{line}

	result, err := RectClipLines32(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines32 failed: %v", err)
	}

	// Should return the line unchanged
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// Check points match
	if result[0][0].X != 150 || result[0][0].Y != 150 {
		t.Errorf("First point mismatch: got (%d,%d), want (150,150)", result[0][0].X, result[0][0].Y)
	}
	if result[0][1].X != 350 || result[0][1].Y != 250 {
		t.Errorf("Second point mismatch: got (%d,%d), want (350,250)", result[0][1].X, result[0][1].Y)
	}
}

// TestRectClipLines32DoubleCrossing tests 32-bit version with crossing
func TestRectClipLines32DoubleCrossing(t *testing.T) {
	rect := Path32{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
		{X: 100, Y: 300},
	}

	// Line passing completely through rectangle (left to right)
	line := Path32{
		{X: 0, Y: 200},   // Outside left
		{X: 500, Y: 200}, // Outside right
	}
	paths := Paths32{line}

	result, err := RectClipLines32(rect, paths)
	if err != nil {
		t.Fatalf("RectClipLines32 failed: %v", err)
	}

	// Should return one segment clipped to rectangle
	if len(result) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(result[0]))
	}

	// First point should be on left boundary
	if result[0][0].X != 100 {
		t.Errorf("Expected first point at X=100, got X=%d", result[0][0].X)
	}

	// Second point should be on right boundary
	if result[0][1].X != 400 {
		t.Errorf("Expected second point at X=400, got X=%d", result[0][1].X)
	}
}

// TestRectClipLines32InvalidRectangle tests error handling
func TestRectClipLines32InvalidRectangle(t *testing.T) {
	// Rectangle with only 3 points (invalid)
	rect := Path32{
		{X: 100, Y: 100},
		{X: 400, Y: 100},
		{X: 400, Y: 300},
	}

	line := Path32{
		{X: 150, Y: 150},
		{X: 350, Y: 250},
	}
	paths := Paths32{line}

	_, err := RectClipLines32(rect, paths)
	if err != ErrInvalidRectangle {
		t.Fatalf("Expected ErrInvalidRectangle, got %v", err)
	}
}
