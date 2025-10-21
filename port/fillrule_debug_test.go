package clipper

import (
	"bytes"
	"testing"
)

func TestPositiveFillRuleDebug(t *testing.T) {
	// Enable debug for this test
	VattiDebug = true
	debugBuf := &bytes.Buffer{}
	VattiDebugOutput = debugBuf
	defer func() {
		VattiDebug = false
		VattiDebugOutput = nil
	}()

	// Two overlapping rectangles (CCW orientation)
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	// Check orientation
	subjArea := Area64(subject[0])
	clipArea := Area64(clip[0])
	t.Logf("Subject area: %.1f (CCW=%v)", subjArea, IsPositive64(subject[0]))
	t.Logf("Clip area: %.1f (CCW=%v)", clipArea, IsPositive64(clip[0]))

	result, err := Intersect64(subject, clip, Positive)

	// Print debug output
	t.Log("\n=== DEBUG LOG ===\n" + debugBuf.String() + "\n=== END DEBUG LOG ===")

	if err != nil {
		t.Fatalf("Intersect64 failed: %v", err)
	}

	t.Logf("Positive fill rule result: %v", result)

	if len(result) == 0 {
		t.Error("Expected non-empty result for Positive fill rule")
	}

	// Try with CW orientation (reversed polygons)
	subjectCW := Paths64{Reverse64(subject[0])}
	clipCW := Paths64{Reverse64(clip[0])}

	subjAreaCW := Area64(subjectCW[0])
	clipAreaCW := Area64(clipCW[0])
	t.Logf("Subject CW area: %.1f (CCW=%v)", subjAreaCW, IsPositive64(subjectCW[0]))
	t.Logf("Clip CW area: %.1f (CCW=%v)", clipAreaCW, IsPositive64(clipCW[0]))

	resultCW, err := Intersect64(subjectCW, clipCW, Positive)
	if err != nil {
		t.Fatalf("Intersect64 with CW failed: %v", err)
	}

	t.Logf("Positive fill rule result (CW): %v", resultCW)
}
