package clipper

import (
	"bytes"
	"testing"
)

func TestIntersect64BasicWithDebug(t *testing.T) {
	// Enable debug logging
	VattiDebug = true
	debugBuf := &bytes.Buffer{}
	VattiDebugOutput = debugBuf
	defer func() {
		VattiDebug = false
		VattiDebugOutput = nil
	}()

	// Two overlapping rectangles
	// Subject: (0,0) to (10,10)
	// Clip: (5,5) to (15,15)
	// Expected intersection: (5,5) to (10,10)
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	t.Logf("Subject: %v", subject)
	t.Logf("Clip: %v", clip)

	result, err := Intersect64(subject, clip, NonZero)

	// Always print debug log, even on error
	t.Log("\n=== DEBUG LOG ===\n" + debugBuf.String() + "\n=== END DEBUG LOG ===")

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

	// Print debug log
	t.Log("\n" + debugBuf.String())

	// Validate the result
	// The intersection should be a square from (5,5) to (10,10)
	// with 4 vertices forming a proper polygon
	if len(result) != 1 {
		t.Errorf("Expected 1 polygon in result, got %d", len(result))
	}

	if len(result[0]) != 4 {
		t.Errorf("Expected 4 vertices in intersection polygon, got %d", len(result[0]))
	}

	// Check that all points are within the expected bounds
	for i, pt := range result[0] {
		if pt.X < 5 || pt.X > 10 || pt.Y < 5 || pt.Y > 10 {
			t.Errorf("Point %d (%v) is outside expected intersection bounds (5,5) to (10,10)", i, pt)
		}
		t.Logf("Result point %d: %v", i, pt)
	}

	// Check the area of the result
	area := Area64(result[0])
	expectedArea := 25.0 // 5x5 square
	if area != expectedArea {
		t.Logf("Warning: Expected area %v, got %v", expectedArea, area)
	}
}
