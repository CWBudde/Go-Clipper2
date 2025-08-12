package clipper

// IntersectionType represents the type of intersection between two line segments
type IntersectionType uint8

const (
	NoIntersection      IntersectionType = iota // segments do not intersect
	PointIntersection                           // segments intersect at a single point
	OverlapIntersection                         // segments overlap along a line
)

// PolygonLocation represents the location of a point relative to a polygon
type PolygonLocation uint8

const (
	Outside    PolygonLocation = iota // point is outside the polygon
	Inside                            // point is inside the polygon
	OnBoundary                        // point is on the polygon boundary
)

// SegmentIntersection finds the intersection between two line segments
// Returns the intersection point, intersection type, and any error
func SegmentIntersection(seg1a, seg1b, seg2a, seg2b Point64) (Point64, IntersectionType, error) {
	// First check if segments are collinear
	if IsCollinear(seg1a, seg1b, seg2a) && IsCollinear(seg1a, seg1b, seg2b) {
		return handleCollinearSegments(seg1a, seg1b, seg2a, seg2b)
	}

	// Calculate cross products to determine intersection
	d1 := CrossProduct128(seg2a, seg2b, seg1a)
	d2 := CrossProduct128(seg2a, seg2b, seg1b)
	d3 := CrossProduct128(seg1a, seg1b, seg2a)
	d4 := CrossProduct128(seg1a, seg1b, seg2b)

	// Check if segments are parallel (cross products should be zero)
	if d1.IsZero() && d2.IsZero() && d3.IsZero() && d4.IsZero() {
		// Segments are collinear - this should have been caught above
		return handleCollinearSegments(seg1a, seg1b, seg2a, seg2b)
	}

	// Check for proper intersection
	if ((d1.IsNegative()) != (d2.IsNegative())) && ((d3.IsNegative()) != (d4.IsNegative())) {
		// Segments intersect at a point
		point, err := calculateIntersectionPoint(seg1a, seg1b, seg2a, seg2b, d1, d2)
		return point, PointIntersection, err
	}

	// Check for endpoint intersections
	if d1.IsZero() && isPointOnSegment(seg1a, seg2a, seg2b) {
		return seg1a, PointIntersection, nil
	}
	if d2.IsZero() && isPointOnSegment(seg1b, seg2a, seg2b) {
		return seg1b, PointIntersection, nil
	}
	if d3.IsZero() && isPointOnSegment(seg2a, seg1a, seg1b) {
		return seg2a, PointIntersection, nil
	}
	if d4.IsZero() && isPointOnSegment(seg2b, seg1a, seg1b) {
		return seg2b, PointIntersection, nil
	}

	return Point64{}, NoIntersection, nil
}

// IsCollinear checks if three points are collinear using robust 128-bit cross product
func IsCollinear(p1, p2, p3 Point64) bool {
	return CrossProduct128(p1, p2, p3).IsZero()
}

// IsParallel checks if two line segments are parallel
func IsParallel(seg1a, seg1b, seg2a, seg2b Point64) bool {
	// Calculate direction vectors
	v1x := seg1b.X - seg1a.X
	v1y := seg1b.Y - seg1a.Y
	v2x := seg2b.X - seg2a.X
	v2y := seg2b.Y - seg2a.Y

	// Cross product of direction vectors
	cross := NewInt128(v1x).Mul64(v2y).Sub(NewInt128(v1y).Mul64(v2x))
	return cross.IsZero()
}

// PointInPolygon determines if a point is inside, outside, or on the boundary of a polygon
func PointInPolygon(point Point64, polygon Path64, fillRule FillRule) PolygonLocation {
	if len(polygon) < 3 {
		return Outside
	}

	// First check if point is on any edge
	for i := 0; i < len(polygon); i++ {
		j := (i + 1) % len(polygon)
		if isPointOnSegment(point, polygon[i], polygon[j]) {
			return OnBoundary
		}
	}

	// Calculate winding number
	wn := WindingNumber(point, polygon)

	// Apply fill rule to determine inside/outside
	switch fillRule {
	case EvenOdd:
		if wn%2 != 0 {
			return Inside
		}
	case NonZero:
		if wn != 0 {
			return Inside
		}
	case Positive:
		if wn > 0 {
			return Inside
		}
	case Negative:
		if wn < 0 {
			return Inside
		}
	}

	return Outside
}

// WindingNumber calculates the winding number of a point with respect to a polygon
// Uses robust 128-bit arithmetic to handle edge cases
func WindingNumber(point Point64, polygon Path64) int {
	if len(polygon) < 3 {
		return 0
	}

	wn := 0

	for i := 0; i < len(polygon); i++ {
		j := (i + 1) % len(polygon)

		if polygon[i].Y <= point.Y {
			if polygon[j].Y > point.Y { // upward crossing
				if isLeft(polygon[i], polygon[j], point) {
					wn++
				}
			}
		} else {
			if polygon[j].Y <= point.Y { // downward crossing
				if !isLeft(polygon[i], polygon[j], point) {
					wn--
				}
			}
		}
	}

	return wn
}

// calculateIntersectionPoint calculates the exact intersection point of two segments
// using the cross product values already computed
func calculateIntersectionPoint(seg1a, seg1b, _seg2a, _seg2b Point64, d1, d2 Int128) (Point64, error) {
	// Calculate the intersection using parametric form
	// P = seg1a + t * (seg1b - seg1a) where t = d1 / (d1 - d2)

	denominator := d1.Sub(d2)
	if denominator.IsZero() {
		return Point64{}, ErrInvalidInput
	}

	// Use floating point for final intersection calculation
	// This is acceptable as we've already used robust arithmetic for the critical decisions
	t := d1.ToFloat64() / denominator.ToFloat64()

	x := float64(seg1a.X) + t*float64(seg1b.X-seg1a.X)
	y := float64(seg1a.Y) + t*float64(seg1b.Y-seg1a.Y)

	return Point64{X: int64(x + 0.5), Y: int64(y + 0.5)}, nil
}

// handleCollinearSegments handles intersection of collinear segments
func handleCollinearSegments(seg1a, seg1b, seg2a, seg2b Point64) (Point64, IntersectionType, error) {
	// For collinear segments, check if they overlap
	// Project onto the axis with larger coordinate range for better numerical stability

	dx1 := abs64(seg1b.X - seg1a.X)
	dy1 := abs64(seg1b.Y - seg1a.Y)

	if dx1 >= dy1 {
		// Project onto X axis
		min1, max1 := minMax64(seg1a.X, seg1b.X)
		min2, max2 := minMax64(seg2a.X, seg2b.X)

		if max1 < min2 || max2 < min1 {
			return Point64{}, NoIntersection, nil
		}

		// Find overlap
		overlapMin := max64(min1, min2)
		overlapMax := min64(max1, max2)

		if overlapMin == overlapMax {
			// Single point overlap
			y := seg1a.Y + (seg1b.Y-seg1a.Y)*(overlapMin-seg1a.X)/(seg1b.X-seg1a.X)
			return Point64{X: overlapMin, Y: y}, PointIntersection, nil
		}

		// Line segment overlap
		y := seg1a.Y + (seg1b.Y-seg1a.Y)*(overlapMin-seg1a.X)/(seg1b.X-seg1a.X)
		return Point64{X: overlapMin, Y: y}, OverlapIntersection, nil
	} else {
		// Project onto Y axis
		min1, max1 := minMax64(seg1a.Y, seg1b.Y)
		min2, max2 := minMax64(seg2a.Y, seg2b.Y)

		if max1 < min2 || max2 < min1 {
			return Point64{}, NoIntersection, nil
		}

		// Find overlap
		overlapMin := max64(min1, min2)
		overlapMax := min64(max1, max2)

		if overlapMin == overlapMax {
			// Single point overlap
			x := seg1a.X + (seg1b.X-seg1a.X)*(overlapMin-seg1a.Y)/(seg1b.Y-seg1a.Y)
			return Point64{X: x, Y: overlapMin}, PointIntersection, nil
		}

		// Line segment overlap
		x := seg1a.X + (seg1b.X-seg1a.X)*(overlapMin-seg1a.Y)/(seg1b.Y-seg1a.Y)
		return Point64{X: x, Y: overlapMin}, OverlapIntersection, nil
	}
}

// isPointOnSegment checks if a point lies on a line segment
func isPointOnSegment(point, segA, segB Point64) bool {
	if !IsCollinear(segA, segB, point) {
		return false
	}

	// Check if point is within segment bounds
	return (point.X >= min64(segA.X, segB.X) && point.X <= max64(segA.X, segB.X) &&
		point.Y >= min64(segA.Y, segB.Y) && point.Y <= max64(segA.Y, segB.Y))
}

// isLeft tests if point is left of the line from p1 to p2
func isLeft(p1, p2, point Point64) bool {
	return !CrossProduct128(p1, p2, point).IsNegative() // left is positive in our coordinate system
}

// Helper functions for int64 operations
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minMax64(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}
