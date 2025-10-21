// Package clipper provides pure Go implementation of polygon clipping and offsetting operations.
// This is a port of the Clipper2 library (https://github.com/AngusJohnson/Clipper2).
package clipper

// Union64 returns the union of subject and clip polygons
func Union64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Union, fillRule, subjects, nil, clips)
	return result, err
}

// Intersect64 returns the intersection of subject and clip polygons
func Intersect64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Intersection, fillRule, subjects, nil, clips)
	return result, err
}

// Difference64 returns the difference of subject and clip polygons (subject - clip)
func Difference64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Difference, fillRule, subjects, nil, clips)
	return result, err
}

// Xor64 returns the symmetric difference (XOR) of subject and clip polygons
func Xor64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Xor, fillRule, subjects, nil, clips)
	return result, err
}

// BooleanOp64 performs the specified boolean operation on the input polygons
func BooleanOp64(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution, solutionOpen Paths64, err error) {
	return booleanOp64Impl(clipType, fillRule, subjects, subjectsOpen, clips)
}

// InflatePaths64 inflates (offsets) paths by the specified delta
func InflatePaths64(paths Paths64, delta float64, joinType JoinType, endType EndType, opts ...OffsetOptions) (Paths64, error) {
	var options OffsetOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		}
	}
	return inflatePathsImpl(paths, delta, joinType, endType, options)
}

// RectClip64 clips paths against a rectangular window
func RectClip64(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) != 4 {
		return nil, ErrInvalidRectangle
	}
	return rectClipImpl(rect, paths)
}

// Area64 calculates the area of a path
func Area64(path Path64) float64 {
	return areaImpl(path)
}

// IsPositive64 returns true if the path has positive orientation (counter-clockwise)
func IsPositive64(path Path64) bool {
	return Area64(path) > 0
}

// Reverse64 reverses the order of points in a path
func Reverse64(path Path64) Path64 {
	result := make(Path64, len(path))
	for i, j := 0, len(path)-1; i < len(path); i, j = i+1, j-1 {
		result[i] = path[j]
	}
	return result
}

// ReversePaths64 reverses the order of points in each path
func ReversePaths64(paths Paths64) Paths64 {
	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = Reverse64(path)
	}
	return result
}

// Bounds64 calculates the bounding rectangle of a path
func Bounds64(path Path64) Rect64 {
	return bounds64Impl(path)
}

// BoundsPaths64 calculates the bounding rectangle of multiple paths
func BoundsPaths64(paths Paths64) Rect64 {
	return boundsPaths64Impl(paths)
}

// PointInPolygon64 determines if a point is inside, outside, or on the boundary of a polygon
func PointInPolygon64(pt Point64, polygon Path64, fillRule FillRule) PolygonLocation {
	return PointInPolygon(pt, polygon, fillRule)
}

// SimplifyPath64 simplifies a path by removing points within epsilon distance of the line between neighbors
// Uses perpendicular distance-based algorithm (Visvalingam-Whyatt inspired)
func SimplifyPath64(path Path64, epsilon float64, isClosedPath bool) Path64 {
	return simplifyPath64Impl(path, epsilon, isClosedPath)
}

// SimplifyPaths64 simplifies multiple paths
func SimplifyPaths64(paths Paths64, epsilon float64, isClosedPath bool) Paths64 {
	result := make(Paths64, 0, len(paths))
	for _, path := range paths {
		simplified := SimplifyPath64(path, epsilon, isClosedPath)
		if len(simplified) > 0 {
			result = append(result, simplified)
		}
	}
	return result
}
