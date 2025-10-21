package clipper

import (
	"testing"
)

// ==============================================================================
// Rect64 Tests
// ==============================================================================

func TestRect64Width(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}
	if got := rect.Width(); got != 40 {
		t.Errorf("Width() = %d, want 40", got)
	}
}

func TestRect64Height(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}
	if got := rect.Height(); got != 60 {
		t.Errorf("Height() = %d, want 60", got)
	}
}

func TestRect64MidPoint(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}
	mid := rect.MidPoint()
	expected := Point64{X: 30, Y: 50}
	if mid != expected {
		t.Errorf("MidPoint() = %v, want %v", mid, expected)
	}
}

func TestRect64AsPath(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}
	path := rect.AsPath()
	expected := Path64{
		{X: 10, Y: 20},
		{X: 50, Y: 20},
		{X: 50, Y: 80},
		{X: 10, Y: 80},
	}
	if len(path) != 4 {
		t.Fatalf("AsPath() length = %d, want 4", len(path))
	}
	for i, pt := range path {
		if pt != expected[i] {
			t.Errorf("AsPath()[%d] = %v, want %v", i, pt, expected[i])
		}
	}
}

func TestRect64Contains(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}

	tests := []struct {
		pt   Point64
		want bool
	}{
		{Point64{30, 40}, true},  // Inside
		{Point64{10, 40}, false}, // On left edge (exclusive)
		{Point64{50, 40}, false}, // On right edge (exclusive)
		{Point64{30, 20}, false}, // On top edge (exclusive)
		{Point64{30, 80}, false}, // On bottom edge (exclusive)
		{Point64{5, 40}, false},  // Outside left
		{Point64{60, 40}, false}, // Outside right
		{Point64{30, 10}, false}, // Outside top
		{Point64{30, 90}, false}, // Outside bottom
	}

	for _, tt := range tests {
		if got := rect.Contains(tt.pt); got != tt.want {
			t.Errorf("Contains(%v) = %v, want %v", tt.pt, got, tt.want)
		}
	}
}

func TestRect64ContainsRect(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}

	tests := []struct {
		other Rect64
		want  bool
	}{
		{Rect64{20, 30, 40, 70}, true},  // Fully inside
		{Rect64{10, 20, 50, 80}, true},  // Exact match (inclusive)
		{Rect64{5, 30, 40, 70}, false},  // Extends left
		{Rect64{20, 30, 60, 70}, false}, // Extends right
		{Rect64{20, 10, 40, 70}, false}, // Extends top
		{Rect64{20, 30, 40, 90}, false}, // Extends bottom
	}

	for _, tt := range tests {
		if got := rect.ContainsRect(tt.other); got != tt.want {
			t.Errorf("ContainsRect(%v) = %v, want %v", tt.other, got, tt.want)
		}
	}
}

func TestRect64IsEmpty(t *testing.T) {
	tests := []struct {
		rect Rect64
		want bool
	}{
		{Rect64{10, 20, 50, 80}, false}, // Normal rectangle
		{Rect64{10, 20, 10, 80}, true},  // Zero width
		{Rect64{10, 20, 50, 20}, true},  // Zero height
		{Rect64{50, 20, 10, 80}, true},  // Negative width
		{Rect64{10, 80, 50, 20}, true},  // Negative height
	}

	for _, tt := range tests {
		if got := tt.rect.IsEmpty(); got != tt.want {
			t.Errorf("IsEmpty(%v) = %v, want %v", tt.rect, got, tt.want)
		}
	}
}

func TestRect64Intersects(t *testing.T) {
	rect := Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80}

	tests := []struct {
		other Rect64
		want  bool
	}{
		{Rect64{30, 40, 70, 100}, true},  // Overlapping
		{Rect64{5, 10, 15, 30}, true},    // Overlapping corner
		{Rect64{60, 40, 100, 60}, false}, // No overlap (right)
		{Rect64{5, 90, 15, 100}, false},  // No overlap (below)
		{Rect64{10, 20, 50, 80}, true},   // Exact match
		{Rect64{50, 80, 60, 90}, true},   // Touching corner (the C++ version uses <= so this intersects)
	}

	for _, tt := range tests {
		if got := rect.Intersects(tt.other); got != tt.want {
			t.Errorf("Intersects(%v) = %v, want %v", tt.other, got, tt.want)
		}
	}
}

// ==============================================================================
// Bounds64 Tests
// ==============================================================================

func TestBounds64(t *testing.T) {
	tests := []struct {
		name string
		path Path64
		want Rect64
	}{
		{
			name: "Empty path",
			path: Path64{},
			want: Rect64{},
		},
		{
			name: "Single point",
			path: Path64{{X: 10, Y: 20}},
			want: Rect64{Left: 10, Top: 20, Right: 10, Bottom: 20},
		},
		{
			name: "Rectangle",
			path: Path64{
				{X: 10, Y: 20},
				{X: 50, Y: 20},
				{X: 50, Y: 80},
				{X: 10, Y: 80},
			},
			want: Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80},
		},
		{
			name: "Complex polygon",
			path: Path64{
				{X: 25, Y: 35},
				{X: 75, Y: 15},
				{X: 100, Y: 60},
				{X: 50, Y: 90},
				{X: 5, Y: 70},
			},
			want: Rect64{Left: 5, Top: 15, Right: 100, Bottom: 90},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Bounds64(tt.path)
			if got != tt.want {
				t.Errorf("Bounds64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoundsPaths64(t *testing.T) {
	tests := []struct {
		name  string
		paths Paths64
		want  Rect64
	}{
		{
			name:  "Empty paths",
			paths: Paths64{},
			want:  Rect64{},
		},
		{
			name: "Single path",
			paths: Paths64{
				{{X: 10, Y: 20}, {X: 50, Y: 80}},
			},
			want: Rect64{Left: 10, Top: 20, Right: 50, Bottom: 80},
		},
		{
			name: "Multiple paths",
			paths: Paths64{
				{{X: 10, Y: 20}, {X: 50, Y: 60}},
				{{X: 30, Y: 40}, {X: 70, Y: 80}},
			},
			want: Rect64{Left: 10, Top: 20, Right: 70, Bottom: 80},
		},
		{
			name: "Paths with gaps",
			paths: Paths64{
				{{X: 0, Y: 0}, {X: 10, Y: 10}},
				{{X: 100, Y: 100}, {X: 110, Y: 110}},
			},
			want: Rect64{Left: 0, Top: 0, Right: 110, Bottom: 110},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoundsPaths64(tt.paths)
			if got != tt.want {
				t.Errorf("BoundsPaths64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==============================================================================
// ReversePaths64 Tests
// ==============================================================================

func TestReversePaths64(t *testing.T) {
	paths := Paths64{
		{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
		{{X: 20, Y: 20}, {X: 30, Y: 20}, {X: 30, Y: 30}},
	}

	reversed := ReversePaths64(paths)

	// Check first path is reversed
	expected1 := Path64{{X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}}
	if len(reversed[0]) != len(expected1) {
		t.Fatalf("ReversePaths64()[0] length = %d, want %d", len(reversed[0]), len(expected1))
	}
	for i, pt := range reversed[0] {
		if pt != expected1[i] {
			t.Errorf("ReversePaths64()[0][%d] = %v, want %v", i, pt, expected1[i])
		}
	}

	// Check second path is reversed
	expected2 := Path64{{X: 30, Y: 30}, {X: 30, Y: 20}, {X: 20, Y: 20}}
	if len(reversed[1]) != len(expected2) {
		t.Fatalf("ReversePaths64()[1] length = %d, want %d", len(reversed[1]), len(expected2))
	}
	for i, pt := range reversed[1] {
		if pt != expected2[i] {
			t.Errorf("ReversePaths64()[1][%d] = %v, want %v", i, pt, expected2[i])
		}
	}
}

// ==============================================================================
// PointInPolygon64 Tests
// ==============================================================================

func TestPointInPolygon64(t *testing.T) {
	// Square polygon
	square := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
		{X: 100, Y: 100},
		{X: 0, Y: 100},
	}

	tests := []struct {
		name     string
		pt       Point64
		fillRule FillRule
		want     PolygonLocation
	}{
		{"Inside", Point64{50, 50}, NonZero, Inside},
		{"Outside", Point64{150, 50}, NonZero, Outside},
		{"On edge", Point64{50, 0}, NonZero, OnBoundary},
		{"On vertex", Point64{0, 0}, NonZero, OnBoundary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointInPolygon64(tt.pt, square, tt.fillRule)
			if got != tt.want {
				t.Errorf("PointInPolygon64(%v) = %v, want %v", tt.pt, got, tt.want)
			}
		})
	}
}

// ==============================================================================
// SimplifyPath64 Tests
// ==============================================================================

func TestSimplifyPath64Collinear(t *testing.T) {
	// Path with collinear points that should be removed
	path := Path64{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 20, Y: 0}, // Collinear - should be removed with small epsilon
		{X: 30, Y: 0},
		{X: 30, Y: 10},
	}

	simplified := SimplifyPath64(path, 0.1, false)

	// With small epsilon, collinear points should be removed
	// Expected: only endpoints and corners remain
	if len(simplified) >= len(path) {
		t.Errorf("SimplifyPath64() did not simplify: got %d points, original %d", len(simplified), len(path))
	}
}

func TestSimplifyPath64SmallPath(t *testing.T) {
	// Paths with less than 4 points should be returned unchanged
	paths := []Path64{
		{},
		{{X: 0, Y: 0}},
		{{X: 0, Y: 0}, {X: 10, Y: 10}},
		{{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 20, Y: 0}},
	}

	for i, path := range paths {
		simplified := SimplifyPath64(path, 1.0, true)
		if len(simplified) != len(path) {
			t.Errorf("SimplifyPath64(path[%d]) changed length: got %d, want %d", i, len(simplified), len(path))
		}
	}
}

func TestSimplifyPath64ClosedVsOpen(t *testing.T) {
	// Test that closed and open paths behave differently
	path := Path64{
		{X: 0, Y: 0},
		{X: 10, Y: 5},
		{X: 20, Y: 0},
		{X: 30, Y: 5},
		{X: 40, Y: 0},
	}

	closed := SimplifyPath64(path, 10.0, true)
	open := SimplifyPath64(path, 10.0, false)

	// Open paths should preserve endpoints
	if len(open) > 0 && (open[0] != path[0] || open[len(open)-1] != path[len(path)-1]) {
		t.Error("SimplifyPath64(open) did not preserve endpoints")
	}

	t.Logf("Original: %d points, Closed: %d points, Open: %d points", len(path), len(closed), len(open))
}

func TestSimplifyPaths64(t *testing.T) {
	paths := Paths64{
		{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
		{{X: 20, Y: 20}, {X: 25, Y: 20}, {X: 30, Y: 20}, {X: 30, Y: 30}},
	}

	simplified := SimplifyPaths64(paths, 0.1, false)

	if len(simplified) != len(paths) {
		t.Errorf("SimplifyPaths64() changed path count: got %d, want %d", len(simplified), len(paths))
	}

	// Each path should be simplified
	for i := range simplified {
		if len(simplified[i]) > len(paths[i]) {
			t.Errorf("SimplifyPaths64()[%d] increased size: got %d, original %d", i, len(simplified[i]), len(paths[i]))
		}
	}
}

// ==============================================================================
// Helper Tests
// ==============================================================================

func TestPerpendicDistFromLineSqrd(t *testing.T) {
	tests := []struct {
		name             string
		pt, line1, line2 Point64
		wantApprox       float64
		tolerance        float64
	}{
		{
			name:       "Point on line",
			pt:         Point64{5, 5},
			line1:      Point64{0, 0},
			line2:      Point64{10, 10},
			wantApprox: 0,
			tolerance:  0.01,
		},
		{
			name:       "Point perpendicular to line",
			pt:         Point64{5, 0},
			line1:      Point64{0, 0},
			line2:      Point64{10, 0},
			wantApprox: 0,
			tolerance:  0.01,
		},
		{
			name:       "Point above horizontal line",
			pt:         Point64{5, 10},
			line1:      Point64{0, 0},
			line2:      Point64{10, 0},
			wantApprox: 100, // 10^2 = 100
			tolerance:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := perpendicDistFromLineSqrd(tt.pt, tt.line1, tt.line2)
			diff := got - tt.wantApprox
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("perpendicDistFromLineSqrd() = %v, want ~%v", got, tt.wantApprox)
			}
		})
	}
}
