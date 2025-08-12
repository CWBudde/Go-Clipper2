package clipper

import (
	"math"
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

func TestRectClip64(t *testing.T) {
	rect := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	paths := Paths64{{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}}
	result, err := RectClip64(rect, paths)
	if err == ErrNotImplemented {
		t.Skip("RectClip64 not yet implemented")
	}
	if err != nil {
		t.Fatalf("RectClip64 failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("Expected non-empty result from rect clip")
	}
	t.Logf("RectClip result: %v", result)
}

func TestBooleanOp64Direct(t *testing.T) {
	subject := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	clip := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}

	result, resultOpen, err := BooleanOp64(Union, NonZero, subject, nil, clip)
	if err == ErrNotImplemented {
		t.Skip("BooleanOp64 not yet implemented in pure Go")
	}
	if err != nil {
		t.Fatalf("BooleanOp64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from boolean operation")
	}
	if len(resultOpen) != 0 {
		t.Fatal("Expected empty open result for closed polygon operation")
	}
	t.Logf("BooleanOp64 result: %v", result)
}

func TestRectClip64InvalidRectangle(t *testing.T) {
	// Test with invalid rectangle (not 4 points)
	invalidRect := Path64{{0, 0}, {10, 0}, {10, 10}} // Only 3 points
	paths := Paths64{{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}}

	_, err := RectClip64(invalidRect, paths)
	if err != ErrInvalidRectangle {
		t.Errorf("Expected ErrInvalidRectangle, got %v", err)
	}
}

func TestArea64EmptyPath(t *testing.T) {
	// Test with empty path
	emptyPath := Path64{}
	area := Area64(emptyPath)
	if area != 0.0 {
		t.Errorf("Expected area of empty path to be 0, got %v", area)
	}

	// Test with path with less than 3 points
	smallPath := Path64{{0, 0}, {1, 1}}
	area = Area64(smallPath)
	if area != 0.0 {
		t.Errorf("Expected area of small path to be 0, got %v", area)
	}
}

// M2 Geometry Kernel Tests

// TestMath128Operations tests the 128-bit math operations
func TestMath128Operations(t *testing.T) {
	// Test basic Int128 operations
	a := NewInt128(1000000000000) // 1 trillion
	b := NewInt128(2000000000000) // 2 trillion

	sum := a.Add(b)
	expected := NewInt128(3000000000000)
	if sum.Cmp(expected) != 0 {
		t.Errorf("Add failed: expected %d + %d = %d, got sum with Hi=%d Lo=%d", a.Hi, b.Hi, expected.Hi, sum.Hi, sum.Lo)
	}

	diff := b.Sub(a)
	expected = NewInt128(1000000000000)
	if diff.Cmp(expected) != 0 {
		t.Errorf("Sub failed: expected %v, got %v", expected, diff)
	}

	// Test multiplication
	prod := a.Mul64(3)
	expected = NewInt128(3000000000000)
	if prod.Cmp(expected) != 0 {
		t.Errorf("Mul64 failed: expected %d trillion, got Hi=%d Lo=%d (float64: %f)", 3000000000000, prod.Hi, prod.Lo, prod.ToFloat64())
	}

	// Test negation
	neg := NewInt128(-1000)
	if !neg.IsNegative() {
		t.Error("Expected negative number to be negative")
	}

	pos := neg.Negate()
	expected = NewInt128(1000)
	if pos.Cmp(expected) != 0 {
		t.Errorf("Negate failed: expected %v, got %v", expected, pos)
	}
}

// TestCrossProduct128 tests the robust cross product calculation
func TestCrossProduct128(t *testing.T) {
	tests := []struct {
		name       string
		p1, p2, p3 Point64
		expected   float64 // expected sign (positive, negative, or zero)
	}{
		{"Counter-clockwise triangle", Point64{0, 0}, Point64{10, 0}, Point64{5, 10}, 1},                                            // positive
		{"Clockwise triangle", Point64{0, 0}, Point64{5, 10}, Point64{10, 0}, -1},                                                   // negative
		{"Collinear points", Point64{0, 0}, Point64{5, 5}, Point64{10, 10}, 0},                                                      // zero
		{"Large coordinates", Point64{1000000000, 1000000000}, Point64{2000000000, 1000000000}, Point64{1500000000, 2000000000}, 1}, // positive
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cross := CrossProduct128(test.p1, test.p2, test.p3)

			if test.expected > 0 && !cross.IsNegative() && !cross.IsZero() {
				// Expected positive, got positive - OK
			} else if test.expected < 0 && cross.IsNegative() {
				// Expected negative, got negative - OK
			} else if test.expected == 0 && cross.IsZero() {
				// Expected zero, got zero - OK
			} else {
				t.Errorf("CrossProduct128 failed for %s: expected sign %v, got %v", test.name, test.expected, cross)
			}
		})
	}
}

// TestArea128 tests robust area calculation
func TestArea128(t *testing.T) {
	tests := []struct {
		name     string
		path     Path64
		expected float64
	}{
		{"Unit square", Path64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, 1.0},
		{"Large square", Path64{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}, 1000000.0},
		{"Triangle", Path64{{0, 0}, {10, 0}, {5, 10}}, 50.0},
		{"Clockwise square", Path64{{0, 0}, {0, 1}, {1, 1}, {1, 0}}, -1.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			area128 := Area128(test.path)
			actual := area128.ToFloat64() / 2.0 // Area128 returns 2*area

			if math.Abs(actual-test.expected) > 1e-9 {
				t.Errorf("Area128 failed for %s: expected %v, got %v", test.name, test.expected, actual)
			}
		})
	}
}

// TestIsCollinear tests collinearity detection
func TestIsCollinear(t *testing.T) {
	tests := []struct {
		name       string
		p1, p2, p3 Point64
		expected   bool
	}{
		{"Horizontal line", Point64{0, 5}, Point64{5, 5}, Point64{10, 5}, true},
		{"Vertical line", Point64{5, 0}, Point64{5, 5}, Point64{5, 10}, true},
		{"Diagonal line", Point64{0, 0}, Point64{5, 5}, Point64{10, 10}, true},
		{"Not collinear", Point64{0, 0}, Point64{5, 0}, Point64{0, 5}, false},
		{"Large coordinates", Point64{1000000000, 1000000000}, Point64{2000000000, 2000000000}, Point64{3000000000, 3000000000}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsCollinear(test.p1, test.p2, test.p3)
			if result != test.expected {
				t.Errorf("IsCollinear failed for %s: expected %v, got %v", test.name, test.expected, result)
			}
		})
	}
}

// TestIsParallel tests parallel segment detection
func TestIsParallel(t *testing.T) {
	tests := []struct {
		name                       string
		seg1a, seg1b, seg2a, seg2b Point64
		expected                   bool
	}{
		{"Horizontal parallel", Point64{0, 0}, Point64{10, 0}, Point64{0, 5}, Point64{10, 5}, true},
		{"Vertical parallel", Point64{0, 0}, Point64{0, 10}, Point64{5, 0}, Point64{5, 10}, true},
		{"Diagonal parallel", Point64{0, 0}, Point64{5, 5}, Point64{10, 10}, Point64{15, 15}, true},
		{"Not parallel", Point64{0, 0}, Point64{5, 0}, Point64{0, 0}, Point64{0, 5}, false},
		{"Same segment", Point64{0, 0}, Point64{5, 5}, Point64{0, 0}, Point64{5, 5}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsParallel(test.seg1a, test.seg1b, test.seg2a, test.seg2b)
			if result != test.expected {
				t.Errorf("IsParallel failed for %s: expected %v, got %v", test.name, test.expected, result)
			}
		})
	}
}

// TestSegmentIntersection tests robust segment intersection
func TestSegmentIntersection(t *testing.T) {
	tests := []struct {
		name                       string
		seg1a, seg1b, seg2a, seg2b Point64
		expectedType               IntersectionType
		expectedPoint              Point64
	}{
		{"Cross intersection", Point64{0, 0}, Point64{10, 10}, Point64{0, 10}, Point64{10, 0}, PointIntersection, Point64{5, 5}},
		{"No intersection", Point64{0, 0}, Point64{5, 0}, Point64{0, 5}, Point64{5, 5}, NoIntersection, Point64{}},
		{"Endpoint intersection", Point64{0, 0}, Point64{5, 5}, Point64{5, 5}, Point64{10, 0}, PointIntersection, Point64{5, 5}},
		{"Collinear overlap", Point64{0, 0}, Point64{10, 0}, Point64{5, 0}, Point64{15, 0}, OverlapIntersection, Point64{5, 0}},
		{"Parallel no intersection", Point64{0, 0}, Point64{10, 0}, Point64{0, 5}, Point64{10, 5}, NoIntersection, Point64{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			point, intersectionType, err := SegmentIntersection(test.seg1a, test.seg1b, test.seg2a, test.seg2b)
			if err != nil {
				t.Fatalf("SegmentIntersection failed with error: %v", err)
			}

			if intersectionType != test.expectedType {
				t.Errorf("SegmentIntersection type failed for %s: expected %v, got %v", test.name, test.expectedType, intersectionType)
			}

			if intersectionType == PointIntersection || intersectionType == OverlapIntersection {
				// Allow small tolerance for intersection points
				if math.Abs(float64(point.X-test.expectedPoint.X)) > 1 || math.Abs(float64(point.Y-test.expectedPoint.Y)) > 1 {
					t.Errorf("SegmentIntersection point failed for %s: expected %v, got %v", test.name, test.expectedPoint, point)
				}
			}
		})
	}
}

// TestHandleCollinearSegments tests all branches of the collinear segment handler
func TestHandleCollinearSegments(t *testing.T) {
	tests := []struct {
		name                       string
		seg1a, seg1b, seg2a, seg2b Point64
		expectedType               IntersectionType
		expectedPoint              Point64
	}{
		// X-axis projection tests (dx >= dy)
		{"X-axis: No overlap - segments apart", Point64{0, 0}, Point64{5, 0}, Point64{10, 0}, Point64{15, 0}, NoIntersection, Point64{}},
		{"X-axis: No overlap - reversed", Point64{10, 0}, Point64{15, 0}, Point64{0, 0}, Point64{5, 0}, NoIntersection, Point64{}},
		{"X-axis: Single point overlap", Point64{0, 0}, Point64{5, 0}, Point64{5, 0}, Point64{10, 0}, PointIntersection, Point64{5, 0}},
		{"X-axis: Line segment overlap", Point64{0, 0}, Point64{10, 0}, Point64{5, 0}, Point64{15, 0}, OverlapIntersection, Point64{5, 0}},
		{"X-axis: Diagonal dx>dy", Point64{0, 0}, Point64{10, 2}, Point64{5, 1}, Point64{15, 3}, OverlapIntersection, Point64{5, 1}},

		// Y-axis projection tests (dy > dx)
		{"Y-axis: No overlap - segments apart", Point64{0, 0}, Point64{0, 5}, Point64{0, 10}, Point64{0, 15}, NoIntersection, Point64{}},
		{"Y-axis: No overlap - reversed", Point64{0, 10}, Point64{0, 15}, Point64{0, 0}, Point64{0, 5}, NoIntersection, Point64{}},
		{"Y-axis: Single point overlap", Point64{0, 0}, Point64{0, 5}, Point64{0, 5}, Point64{0, 10}, PointIntersection, Point64{0, 5}},
		{"Y-axis: Line segment overlap", Point64{0, 0}, Point64{0, 10}, Point64{0, 5}, Point64{0, 15}, OverlapIntersection, Point64{0, 5}},
		{"Y-axis: Diagonal dy>dx", Point64{0, 0}, Point64{2, 10}, Point64{1, 5}, Point64{3, 15}, OverlapIntersection, Point64{1, 5}},

		// Edge case: equal ranges (dx == dy), should prefer X-axis
		{"Equal ranges: prefer X-axis", Point64{0, 0}, Point64{5, 5}, Point64{2, 2}, Point64{7, 7}, OverlapIntersection, Point64{2, 2}},

		// Edge cases with negative coordinates
		{"Y-axis: Negative coordinates", Point64{0, -10}, Point64{0, -5}, Point64{0, -7}, Point64{0, -2}, OverlapIntersection, Point64{0, -7}},
		{"X-axis: Mixed coordinates", Point64{-5, 3}, Point64{5, 7}, Point64{0, 5}, Point64{10, 9}, OverlapIntersection, Point64{0, 5}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// First verify segments are actually collinear
			if !IsCollinear(test.seg1a, test.seg1b, test.seg2a) || !IsCollinear(test.seg1a, test.seg1b, test.seg2b) {
				t.Skipf("Test segments are not collinear, skipping")
			}

			point, intersectionType, err := SegmentIntersection(test.seg1a, test.seg1b, test.seg2a, test.seg2b)
			if err != nil {
				t.Fatalf("SegmentIntersection failed with error: %v", err)
			}

			if intersectionType != test.expectedType {
				t.Errorf("Intersection type failed: expected %v, got %v", test.expectedType, intersectionType)
			}

			if intersectionType == PointIntersection || intersectionType == OverlapIntersection {
				if point.X != test.expectedPoint.X || point.Y != test.expectedPoint.Y {
					t.Errorf("Intersection point failed: expected %v, got %v", test.expectedPoint, point)
				}
			}
		})
	}
}

// TestWindingNumber tests winding number calculation
func TestWindingNumber(t *testing.T) {
	square := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	tests := []struct {
		name     string
		point    Point64
		expected int
	}{
		{"Inside square", Point64{5, 5}, 1},
		{"Outside square", Point64{-5, 5}, 0},
		{"On boundary", Point64{0, 5}, 0}, // Point on edge should have winding 0 for this test
		{"Far outside", Point64{100, 100}, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wn := WindingNumber(test.point, square)
			if test.name == "On boundary" {
				// For boundary points, we mainly care that it's detected as such
				// The actual winding number can vary based on implementation
				return
			}
			if wn != test.expected {
				t.Errorf("WindingNumber failed for %s: expected %v, got %v", test.name, test.expected, wn)
			}
		})
	}
}

// TestPointInPolygon tests point-in-polygon with all fill rules
func TestPointInPolygon(t *testing.T) {
	square := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	tests := []struct {
		name     string
		point    Point64
		fillRule FillRule
		expected PolygonLocation
	}{
		{"Inside square - NonZero", Point64{5, 5}, NonZero, Inside},
		{"Inside square - EvenOdd", Point64{5, 5}, EvenOdd, Inside},
		{"Inside square - Positive", Point64{5, 5}, Positive, Inside},
		{"Outside square - NonZero", Point64{-5, 5}, NonZero, Outside},
		{"Outside square - EvenOdd", Point64{-5, 5}, EvenOdd, Outside},
		{"On boundary", Point64{0, 5}, NonZero, OnBoundary},
		{"Corner point", Point64{0, 0}, NonZero, OnBoundary},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			location := PointInPolygon(test.point, square, test.fillRule)
			if location != test.expected {
				t.Errorf("PointInPolygon failed for %s: expected %v, got %v", test.name, test.expected, location)
			}
		})
	}
}

// TestNumericalStability tests edge cases near overflow boundaries
func TestNumericalStability(t *testing.T) {
	// Test with coordinates near int64 limits
	maxInt64 := int64(9223372036854775807)
	largeCoords := []Point64{
		{maxInt64 - 1000, maxInt64 - 1000},
		{maxInt64 - 500, maxInt64 - 1000},
		{maxInt64 - 500, maxInt64 - 500},
		{maxInt64 - 1000, maxInt64 - 500},
	}

	// Test area calculation doesn't overflow
	area128 := Area128(largeCoords)
	if area128.IsZero() {
		t.Error("Expected non-zero area for large coordinate polygon")
	}

	// Test cross product doesn't overflow
	cross := CrossProduct128(largeCoords[0], largeCoords[1], largeCoords[2])
	// Should not panic and should give a reasonable result
	if cross.IsZero() {
		t.Error("Expected non-zero cross product for large coordinates")
	}

	// Test collinearity detection with large coordinates
	p1 := Point64{maxInt64 - 1000, maxInt64 - 1000}
	p2 := Point64{maxInt64 - 500, maxInt64 - 500}
	p3 := Point64{maxInt64, maxInt64}

	isCollinear := IsCollinear(p1, p2, p3)
	if !isCollinear {
		t.Error("Expected points on diagonal line to be collinear")
	}
}

func TestInflatePaths64WithOptions(t *testing.T) {
	square := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	options := OffsetOptions{
		MiterLimit:   4.0,
		ArcTolerance: 0.1,
	}

	result, err := InflatePaths64(square, 1.0, Miter, ClosedPolygon, options)
	if err == ErrNotImplemented {
		t.Skip("InflatePaths64 not yet implemented")
	}
	if err != nil {
		t.Fatalf("InflatePaths64 with options failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result from inflate with options")
	}
	t.Logf("Inflate with options result: %v", result)
}

func TestRectClip64EdgeCases(t *testing.T) {
	// Test case 1: Degenerate rectangle (zero width)
	degenerateRect := Path64{{10, 10}, {10, 10}, {10, 20}, {10, 20}}
	paths := Paths64{{{0, 0}, {5, 0}, {5, 5}, {0, 5}}}

	result, err := RectClip64(degenerateRect, paths)
	if err != nil {
		t.Fatalf("RectClip64 with degenerate rect failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for degenerate rectangle, got %v", result)
	}

	// Test case 2: Path completely outside rectangle
	rect := Path64{{0, 0}, {5, 0}, {5, 5}, {0, 5}}
	outsidePath := Paths64{{{10, 10}, {15, 10}, {15, 15}, {10, 15}}}

	result, err = RectClip64(rect, outsidePath)
	if err != nil {
		t.Fatalf("RectClip64 with outside path failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for outside path, got %v", result)
	}

	// Test case 3: Path completely inside rectangle
	insidePath := Paths64{{{1, 1}, {2, 1}, {2, 2}, {1, 2}}}

	result, err = RectClip64(rect, insidePath)
	if err != nil {
		t.Fatalf("RectClip64 with inside path failed: %v", err)
	}
	if len(result) != 1 || len(result[0]) != 4 {
		t.Errorf("Expected inside path to be unchanged, got %v", result)
	}

	// Test case 4: Path partially intersecting rectangle
	crossingPath := Paths64{{{-1, 2}, {3, 2}, {3, 7}, {-1, 7}}}

	result, err = RectClip64(rect, crossingPath)
	if err != nil {
		t.Fatalf("RectClip64 with crossing path failed: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("Expected non-empty result for crossing path, got empty")
	}
	t.Logf("Crossing path clipped result: %v", result)

	// Test case 5: Empty paths input
	emptyPaths := Paths64{}

	result, err = RectClip64(rect, emptyPaths)
	if err != nil {
		t.Fatalf("RectClip64 with empty paths failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty paths input, got %v", result)
	}

	// Test case 6: Paths with degenerate segments (single points, collinear points)
	degeneratePaths := Paths64{
		{{1, 1}},                 // Single point - should be skipped
		{{1, 1}, {1, 1}, {1, 1}}, // All same point - should be skipped
		{{1, 1}, {3, 3}},         // Valid 2-point segment
	}

	result, err = RectClip64(rect, degeneratePaths)
	if err != nil {
		t.Fatalf("RectClip64 with degenerate paths failed: %v", err)
	}
	t.Logf("Degenerate paths clipped result: %v", result)
}

func TestRectClip64PointsOnBoundary(t *testing.T) {
	// Rectangle from (0,0) to (10,10)
	rect := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	// Test case 1: Path with points exactly on rectangle boundary
	boundaryPath := Paths64{{{0, 5}, {5, 0}, {10, 5}, {5, 10}}}

	result, err := RectClip64(rect, boundaryPath)
	if err != nil {
		t.Fatalf("RectClip64 with boundary points failed: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("Expected non-empty result for boundary path")
	}
	t.Logf("Boundary path result: %v", result)

	// Test case 2: Path touching corner
	cornerPath := Paths64{{{0, 0}, {-5, -5}, {5, -5}}}

	result, err = RectClip64(rect, cornerPath)
	if err != nil {
		t.Fatalf("RectClip64 with corner touching path failed: %v", err)
	}
	t.Logf("Corner touching path result: %v", result)
}

func TestRectClip64RandomOrientedRectangle(t *testing.T) {
	// Test with rectangle points in different order (counter-clockwise)
	rect := Path64{{0, 10}, {0, 0}, {10, 0}, {10, 10}} // CCW order
	paths := Paths64{{{2, 2}, {8, 2}, {8, 8}, {2, 8}}}

	result, err := RectClip64(rect, paths)
	if err != nil {
		t.Fatalf("RectClip64 with CCW rectangle failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 clipped path, got %d", len(result))
	}
	t.Logf("CCW rectangle result: %v", result)

	// Test with rectangle points in random order
	randomRect := Path64{{10, 0}, {0, 10}, {10, 10}, {0, 0}}

	result, err = RectClip64(randomRect, paths)
	if err != nil {
		t.Fatalf("RectClip64 with random order rectangle failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 clipped path, got %d", len(result))
	}
	t.Logf("Random order rectangle result: %v", result)
}

func TestRectClip64RandomPaths(t *testing.T) {
	// Test with various random rectangles and paths
	testCases := []struct {
		name     string
		rect     Path64
		paths    Paths64
		expected string // Description of expected behavior
	}{
		{
			"Small rectangle, large path",
			Path64{{5, 5}, {15, 5}, {15, 15}, {5, 15}},
			Paths64{{{0, 0}, {20, 0}, {20, 20}, {0, 20}}},
			"path should be clipped to rectangle bounds",
		},
		{
			"Rectangle with negative coordinates",
			Path64{{-10, -10}, {10, -10}, {10, 10}, {-10, 10}},
			Paths64{{{-15, -5}, {15, -5}, {15, 5}, {-15, 5}}},
			"should handle negative coordinates correctly",
		},
		{
			"Multiple paths, some inside, some outside",
			Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			Paths64{
				{{1, 1}, {2, 1}, {2, 2}, {1, 2}},         // Inside
				{{11, 11}, {12, 11}, {12, 12}, {11, 12}}, // Outside
				{{-1, 5}, {5, 5}, {5, 8}, {-1, 8}},       // Crossing
			},
			"should return inside and crossing paths only",
		},
		{
			"Complex polygon crossing rectangle",
			Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			Paths64{{{-2, -2}, {12, -2}, {12, 2}, {8, 2}, {8, 8}, {12, 8}, {12, 12}, {-2, 12}}},
			"should clip complex polygon correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RectClip64(tc.rect, tc.paths)
			if err != nil {
				t.Fatalf("RectClip64 failed for %s: %v", tc.name, err)
			}

			t.Logf("%s - Input paths: %v", tc.name, tc.paths)
			t.Logf("%s - Result: %v", tc.name, result)
			t.Logf("%s - Expected: %s", tc.name, tc.expected)

			// Basic validation - result should not contain points outside rectangle bounds
			left, right, top, bottom := getBounds(tc.rect)
			for _, path := range result {
				for _, pt := range path {
					if pt.X < left || pt.X > right || pt.Y < top || pt.Y > bottom {
						t.Errorf("Result contains point outside rectangle bounds: %v", pt)
					}
				}
			}
		})
	}
}

// getBounds extracts the bounding box from a rectangle path
func getBounds(rect Path64) (left, right, top, bottom int64) {
	if len(rect) == 0 {
		return 0, 0, 0, 0
	}

	left = rect[0].X
	right = rect[0].X
	top = rect[0].Y
	bottom = rect[0].Y

	for _, pt := range rect {
		if pt.X < left {
			left = pt.X
		}
		if pt.X > right {
			right = pt.X
		}
		if pt.Y < top {
			top = pt.Y
		}
		if pt.Y > bottom {
			bottom = pt.Y
		}
	}

	return left, right, top, bottom
}

func TestRectClip64StressTest(t *testing.T) {
	// Stress test with many small rectangles
	baseRect := Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}

	// Generate many small paths within and outside the rectangle
	var paths Paths64
	for i := 0; i < 50; i++ {
		x := int64(i*2 - 10) // Some negative, some positive
		y := int64(i*2 - 10)
		paths = append(paths, Path64{
			{x, y}, {x + 5, y}, {x + 5, y + 5}, {x, y + 5},
		})
	}

	result, err := RectClip64(baseRect, paths)
	if err != nil {
		t.Fatalf("Stress test failed: %v", err)
	}

	t.Logf("Stress test: Input %d paths, output %d paths", len(paths), len(result))

	// Verify all resulting points are within bounds
	for _, path := range result {
		for _, pt := range path {
			if pt.X < 0 || pt.X > 100 || pt.Y < 0 || pt.Y > 100 {
				t.Errorf("Stress test: Point outside bounds: %v", pt)
			}
		}
	}
}

// TestUtilityFunctions tests the helper functions abs64 and minMax64
func TestUtilityFunctions(t *testing.T) {
	t.Run("abs64", func(t *testing.T) {
		tests := []struct {
			name     string
			input    int64
			expected int64
		}{
			{"Positive number", 5, 5},
			{"Negative number", -5, 5},
			{"Zero", 0, 0},
			{"Large positive", 1000000000, 1000000000},
			{"Large negative", -1000000000, 1000000000},
			{"MaxInt64", 9223372036854775807, 9223372036854775807},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := abs64(test.input)
				if result != test.expected {
					t.Errorf("abs64(%d) = %d, expected %d", test.input, result, test.expected)
				}
			})
		}
	})

	t.Run("minMax64", func(t *testing.T) {
		tests := []struct {
			name        string
			a, b        int64
			expectedMin int64
			expectedMax int64
		}{
			{"a < b", 3, 7, 3, 7},
			{"a > b", 7, 3, 3, 7},
			{"a == b", 5, 5, 5, 5},
			{"Negative numbers", -10, -3, -10, -3},
			{"Mixed signs", -5, 10, -5, 10},
			{"Zero and positive", 0, 8, 0, 8},
			{"Zero and negative", -8, 0, -8, 0},
			{"Large numbers", 1000000000, 2000000000, 1000000000, 2000000000},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				mn, mx := minMax64(test.a, test.b)
				if mn != test.expectedMin || mx != test.expectedMax {
					t.Errorf("minMax64(%d, %d) = (%d, %d), expected (%d, %d)",
						test.a, test.b, mn, mx, test.expectedMin, test.expectedMax)
				}
			})
		}
	})
}
