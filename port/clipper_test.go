package clipper

import (
	"testing"
)

func TestUnion64Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, err := Union64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Union64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Union64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from union")
	}
	t.Logf("Union result: %v", result)
}

func TestIntersect64Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, err := Intersect64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Intersect64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from intersection")
	}
	t.Logf("Intersection result: %v", result)
}

func TestDifference64Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, err := Difference64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Difference64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Difference64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from difference")
	}
	t.Logf("Difference result: %v", result)
}

func TestXor64Basic(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, err := Xor64(subject, clip, NonZero)
	if err == ErrNotImplemented {
		t.Skip("Xor64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("Xor64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from XOR")
	}
	t.Logf("XOR result: %v", result)
}

func TestArea64(t *testing.T) {
	// Simple square: 10x10 = 100
	square := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	area := Area64(square)
	expected := 100.0
	
	if area != expected {
		t.Errorf("Expected area %v, got %v", expected, area)
	}
}

func TestIsPositive64(t *testing.T) {
	// Counter-clockwise square (positive)
	ccwSquare := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	if !IsPositive64(ccwSquare) {
		t.Error("Expected counter-clockwise square to be positive")
	}

	// Clockwise square (negative)
	cwSquare := Path64{{0, 0}, {0, 10}, {10, 10}, {10, 0}}
	if IsPositive64(cwSquare) {
		t.Error("Expected clockwise square to be negative")
	}
}

func TestReverse64(t *testing.T) {
	original := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	expected := Path64{{0, 10}, {10, 10}, {10, 0}, {0, 0}}
	
	result := Reverse64(original)
	
	if len(result) != len(expected) {
		t.Fatalf("Length mismatch: expected %d, got %d", len(expected), len(result))
	}
	
	for i, pt := range result {
		if pt != expected[i] {
			t.Errorf("Point %d: expected %v, got %v", i, expected[i], pt)
		}
	}
}

func TestInflatePaths64(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	
	result, err := InflatePaths64(square, 1.0, Round, ClosedPolygon)
	if err == ErrNotImplemented {
		t.Skip("InflatePaths64 not yet implemented")
	}
	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}
	
	if len(result) == 0 {
		t.Fatal("Expected non-empty result from inflate")
	}
	t.Logf("Inflate result: %v", result)
}