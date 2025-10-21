package clipper

import (
	"math"
)

// PointD represents a point with float64 coordinates (used for normals and intermediate calculations)
type PointD struct {
	X, Y float64
}

// Negate negates the coordinates of a PointD
func (p *PointD) Negate() {
	p.X = -p.X
	p.Y = -p.Y
}

const (
	floatingPointTolerance = 1e-12
	arcConst               = 0.002 // 1/500 - default arc tolerance ratio
)

// Hypot calculates the hypotenuse (Euclidean distance) of x and y
// Reference: clipper.offset.cpp lines 60-68
func hypot(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

// AlmostZero returns true if the value is very close to zero
// Reference: clipper.offset.cpp lines 81-84
func almostZero(value, epsilon float64) bool {
	return math.Abs(value) < epsilon
}

// GetUnitNormal calculates the perpendicular unit vector from pt1 to pt2
// Reference: clipper.offset.cpp lines 70-79
func getUnitNormal(pt1, pt2 Point64) PointD {
	if pt1 == pt2 {
		return PointD{0.0, 0.0}
	}
	dx := float64(pt2.X - pt1.X)
	dy := float64(pt2.Y - pt1.Y)
	inverseHypot := 1.0 / hypot(dx, dy)
	dx *= inverseHypot
	dy *= inverseHypot
	// Return perpendicular vector (rotated 90° clockwise)
	return PointD{dy, -dx}
}

// GetPerpendic calculates a point offset perpendicular to pt by distance delta along normal
// Reference: clipper.offset.cpp lines 104-111
func getPerpendic(pt Point64, norm PointD, delta float64) Point64 {
	return Point64{
		X: pt.X + int64(norm.X*delta),
		Y: pt.Y + int64(norm.Y*delta),
	}
}

// GetPerpendicD calculates a float point offset perpendicular to pt by distance delta along normal
// Reference: clipper.offset.cpp lines 113-120
func getPerpendicD(pt Point64, norm PointD, delta float64) PointD {
	return PointD{
		X: float64(pt.X) + norm.X*delta,
		Y: float64(pt.Y) + norm.Y*delta,
	}
}

// IsClosedPath returns true if the end type represents a closed path
// Reference: clipper.offset.cpp lines 99-102
func isClosedPath(et EndType) bool {
	return et == EndPolygon || et == EndJoined
}

// GetLowestClosedPathInfo finds the path with the lowest point and determines if it has negative area
// This is used to determine overall orientation for polygon offsetting
// Reference: clipper.offset.cpp lines 36-58
func getLowestClosedPathInfo(paths Paths64) (lowestIdx *int, isNegArea bool) {
	lowestIdx = nil
	botPt := Point64{X: math.MaxInt64, Y: math.MinInt64}

	for i := range paths {
		a := math.MaxFloat64
		for _, pt := range paths[i] {
			// Look for lowest point (highest Y value, then rightmost X)
			if (pt.Y < botPt.Y) || (pt.Y == botPt.Y && pt.X >= botPt.X) {
				continue
			}
			if a == math.MaxFloat64 {
				a = Area64(paths[i])
				if a == 0 {
					break // invalid closed path
				}
				isNegArea = a < 0
			}
			idx := i
			lowestIdx = &idx
			botPt.X = pt.X
			botPt.Y = pt.Y
		}
	}

	return lowestIdx, isNegArea
}

// NormalizeVector normalizes a vector to unit length
// Reference: clipper.offset.cpp lines 86-92
func normalizeVector(vec PointD) PointD {
	h := hypot(vec.X, vec.Y)
	if almostZero(h, 0.001) {
		return PointD{0, 0}
	}
	inverseHypot := 1.0 / h
	return PointD{vec.X * inverseHypot, vec.Y * inverseHypot}
}

// GetAvgUnitVector calculates the average of two unit vectors and normalizes it
// Reference: clipper.offset.cpp lines 94-97
// Note: Used in Phase 3 (Square joins), included here for completeness
func getAvgUnitVector(vec1, vec2 PointD) PointD {
	return normalizeVector(PointD{vec1.X + vec2.X, vec1.Y + vec2.Y})
}

// TranslatePoint translates a point by dx and dy
func translatePoint(pt PointD, dx, dy float64) PointD {
	return PointD{pt.X + dx, pt.Y + dy}
}

// ReflectPoint reflects pt across the line passing through pivot
func reflectPoint(pt, pivot PointD) PointD {
	return PointD{
		X: pivot.X + (pivot.X - pt.X),
		Y: pivot.Y + (pivot.Y - pt.Y),
	}
}

// DotProduct calculates the dot product of two PointD vectors
func dotProduct(v1, v2 PointD) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

// CrossProduct calculates the 2D cross product (z-component) of two PointD vectors
func crossProductD(v1, v2 PointD) float64 {
	return v1.X*v2.Y - v1.Y*v2.X
}

// GetSegmentIntersectPtD finds the intersection point of two line segments (PointD version)
// Reference: clipper.core.h GetSegmentIntersectPt template
// Returns the intersection point and true if segments intersect, or zero point and false if they don't
func getSegmentIntersectPtD(ln1a, ln1b, ln2a, ln2b PointD) (PointD, bool) {
	ln1dy := ln1b.Y - ln1a.Y
	ln1dx := ln1a.X - ln1b.X
	ln2dy := ln2b.Y - ln2a.Y
	ln2dx := ln2a.X - ln2b.X
	det := (ln2dy * ln1dx) - (ln1dy * ln2dx)
	if det == 0.0 {
		return PointD{}, false // Parallel lines
	}

	// Calculate min/max bounds for both segments
	bb0minx := min(ln1a.X, ln1b.X)
	bb0miny := min(ln1a.Y, ln1b.Y)
	bb0maxx := max(ln1a.X, ln1b.X)
	bb0maxy := max(ln1a.Y, ln1b.Y)
	bb1minx := min(ln2a.X, ln2b.X)
	bb1miny := min(ln2a.Y, ln2b.Y)
	bb1maxx := max(ln2a.X, ln2b.X)
	bb1maxy := max(ln2a.Y, ln2b.Y)

	// Check if bounding boxes overlap
	if bb0maxx < bb1minx || bb1maxx < bb0minx ||
		bb0maxy < bb1miny || bb1maxy < bb0miny {
		return PointD{}, false
	}

	// Calculate intersection point
	c1 := (ln1dy*ln1a.X + ln1dx*ln1a.Y)
	c2 := (ln2dy*ln2a.X + ln2dx*ln2a.Y)
	ip := PointD{
		X: (c1*ln2dx - c2*ln1dx) / det,
		Y: (c1*ln2dy - c2*ln1dy) / det,
	}

	// Verify intersection point is within both segments
	if ip.X < bb0minx || ip.X > bb0maxx ||
		ip.Y < bb0miny || ip.Y > bb0maxy {
		return PointD{}, false
	}
	if ip.X < bb1minx || ip.X > bb1maxx ||
		ip.Y < bb1miny || ip.Y > bb1maxy {
		return PointD{}, false
	}

	return ip, true
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ellipse64 generates an elliptical path (or circle if radiusX == radiusY)
// Reference: clipper.h Ellipse template function
func ellipse64(center Point64, radiusX, radiusY float64, steps int) Path64 {
	if radiusX <= 0 {
		return Path64{}
	}
	if radiusY <= 0 {
		radiusY = radiusX
	}
	if steps <= 2 {
		steps = int(math.Pi * math.Sqrt((radiusX+radiusY)/2))
	}

	si := math.Sin(2 * math.Pi / float64(steps))
	co := math.Cos(2 * math.Pi / float64(steps))
	dx := co
	dy := si
	result := make(Path64, steps)
	result[0] = Point64{X: center.X + int64(radiusX), Y: center.Y}

	for i := 1; i < steps; i++ {
		result[i] = Point64{
			X: center.X + int64(radiusX*dx),
			Y: center.Y + int64(radiusY*dy),
		}
		x := dx*co - dy*si
		dy = dy*co + dx*si
		dx = x
	}

	return result
}

// negatePath negates all PointD coordinates in a path (flip direction)
// Reference: clipper.offset.cpp lines 122-131
func negatePath(norms []PointD) {
	for i := range norms {
		norms[i].X = -norms[i].X
		norms[i].Y = -norms[i].Y
	}
}
