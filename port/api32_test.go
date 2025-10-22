package clipper

import (
	"math"
	"testing"
)

// ==============================================================================
// Basic 32-bit API Tests
// ==============================================================================
// These tests validate that 32-bit API functions work correctly.
// Full test coverage is provided by the 64-bit tests since 32-bit delegates to 64-bit.

func TestUnion32_Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths32{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}

	result, err := Union32(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Union32() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Union32() returned empty result")
	}

	// Area should be 100*100 + 100*100 - 50*50 = 17500
	area := Area32(result[0])
	expectedArea := float64(17500)
	tolerance := 1.0
	if math.Abs(area-expectedArea) > tolerance {
		t.Errorf("Union32() area = %f, want approximately %f", area, expectedArea)
	}
}

func TestIntersect32_Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths32{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}

	result, err := Intersect32(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Intersect32() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Intersect32() returned empty result")
	}

	// Intersection area should be 50*50 = 2500
	area := Area32(result[0])
	expectedArea := float64(2500)
	tolerance := 1.0
	if math.Abs(area-expectedArea) > tolerance {
		t.Errorf("Intersect32() area = %f, want approximately %f", area, expectedArea)
	}
}

func TestDifference32_Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths32{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}

	result, err := Difference32(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Difference32() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Difference32() returned empty result")
	}

	// Difference area should be 100*100 - 50*50 = 7500
	area := Area32(result[0])
	expectedArea := float64(7500)
	tolerance := 1.0
	if math.Abs(area-expectedArea) > tolerance {
		t.Errorf("Difference32() area = %f, want approximately %f", area, expectedArea)
	}
}

func TestXor32_Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths32{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}

	result, err := Xor32(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Xor32() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Xor32() returned %d paths, want 2", len(result))
	}
}

// ==============================================================================
// Utility Function Tests
// ==============================================================================

func TestArea32(t *testing.T) {
	tests := []struct {
		name     string
		path     Path32
		expected float64
	}{
		{"Square 100x100", Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}, 10000.0},
		{"Triangle", Path32{{0, 0}, {100, 0}, {50, 100}}, 5000.0},
		{"Negative (clockwise)", Path32{{0, 0}, {0, 100}, {100, 100}, {100, 0}}, -10000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Area32(tt.path)
			if math.Abs(result-tt.expected) > 0.1 {
				t.Errorf("Area32() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestIsPositive32(t *testing.T) {
	ccw := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}} // counter-clockwise
	cw := Path32{{0, 0}, {0, 100}, {100, 100}, {100, 0}}  // clockwise

	if !IsPositive32(ccw) {
		t.Error("IsPositive32() expected true for counter-clockwise path")
	}

	if IsPositive32(cw) {
		t.Error("IsPositive32() expected false for clockwise path")
	}
}

func TestReverse32(t *testing.T) {
	original := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}
	expected := Path32{{0, 100}, {100, 100}, {100, 0}, {0, 0}}

	result := Reverse32(original)

	if len(result) != len(expected) {
		t.Fatalf("Reverse32() length = %d, want %d", len(result), len(expected))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Reverse32()[%d] = %v, want %v", i, result[i], expected[i])
		}
	}
}

func TestBounds32(t *testing.T) {
	path := Path32{{10, 20}, {100, 30}, {90, 200}, {5, 150}}
	expected := Rect32{Left: 5, Top: 20, Right: 100, Bottom: 200}

	result := Bounds32(path)

	if result != expected {
		t.Errorf("Bounds32() = %+v, want %+v", result, expected)
	}
}

func TestBoundsPaths32(t *testing.T) {
	paths := Paths32{
		{{10, 20}, {100, 30}},
		{{5, 150}, {90, 200}},
	}
	expected := Rect32{Left: 5, Top: 20, Right: 100, Bottom: 200}

	result := BoundsPaths32(paths)

	if result != expected {
		t.Errorf("BoundsPaths32() = %+v, want %+v", result, expected)
	}
}

func TestRectClip32_Basic(t *testing.T) {
	rect := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}
	paths := Paths32{
		{{-10, -10}, {110, -10}, {110, 110}, {-10, 110}}, // extends beyond rect
	}

	result, err := RectClip32(rect, paths)
	if err != nil {
		t.Fatalf("RectClip32() error = %v", err)
	}

	if len(result) == 0 {
		t.Fatal("RectClip32() returned empty result")
	}

	// Result should be clipped to the rectangle
	bounds := Bounds32(result[0])
	if bounds.Left < 0 || bounds.Top < 0 || bounds.Right > 100 || bounds.Bottom > 100 {
		t.Errorf("RectClip32() result bounds %+v extend beyond clipping rectangle", bounds)
	}
}

// ==============================================================================
// PolyTree32 Tests
// ==============================================================================

func TestUnion32Tree_Basic(t *testing.T) {
	// Two separate rectangles
	subject := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths32{
		{{150, 150}, {250, 150}, {250, 250}, {150, 250}},
	}

	tree, openPaths, err := Union32Tree(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Union32Tree() error = %v", err)
	}

	if len(openPaths) != 0 {
		t.Errorf("Union32Tree() openPaths = %d, want 0", len(openPaths))
	}

	if tree.Count() == 0 {
		t.Fatal("Union32Tree() returned empty tree")
	}

	// Should have two separate polygons as children of root
	if tree.Count() != 2 {
		t.Errorf("Union32Tree() expected 2 children, got %d", tree.Count())
	}
}

func TestPolyTreeToPaths32(t *testing.T) {
	tree := NewPolyTree32()
	tree.AddChild(Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	tree.AddChild(Path32{{200, 200}, {300, 200}, {300, 300}, {200, 300}})

	result := PolyTreeToPaths32(tree)

	if len(result) != 2 {
		t.Errorf("PolyTreeToPaths32() returned %d paths, want 2", len(result))
	}
}

// ==============================================================================
// Overflow Detection Tests
// ==============================================================================

func TestUnion32_NoOverflowWithValidInputs(t *testing.T) {
	// Verify operations work correctly with large valid int32 coordinates
	subject := Paths32{
		{{math.MaxInt32 - 100, math.MaxInt32 - 100}, {math.MaxInt32 - 10, math.MaxInt32 - 100},
			{math.MaxInt32 - 10, math.MaxInt32 - 10}, {math.MaxInt32 - 100, math.MaxInt32 - 10}},
	}

	// Union with self should work fine
	result, err := Union32(subject, subject, NonZero)
	if err != nil {
		t.Errorf("Union32() with valid large coordinates returned error: %v", err)
	}
	if len(result) == 0 {
		t.Error("Union32() returned empty result")
	}

	// Result should have same area as input (union with self)
	inputArea := math.Abs(Area32(subject[0]))
	resultArea := math.Abs(Area32(result[0]))
	if math.Abs(inputArea-resultArea) > 1.0 {
		t.Errorf("Union32() area changed: input=%f, result=%f", inputArea, resultArea)
	}
}

// ==============================================================================
// SimplifyPath32 Tests
// ==============================================================================

func TestSimplifyPath32_Basic(t *testing.T) {
	// Path with collinear points
	path := Path32{
		{0, 0}, {25, 0}, {50, 0}, {75, 0}, {100, 0}, // collinear points along bottom
		{100, 100}, {0, 100},
	}

	result, err := SimplifyPath32(path, 1.0, true)
	if err != nil {
		t.Fatalf("SimplifyPath32() error = %v", err)
	}

	// Should remove intermediate collinear points
	if len(result) >= len(path) {
		t.Errorf("SimplifyPath32() did not simplify: got %d points, original had %d", len(result), len(path))
	}
}

// ==============================================================================
// Minkowski Operations Tests
// ==============================================================================

func TestMinkowskiSum32_Basic(t *testing.T) {
	// Simple square pattern and path
	pattern := Path32{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	path := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}

	result, err := MinkowskiSum32(pattern, path, true)
	if err != nil {
		t.Fatalf("MinkowskiSum32() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("MinkowskiSum32() returned empty result")
	}

	// Minkowski sum should produce a larger polygon
	originalArea := Area32(path)
	resultArea := 0.0
	for _, p := range result {
		resultArea += math.Abs(Area32(p))
	}

	if resultArea <= originalArea {
		t.Errorf("MinkowskiSum32() result area %f should be larger than original %f", resultArea, originalArea)
	}
}

func TestMinkowskiDiff32_Basic(t *testing.T) {
	// Simple square pattern and path
	pattern := Path32{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	path := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}

	result, err := MinkowskiDiff32(pattern, path, true)
	if err != nil {
		t.Fatalf("MinkowskiDiff32() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("MinkowskiDiff32() returned empty result")
	}
}
