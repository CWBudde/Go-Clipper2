package clipper

import (
	"testing"
)

// TestOffsetDirectSimple tests ClipperOffset directly with a simple square
func TestOffsetDirectSimple(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	co := NewClipperOffset(2.0, 0.25)
	t.Log("Created ClipperOffset")

	co.AddPaths(square, JoinBevel, EndPolygon)
	t.Logf("Added paths, groups: %d", len(co.groups))

	t.Log("Calling DoGroupOffset directly...")
	if len(co.groups) > 0 {
		co.delta = 2.0
		result := co.DoGroupOffset(&co.groups[0])
		t.Logf("DoGroupOffset result: %d paths", len(result))
		for i, path := range result {
			t.Logf("  Path %d: %v", i, path)
		}

		// Now test the Union operation
		t.Log("Testing Union on offset result...")
		cleanedSolution, _, err := booleanOp64Impl(Union, Positive, result, nil, nil)
		if err != nil {
			t.Fatalf("Union failed: %v", err)
		}
		t.Logf("Union result: %d paths", len(cleanedSolution))
	}
}
