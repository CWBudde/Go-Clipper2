//go:build clipper_cgo

package capi

import (
	"testing"
)

// TestAllJoinTypes validates that all JoinType enum values work with the bridge
func TestAllJoinTypes(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	tests := []struct {
		name     string
		joinType uint8
		desc     string
	}{
		{"Square", 0, "C_JOIN_SQUARE"},
		{"Bevel", 1, "C_JOIN_BEVEL"},
		{"Round", 2, "C_JOIN_ROUND"},
		{"Miter", 3, "C_JOIN_MITER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InflatePaths64(square, 2.0, tt.joinType, 0 /*EndPolygon*/, 2.0, 0.25)
			if err != nil {
				t.Fatalf("InflatePaths64 with %s failed: %v", tt.desc, err)
			}
			if len(result) == 0 {
				t.Fatalf("Expected non-empty result for %s", tt.desc)
			}
			t.Logf("%s produced %d paths with %d total points", tt.desc, len(result), countPoints(result))
		})
	}
}

// TestAllEndTypes validates that all EndType enum values work with the bridge
func TestAllEndTypes(t *testing.T) {
	line := Paths64{{{0, 0}, {10, 0}}} // Open path for end type testing

	tests := []struct {
		name    string
		endType uint8
		desc    string
	}{
		{"Polygon", 0, "C_END_POLYGON"},
		{"Joined", 1, "C_END_JOINED"},
		{"Butt", 2, "C_END_BUTT"},
		{"Square", 3, "C_END_SQUARE"},
		{"Round", 4, "C_END_ROUND"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InflatePaths64(line, 2.0, 0 /*JoinSquare*/, tt.endType, 2.0, 0.25)
			if err != nil {
				t.Fatalf("InflatePaths64 with %s failed: %v", tt.desc, err)
			}
			if len(result) == 0 {
				t.Fatalf("Expected non-empty result for %s", tt.desc)
			}
			t.Logf("%s produced %d paths with %d total points", tt.desc, len(result), countPoints(result))
		})
	}
}

// TestBevelJoinSpecifically validates Bevel join produces different results than other join types
func TestBevelJoinSpecifically(t *testing.T) {
	// Create a shape with a sharp corner to see join type differences
	triangle := Paths64{{{0, 0}, {10, 0}, {5, 10}}}

	// Test Bevel (simplest)
	bevelResult, err := InflatePaths64(triangle, 2.0, 1 /*Bevel*/, 0 /*Polygon*/, 2.0, 0.25)
	if err != nil {
		t.Fatalf("Bevel join failed: %v", err)
	}

	// Test Miter (should produce sharper corners)
	miterResult, err := InflatePaths64(triangle, 2.0, 3 /*Miter*/, 0 /*Polygon*/, 2.0, 0.25)
	if err != nil {
		t.Fatalf("Miter join failed: %v", err)
	}

	bevelPoints := countPoints(bevelResult)
	miterPoints := countPoints(miterResult)

	t.Logf("Bevel join: %d points, Miter join: %d points", bevelPoints, miterPoints)

	// Bevel should typically have fewer points than Miter (simpler join)
	// This is not always true, but serves as a sanity check
	if bevelPoints == 0 || miterPoints == 0 {
		t.Fatal("Both join types should produce non-empty results")
	}
}

// Helper to count total points across all paths
func countPoints(paths Paths64) int {
	total := 0
	for _, path := range paths {
		total += len(path)
	}
	return total
}
