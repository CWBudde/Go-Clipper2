package clipper

import (
	"errors"
	"testing"
)

// ==============================================================================
// Error Validation Tests
// ==============================================================================

func TestBooleanOp64_InvalidClipType(t *testing.T) {
	subjects := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clips := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	// ClipType 100 is out of range (valid: 0-3)
	_, _, err := BooleanOp64(ClipType(100), NonZero, subjects, nil, clips)
	if !errors.Is(err, ErrInvalidClipType) {
		t.Errorf("Expected ErrInvalidClipType, got: %v", err)
	}
}

func TestBooleanOp64_InvalidFillRule(t *testing.T) {
	subjects := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clips := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	// FillRule 100 is out of range (valid: 0-3)
	_, _, err := BooleanOp64(Union, FillRule(100), subjects, nil, clips)
	if !errors.Is(err, ErrInvalidFillRule) {
		t.Errorf("Expected ErrInvalidFillRule, got: %v", err)
	}
}

func TestBooleanOp64_EmptyPaths(t *testing.T) {
	// Empty paths should not cause errors - just return empty result
	result, _, err := BooleanOp64(Union, NonZero, nil, nil, nil)
	if err != nil {
		t.Errorf("Unexpected error for empty paths: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty paths, got %d paths", len(result))
	}
}

func TestBooleanOp64_DegeneratePaths(t *testing.T) {
	// Paths with < 3 points should be filtered out
	subjects := Paths64{
		{{0, 0}, {10, 0}},                    // Only 2 points - invalid
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, // Valid rectangle
	}
	clips := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	// Should succeed and filter out degenerate path
	_, _, err := BooleanOp64(Union, NonZero, subjects, nil, clips)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInflatePaths64_InvalidJoinType(t *testing.T) {
	paths := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	// JoinType 100 is out of range (valid: 0-3)
	_, err := InflatePaths64(paths, 5.0, JoinType(100), EndPolygon)
	if !errors.Is(err, ErrInvalidJoinType) {
		t.Errorf("Expected ErrInvalidJoinType, got: %v", err)
	}
}

func TestInflatePaths64_InvalidEndType(t *testing.T) {
	paths := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	// EndType 100 is out of range (valid: 0-4)
	_, err := InflatePaths64(paths, 5.0, JoinSquare, EndType(100))
	if !errors.Is(err, ErrInvalidEndType) {
		t.Errorf("Expected ErrInvalidEndType, got: %v", err)
	}
}

func TestInflatePaths64_InvalidOptions(t *testing.T) {
	paths := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	tests := []struct {
		name string
		opts OffsetOptions
	}{
		{
			name: "negative miter limit",
			opts: OffsetOptions{MiterLimit: -1.0, ArcTolerance: 0.25},
		},
		{
			name: "zero miter limit",
			opts: OffsetOptions{MiterLimit: 0.0, ArcTolerance: 0.25},
		},
		{
			name: "negative arc tolerance",
			opts: OffsetOptions{MiterLimit: 2.0, ArcTolerance: -0.1},
		},
		{
			name: "zero arc tolerance",
			opts: OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InflatePaths64(paths, 5.0, JoinSquare, EndPolygon, tt.opts)
			if !errors.Is(err, ErrInvalidOptions) {
				t.Errorf("Expected ErrInvalidOptions for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestInflatePaths64_EmptyPaths(t *testing.T) {
	// Empty paths should return empty result, not error
	result, err := InflatePaths64(nil, 5.0, JoinSquare, EndPolygon)
	if err != nil {
		t.Errorf("Unexpected error for empty paths: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty paths, got %d paths", len(result))
	}
}

func TestSimplifyPath64_InvalidEpsilon(t *testing.T) {
	path := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	tests := []struct {
		name    string
		epsilon float64
	}{
		{"zero epsilon", 0.0},
		{"negative epsilon", -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SimplifyPath64(path, tt.epsilon, false)
			if !errors.Is(err, ErrInvalidParameter) {
				t.Errorf("Expected ErrInvalidParameter for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestSimplifyPath64_EmptyPath(t *testing.T) {
	// Empty path should return empty result, not error
	result, err := SimplifyPath64(Path64{}, 1.0, false)
	if err != nil {
		t.Errorf("Unexpected error for empty path: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty path, got %d points", len(result))
	}
}

func TestMinkowskiSum64_EmptyPattern(t *testing.T) {
	path := Path64{{0, 0}, {10, 0}, {10, 10}}

	_, err := MinkowskiSum64(Path64{}, path, false)
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Expected ErrEmptyPath for empty pattern, got: %v", err)
	}
}

func TestMinkowskiSum64_EmptyPath(t *testing.T) {
	pattern := Path64{{0, 0}, {5, 0}, {5, 5}, {0, 5}}

	_, err := MinkowskiSum64(pattern, Path64{}, false)
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Expected ErrEmptyPath for empty path, got: %v", err)
	}
}

func TestMinkowskiDiff64_EmptyInputs(t *testing.T) {
	pattern := Path64{{0, 0}, {5, 0}, {5, 5}, {0, 5}}
	path := Path64{{0, 0}, {10, 0}, {10, 10}}

	// Empty pattern
	_, err := MinkowskiDiff64(Path64{}, path, false)
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Expected ErrEmptyPath for empty pattern, got: %v", err)
	}

	// Empty path
	_, err = MinkowskiDiff64(pattern, Path64{}, false)
	if !errors.Is(err, ErrEmptyPath) {
		t.Errorf("Expected ErrEmptyPath for empty path, got: %v", err)
	}
}

func TestRectClip64_InvalidRectangle(t *testing.T) {
	paths := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}

	tests := []struct {
		name string
		rect Path64
	}{
		{"empty rect", Path64{}},
		{"1 point", Path64{{0, 0}}},
		{"2 points", Path64{{0, 0}, {10, 10}}},
		{"3 points", Path64{{0, 0}, {10, 0}, {10, 10}}},
		{"5 points", Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {5, 5}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RectClip64(tt.rect, paths)
			if !errors.Is(err, ErrInvalidRectangle) {
				t.Errorf("Expected ErrInvalidRectangle for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestRectClip64_EmptyPaths(t *testing.T) {
	rect := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	// Empty paths should return empty result, not error
	result, err := RectClip64(rect, nil)
	if err != nil {
		t.Errorf("Unexpected error for empty paths: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty paths, got %d paths", len(result))
	}
}

// ==============================================================================
// Edge Case Tests
// ==============================================================================

func TestArea64_EmptyPath(t *testing.T) {
	area := Area64(Path64{})
	if area != 0 {
		t.Errorf("Expected area 0 for empty path, got %f", area)
	}
}

func TestIsPositive64_EmptyPath(t *testing.T) {
	result := IsPositive64(Path64{})
	if result {
		t.Error("Expected false for empty path")
	}
}

func TestReverse64_EmptyPath(t *testing.T) {
	result := Reverse64(Path64{})
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty path, got %d points", len(result))
	}
}

func TestReversePaths64_EmptyPaths(t *testing.T) {
	result := ReversePaths64(Paths64{})
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty paths, got %d paths", len(result))
	}
}

func TestBounds64_EmptyPath(t *testing.T) {
	result := Bounds64(Path64{})
	// Should return empty/zero rectangle
	if result.Width() != 0 || result.Height() != 0 {
		t.Errorf("Expected zero-size rectangle for empty path, got width=%d height=%d",
			result.Width(), result.Height())
	}
}

func TestBoundsPaths64_EmptyPaths(t *testing.T) {
	result := BoundsPaths64(Paths64{})
	// Should return empty/zero rectangle
	if result.Width() != 0 || result.Height() != 0 {
		t.Errorf("Expected zero-size rectangle for empty paths, got width=%d height=%d",
			result.Width(), result.Height())
	}
}
