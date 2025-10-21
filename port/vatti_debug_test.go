package clipper

import (
	"testing"
	"time"
)

// TestUnionOffsetPolygon tests Union operation on the exact polygon produced by offset
func TestUnionOffsetPolygon(t *testing.T) {
	// This is the exact polygon produced by DoGroupOffset for a 10x10 square with delta=2
	offsetResult := Paths64{{
		{-2, 0}, {0, 0}, {0, -2}, {10, -2}, {10, 0}, {12, 0},
		{12, 10}, {10, 10}, {10, 12}, {0, 12}, {0, 10}, {-2, 10},
	}}

	t.Logf("Testing Union on offset polygon with %d points", len(offsetResult[0]))

	// Run Union in a goroutine with timeout
	done := make(chan bool)
	var result Paths64
	var err error

	go func() {
		result, _, err = booleanOp64Impl(Union, Positive, offsetResult, nil, nil)
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("Union failed: %v", err)
		}
		t.Logf("Union succeeded: %d paths", len(result))
		for i, path := range result {
			t.Logf("  Path %d: %d points", i, len(path))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Union operation timed out (likely infinite loop)")
	}
}

// TestUnionSimpleSquare tests Union on a simple square (should work)
func TestUnionSimpleSquare(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	t.Log("Testing Union on simple square...")
	result, _, err := booleanOp64Impl(Union, Positive, square, nil, nil)
	if err != nil {
		t.Fatalf("Union failed: %v", err)
	}
	t.Logf("Union succeeded: %d paths", len(result))
}

// TestUnionOffsetPolygonCCW tests if orientation matters
func TestUnionOffsetPolygonCCW(t *testing.T) {
	// Try the offset polygon in CCW orientation
	offsetResultCCW := Paths64{{
		{-2, 10}, {0, 10}, {0, 12}, {10, 12}, {10, 10}, {12, 10},
		{12, 0}, {10, 0}, {10, -2}, {0, -2}, {0, 0}, {-2, 0},
	}}

	t.Logf("Testing Union on CCW offset polygon")

	done := make(chan bool)
	var result Paths64
	var err error

	go func() {
		result, _, err = booleanOp64Impl(Union, Positive, offsetResultCCW, nil, nil)
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("Union failed: %v", err)
		}
		t.Logf("Union succeeded: %d paths", len(result))
	case <-time.After(2 * time.Second):
		t.Fatal("Union operation timed out (likely infinite loop)")
	}
}
