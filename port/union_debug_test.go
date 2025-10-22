package clipper

import (
	"testing"
)

// TestUnion64BasicDebug tests union with debug output
func TestUnion64BasicDebug(t *testing.T) {
	// Two overlapping rectangles
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	t.Logf("Subject: %v", subject)
	t.Logf("Clip: %v", clip)

	result, err := Union64(subject, clip, NonZero)

	if err != nil {
		t.Fatalf("Union64 failed: %v", err)
	}

	t.Logf("Union result: %v", result)
	t.Logf("Result has %d polygon(s)", len(result))

	// Compare with oracle
	t.Logf("\nExpected (oracle): 1 merged polygon with 8 points")
	t.Logf("Actual (pure Go): %d polygon(s)", len(result))

	for i, path := range result {
		t.Logf("Polygon %d: %d points", i, len(path))
		for j, pt := range path {
			t.Logf("  Point %d: %v", j, pt)
		}
	}
}
