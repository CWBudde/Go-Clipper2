//go:build clipper_cgo

package capi

import "testing"

func TestUnionTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64( /*Union*/ 1 /*NonZero*/, 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected merged polygon, got empty")
	}
	t.Logf("Union result: %v", got)
}

func TestIntersectionTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64( /*Intersection*/ 0 /*NonZero*/, 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected intersection polygon, got empty")
	}
	t.Logf("Intersection result: %v", got)
}

func TestDifferenceTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64( /*Difference*/ 2 /*NonZero*/, 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected difference polygon, got empty")
	}
	t.Logf("Difference result: %v", got)
}

func TestXorTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64( /*Xor*/ 3 /*NonZero*/, 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected xor polygon, got empty")
	}
	t.Logf("Xor result: %v", got)
}

func TestEmptyInputs(t *testing.T) {
	empty := Paths64{}
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	got, open, err := BooleanOp64( /*Union*/ 1 /*NonZero*/, 1, a, nil, empty)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 path for union with empty, got %d", len(got))
	}
	t.Logf("Union with empty result: %v", got)
}

func TestInternalPackUnpack(t *testing.T) {
	// Test that our internal pack/unpack functions work correctly
	original := Paths64{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
		{{20, 20}, {30, 20}, {30, 30}, {20, 30}},
	}
	cp, cleanup := toCPaths64(original)
	defer cleanup()
	round := fromCPaths64(&cp)
	if len(round) != len(original) {
		t.Fatalf("round trip path count mismatch: got %d want %d", len(round), len(original))
	}
	for i := range original {
		if len(round[i]) != len(original[i]) {
			t.Fatalf("path %d length mismatch", i)
		}
		for j := range original[i] {
			if round[i][j] != original[i][j] {
				t.Fatalf("path %d point %d mismatch: got %v want %v", i, j, round[i][j], original[i][j])
			}
		}
	}
}

func TestBooleanOp64OpenPaths(t *testing.T) {
	subj := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	open := Paths64{{{-5, 5}, {15, 5}}} // Open path that crosses outside the subject polygon
	sol, solOpen, err := BooleanOp64( /*Union*/ 1 /*NonZero*/, 1, subj, open, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sol) != 1 {
		t.Fatalf("expected 1 closed path, got %d", len(sol))
	}
	if len(solOpen) == 0 {
		t.Fatalf("expected at least 1 open path, got 0")
	}
	// The open path crossing the polygon may be split into multiple segments
	t.Logf("Got %d open path(s): %v", len(solOpen), solOpen)

	// Verify all open paths have at least 2 points
	for i, path := range solOpen {
		if len(path) < 2 {
			t.Fatalf("open path %d should have at least 2 points, got %d", i, len(path))
		}
	}
}

// area is a helper to calculate polygon signed area.
func area(path Path64) int64 {
	if len(path) < 3 {
		return 0
	}
	sum := int64(0)
	for i := 0; i < len(path); i++ {
		j := (i + 1) % len(path)
		sum += path[i].X*path[j].Y - path[j].X*path[i].Y
	}
	return sum / 2
}

func TestBooleanOp64WithHole(t *testing.T) {
	outer := Path64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}
	hole := Path64{{15, 5}, {5, 5}, {5, 15}, {15, 15}}
	subj := Paths64{outer, hole}
	sol, open, err := BooleanOp64( /*Union*/ 1 /*NonZero*/, 1, subj, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(sol) != 2 {
		t.Fatalf("expected 2 paths (outer+hole), got %d", len(sol))
	}
	// verify outer is positive orientation and hole negative
	if area(sol[0])*area(sol[1]) > 0 {
		t.Fatalf("expected opposite winding for outer and hole")
	}
}

func TestBooleanOp64IntersectionCoords(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64( /*Intersection*/ 0 /*NonZero*/, 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 path, got %d", len(got))
	}

	// Check that all expected vertices are present (regardless of order/starting point)
	expect := map[Point64]bool{
		{5, 5}: true, {10, 5}: true, {10, 10}: true, {5, 10}: true,
	}
	if len(got[0]) != len(expect) {
		t.Fatalf("intersection vertex count mismatch: got %d want %d", len(got[0]), len(expect))
	}
	for i, vertex := range got[0] {
		if !expect[vertex] {
			t.Fatalf("unexpected vertex %d: got %v", i, vertex)
		}
	}

	// Verify it's the correct intersection area (should be 25 square units)
	expectedArea := int64(25)
	actualArea := abs(area(got[0]))
	if actualArea != expectedArea {
		t.Fatalf("intersection area mismatch: got %d want %d", actualArea, expectedArea)
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestInflatePaths64(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	inflated, err := InflatePaths64(square, 2 /*Miter*/, 2 /*ClosedPolygon*/, 0, 2.0, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if len(inflated) == 0 {
		t.Fatalf("expected non-empty result from inflate")
	}
}

func TestRectClip64(t *testing.T) {
	rect := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	paths := Paths64{{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}}
	clipped, err := RectClip64(rect, paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(clipped) == 0 {
		t.Fatalf("expected non-empty clipped result")
	}
}
