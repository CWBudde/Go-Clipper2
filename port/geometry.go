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

// GetSegmentIntersectPt calculates the intersection point of two line segments
// Returns the intersection point and true if segments intersect, or zero point and false if parallel
// This is a C++-compatible wrapper around SegmentIntersection
func GetSegmentIntersectPt(seg1a, seg1b, seg2a, seg2b Point64) (Point64, bool) {
	pt, intersectType, err := SegmentIntersection(seg1a, seg1b, seg2a, seg2b)
	if err != nil || intersectType == NoIntersection {
		return Point64{}, false
	}
	return pt, true
}

// GetClosestPointOnSegment finds the closest point on a line segment to a given point
// This implements the standard point-to-segment projection algorithm
func GetClosestPointOnSegment(pt, segA, segB Point64) Point64 {
	// Vector from segA to segB
	dx := float64(segB.X - segA.X)
	dy := float64(segB.Y - segA.Y)

	// If segment is a point, return that point
	if dx == 0 && dy == 0 {
		return segA
	}

	// Calculate projection parameter t (0 ≤ t ≤ 1)
	// t = dot(pt - segA, segB - segA) / dot(segB - segA, segB - segA)
	t := (float64(pt.X-segA.X)*dx + float64(pt.Y-segA.Y)*dy) / (dx*dx + dy*dy)

	// Clamp t to [0, 1] to stay within segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Calculate closest point
	return Point64{
		X: segA.X + int64(t*dx+0.5), // Round to nearest
		Y: segA.Y + int64(t*dy+0.5),
	}
}

// ==============================================================================
// 32-bit Coordinate Geometry Functions
// ==============================================================================
// These functions handle 32-bit coordinate types but use 64-bit intermediate
// calculations to avoid overflow

// CrossProduct64_32 calculates the cross product of vectors (p2-p1) and (p3-p1)
// for 32-bit points, returning int64 result (no overflow possible)
func CrossProduct64_32(p1, p2, p3 Point32) int64 {
	// Calculate vectors (promote to int64)
	v1x := int64(p2.X - p1.X)
	v1y := int64(p2.Y - p1.Y)
	v2x := int64(p3.X - p1.X)
	v2y := int64(p3.Y - p1.Y)

	// Cross product: v1x * v2y - v1y * v2x
	// int32 * int32 fits in int64, so no overflow
	return v1x*v2y - v1y*v2x
}

// CrossProductSign32 returns the sign of the cross product for 32-bit points
// Returns: -1 if negative, 0 if zero, +1 if positive
func CrossProductSign32(p1, p2, p3 Point32) int {
	cp := CrossProduct64_32(p1, p2, p3)
	if cp < 0 {
		return -1
	} else if cp == 0 {
		return 0
	}
	return 1
}

// IsCollinear32 checks if three 32-bit points are collinear
func IsCollinear32(p1, p2, p3 Point32) bool {
	return CrossProduct64_32(p1, p2, p3) == 0
}

// IsParallel32 checks if two line segments are parallel (32-bit)
func IsParallel32(seg1a, seg1b, seg2a, seg2b Point32) bool {
	// Segments are parallel if cross product of direction vectors is zero
	// Direction vector 1: (seg1b - seg1a)
	// Direction vector 2: (seg2b - seg2a)
	dx1 := int64(seg1b.X - seg1a.X)
	dy1 := int64(seg1b.Y - seg1a.Y)
	dx2 := int64(seg2b.X - seg2a.X)
	dy2 := int64(seg2b.Y - seg2a.Y)

	// Cross product of direction vectors
	return dx1*dy2 == dy1*dx2
}

// WindingNumber32 computes the winding number for a point relative to a polygon (32-bit)
// The winding number indicates how many times the polygon winds around the point
func WindingNumber32(point Point32, polygon Path32) int {
	wn := 0
	n := len(polygon)

	for i := 0; i < n; i++ {
		p1 := polygon[i]
		p2 := polygon[(i+1)%n]

		if p1.Y <= point.Y {
			if p2.Y > point.Y {
				// Upward crossing
				if CrossProduct64_32(p1, p2, point) > 0 {
					wn++
				}
			}
		} else {
			if p2.Y <= point.Y {
				// Downward crossing
				if CrossProduct64_32(p1, p2, point) < 0 {
					wn--
				}
			}
		}
	}

	return wn
}

// PointInPolygon32 determines if a point is inside, outside, or on the boundary of a polygon (32-bit)
func PointInPolygon32(point Point32, polygon Path32, fillRule FillRule) PolygonLocation {
	if len(polygon) < 3 {
		return Outside
	}

	wn := WindingNumber32(point, polygon)

	// Check if point is on boundary (winding number unreliable for boundary points)
	for i := 0; i < len(polygon); i++ {
		p1 := polygon[i]
		p2 := polygon[(i+1)%len(polygon)]

		// Check if point is on edge segment
		if IsCollinear32(p1, p2, point) {
			// Point is collinear, check if it's between p1 and p2
			minX := min32(p1.X, p2.X)
			maxX := max32(p1.X, p2.X)
			minY := min32(p1.Y, p2.Y)
			maxY := max32(p1.Y, p2.Y)

			if point.X >= minX && point.X <= maxX && point.Y >= minY && point.Y <= maxY {
				return OnBoundary
			}
		}
	}

	// Apply fill rule to winding number
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

// Helper functions for 32-bit coordinates
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func minMax32(a, b int32) (int32, int32) {
	if a < b {
		return a, b
	}
	return b, a
}
