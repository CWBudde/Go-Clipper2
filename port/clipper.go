// Package clipper provides pure Go implementation of polygon clipping and offsetting operations.
// This is a port of the Clipper2 library (https://github.com/AngusJohnson/Clipper2).
//
// # Overview
//
// The clipper package implements robust 2D polygon operations using 64-bit integer coordinates
// for numerical stability. It provides:
//   - Boolean operations: Union, Intersection, Difference, XOR
//   - Polygon offsetting: Expansion/contraction with various join and end types
//   - Utility functions: Area calculation, simplification, bounds, point-in-polygon tests
//   - Advanced operations: Minkowski sum/difference, hierarchical output with PolyTree
//
// # Error Handling
//
// All functions that can fail return an error as their last return value. Common errors include:
//   - ErrInvalidFillRule: Fill rule out of valid range (0-3)
//   - ErrInvalidClipType: Clip type out of valid range (0-3)
//   - ErrInvalidParameter: Invalid numeric parameter (epsilon <= 0, etc.)
//   - ErrInvalidOptions: Invalid option values (miterLimit <= 0, etc.)
//   - ErrInvalidJoinType: Join type out of valid range (0-3)
//   - ErrInvalidEndType: End type out of valid range (0-4)
//   - ErrEmptyPath: Nil or empty path where valid path required
//   - ErrDegeneratePolygon: Polygon with < 3 points
//
// # Input Validation
//
// Functions automatically filter degenerate paths (< 3 points for closed polygons, < 2 for open paths).
// Invalid enum values are detected and return appropriate errors. Empty or nil paths are handled gracefully.
//
// # Coordinate System
//
// All coordinates use 64-bit integers (int64) to avoid floating-point precision issues.
// Positive Y is typically down (screen coordinates), but the library works with any consistent orientation.
package clipper

// Union64 returns the union of subject and clip polygons.
// Combines both sets of polygons into a single result where overlapping areas are merged.
//
// Possible errors: ErrInvalidFillRule
func Union64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Union, fillRule, subjects, nil, clips)
	return result, err
}

// Intersect64 returns the intersection of subject and clip polygons.
// Returns only the areas where subject and clip polygons overlap.
//
// Possible errors: ErrInvalidFillRule
func Intersect64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Intersection, fillRule, subjects, nil, clips)
	return result, err
}

// Difference64 returns the difference of subject and clip polygons (subject - clip).
// Subtracts the clip polygon areas from the subject polygons.
//
// Possible errors: ErrInvalidFillRule
func Difference64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Difference, fillRule, subjects, nil, clips)
	return result, err
}

// Xor64 returns the symmetric difference (XOR) of subject and clip polygons.
// Returns areas that are in either subject or clip, but not in both.
//
// Possible errors: ErrInvalidFillRule
func Xor64(subjects, clips Paths64, fillRule FillRule) (Paths64, error) {
	result, _, err := BooleanOp64(Xor, fillRule, subjects, nil, clips)
	return result, err
}

// BooleanOp64 performs the specified boolean operation on the input polygons.
//
// Parameters:
//   - clipType: The boolean operation to perform (Intersection, Union, Difference, Xor)
//   - fillRule: How to determine polygon interiors (EvenOdd, NonZero, Positive, Negative)
//   - subjects: Subject paths for the operation (closed polygons)
//   - subjectsOpen: Optional open paths for clipping (can be nil)
//   - clips: Clip paths for the operation (closed polygons)
//
// Returns:
//   - solution: Resulting closed paths
//   - solutionOpen: Resulting open paths (if subjectsOpen was provided)
//   - err: Error if validation fails or operation cannot be completed
//
// Possible errors: ErrInvalidClipType, ErrInvalidFillRule
//
// Note: Degenerate paths (< 3 points) are automatically filtered out.
func BooleanOp64(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution, solutionOpen Paths64, err error) {
	// Validate clip type and fill rule
	if err := validateClipType(clipType); err != nil {
		return nil, nil, err
	}
	if err := validateFillRule(fillRule); err != nil {
		return nil, nil, err
	}

	// Filter out degenerate paths (< 3 points for closed polygons)
	subjects, _ = filterValidPaths(subjects, 3)
	clips, _ = filterValidPaths(clips, 3)

	// For open paths, we allow paths with 2+ points
	if subjectsOpen != nil {
		subjectsOpen, _ = filterValidPaths(subjectsOpen, 2)
	}

	return booleanOp64Impl(clipType, fillRule, subjects, subjectsOpen, clips)
}

// InflatePaths64 inflates (offsets) paths by the specified delta.
// Positive delta expands paths outward, negative delta shrinks them inward.
//
// Parameters:
//   - paths: Paths to offset
//   - delta: Offset distance (positive = expand, negative = shrink)
//   - joinType: How to join path segments (JoinSquare, JoinBevel, JoinRound, JoinMiter)
//   - endType: How to handle path ends (EndPolygon, EndJoined, EndButt, EndSquare, EndRound)
//   - opts: Optional offset parameters (miterLimit, arcTolerance, etc.)
//
// Possible errors: ErrInvalidJoinType, ErrInvalidEndType, ErrInvalidOptions
//
// Note: Empty paths are automatically filtered out.
func InflatePaths64(paths Paths64, delta float64, joinType JoinType, endType EndType, opts ...OffsetOptions) (Paths64, error) {
	// Validate join type and end type
	if err := validateJoinType(joinType); err != nil {
		return nil, err
	}
	if err := validateEndType(endType); err != nil {
		return nil, err
	}

	var options OffsetOptions
	if len(opts) > 0 {
		options = opts[0]
		// Validate options
		if options.MiterLimit <= 0 {
			return nil, ErrInvalidOptions
		}
		if options.ArcTolerance <= 0 {
			return nil, ErrInvalidOptions
		}
	} else {
		options = OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		}
	}

	// Filter out empty paths
	if paths == nil {
		return Paths64{}, nil
	}

	return inflatePathsImpl(paths, delta, joinType, endType, options)
}

// RectClip64 clips paths against a rectangular window.
// The rectangle must be specified as exactly 4 points.
//
// Possible errors: ErrInvalidRectangle
func RectClip64(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) != 4 {
		return nil, ErrInvalidRectangle
	}
	if paths == nil {
		return Paths64{}, nil
	}
	return rectClipImpl(rect, paths)
}

// RectClipLines64 clips open paths (lines) against a rectangular clipping region.
// Unlike RectClip64 which is designed for closed polygons, RectClipLines64 handles
// line segments that can enter and exit the rectangle multiple times, producing
// multiple output segments.
//
// Parameters:
//   - rect: Rectangle defining the clipping region (must have 4 points)
//   - paths: Open paths (lines) to clip
//
// Returns clipped line segments. Lines completely outside the rectangle are removed.
// Lines completely inside are returned unchanged. Lines that cross the rectangle
// boundaries are split into multiple segments.
//
// Possible errors: ErrInvalidRectangle
func RectClipLines64(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) != 4 {
		return nil, ErrInvalidRectangle
	}
	if paths == nil {
		return Paths64{}, nil
	}
	return rectClipLinesImpl(rect, paths)
}

// Area64 calculates the area of a path.
// Returns 0 for paths with fewer than 3 points.
// Positive area indicates counter-clockwise orientation.
func Area64(path Path64) float64 {
	return areaImpl(path)
}

// IsPositive64 returns true if the path has positive orientation (counter-clockwise).
// Returns false for paths with fewer than 3 points.
func IsPositive64(path Path64) bool {
	return Area64(path) > 0
}

// Reverse64 reverses the order of points in a path.
// Returns a new path with points in reverse order.
func Reverse64(path Path64) Path64 {
	if len(path) == 0 {
		return Path64{}
	}
	result := make(Path64, len(path))
	for i, j := 0, len(path)-1; i < len(path); i, j = i+1, j-1 {
		result[i] = path[j]
	}
	return result
}

// ReversePaths64 reverses the order of points in each path.
// Returns a new collection with each path reversed.
func ReversePaths64(paths Paths64) Paths64 {
	if len(paths) == 0 {
		return Paths64{}
	}
	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = Reverse64(path)
	}
	return result
}

// Bounds64 calculates the bounding rectangle of a path.
// Returns an empty rectangle if the path is empty.
func Bounds64(path Path64) Rect64 {
	return bounds64Impl(path)
}

// BoundsPaths64 calculates the bounding rectangle of multiple paths.
// Returns an empty rectangle if all paths are empty.
func BoundsPaths64(paths Paths64) Rect64 {
	return boundsPaths64Impl(paths)
}

// PointInPolygon64 determines if a point is inside, outside, or on the boundary of a polygon.
//
// Possible errors: ErrInvalidFillRule (returned via PolygonLocation.Error if needed)
func PointInPolygon64(pt Point64, polygon Path64, fillRule FillRule) PolygonLocation {
	return PointInPolygon(pt, polygon, fillRule)
}

// SimplifyPath64 simplifies a path by removing points within epsilon distance of the line between neighbors.
// Uses perpendicular distance-based algorithm (Visvalingam-Whyatt inspired).
//
// Parameters:
//   - path: The path to simplify
//   - epsilon: Maximum perpendicular distance threshold
//   - isClosedPath: Whether the path is closed (first point connects to last)
//
// Possible errors: ErrInvalidParameter (if epsilon <= 0)
func SimplifyPath64(path Path64, epsilon float64, isClosedPath bool) (Path64, error) {
	if epsilon <= 0 {
		return nil, ErrInvalidParameter
	}
	if len(path) == 0 {
		return Path64{}, nil
	}
	return simplifyPath64Impl(path, epsilon, isClosedPath), nil
}

// SimplifyPaths64 simplifies multiple paths.
//
// Possible errors: ErrInvalidParameter (if epsilon <= 0)
func SimplifyPaths64(paths Paths64, epsilon float64, isClosedPath bool) (Paths64, error) {
	if epsilon <= 0 {
		return nil, ErrInvalidParameter
	}
	if len(paths) == 0 {
		return Paths64{}, nil
	}

	result := make(Paths64, 0, len(paths))
	for _, path := range paths {
		simplified, _ := SimplifyPath64(path, epsilon, isClosedPath)
		if len(simplified) > 0 {
			result = append(result, simplified)
		}
	}
	return result, nil
}

// MinkowskiSum64 returns the Minkowski sum of pattern and path.
// The Minkowski sum is used for polygon dilation along a shape pattern.
// For each point in the path, the pattern is placed at that point and all
// resulting quadrilaterals are unioned together.
//
// Parameters:
//   - pattern: The shape to apply at each path point (must have at least 1 point)
//   - path: The path along which to apply the pattern (must have at least 1 point)
//   - isClosed: Whether the path is closed (connects last to first point)
//
// Use cases:
//   - Robot path planning (expanding obstacles by robot size)
//   - Collision detection (expanding objects by collision radius)
//   - Shape dilation with arbitrary patterns
//
// Possible errors: ErrEmptyPath (if pattern or path is empty)
func MinkowskiSum64(pattern, path Path64, isClosed bool) (Paths64, error) {
	if len(pattern) == 0 || len(path) == 0 {
		return nil, ErrEmptyPath
	}
	return minkowskiSum64Impl(pattern, path, isClosed)
}

// MinkowskiDiff64 returns the Minkowski difference of pattern and path.
// The Minkowski difference is used for polygon erosion along a shape pattern.
// It's the inverse operation of Minkowski sum.
//
// Parameters:
//   - pattern: The shape to subtract at each path point (must have at least 1 point)
//   - path: The path along which to apply the subtraction (must have at least 1 point)
//   - isClosed: Whether the path is closed (connects last to first point)
//
// Use cases:
//   - Inverse collision detection
//   - Shape erosion with arbitrary patterns
//   - Configuration space computation
//
// Possible errors: ErrEmptyPath (if pattern or path is empty)
func MinkowskiDiff64(pattern, path Path64, isClosed bool) (Paths64, error) {
	if len(pattern) == 0 || len(path) == 0 {
		return nil, ErrEmptyPath
	}
	return minkowskiDiff64Impl(pattern, path, isClosed)
}

// ==============================================================================
// PolyTree API - Hierarchical Output
// ==============================================================================

// Union64Tree returns the union of subject and clip polygons as a hierarchical PolyTree.
// The PolyTree preserves parent-child relationships (outer polygons, holes, islands).
//
// Returns:
//   - polytree: Hierarchical structure with closed paths
//   - openPaths: Flat slice of open paths (if any)
//   - error: Error if operation fails
//
// Possible errors: ErrInvalidFillRule
func Union64Tree(subjects, clips Paths64, fillRule FillRule) (*PolyTree64, Paths64, error) {
	return BooleanOp64Tree(Union, fillRule, subjects, clips)
}

// Intersect64Tree returns the intersection of subject and clip polygons as a PolyTree.
//
// Possible errors: ErrInvalidFillRule, ErrInvalidClipType
func Intersect64Tree(subjects, clips Paths64, fillRule FillRule) (*PolyTree64, Paths64, error) {
	return BooleanOp64Tree(Intersection, fillRule, subjects, clips)
}

// Difference64Tree returns the difference (subject - clip) as a PolyTree.
//
// Possible errors: ErrInvalidFillRule, ErrInvalidClipType
func Difference64Tree(subjects, clips Paths64, fillRule FillRule) (*PolyTree64, Paths64, error) {
	return BooleanOp64Tree(Difference, fillRule, subjects, clips)
}

// Xor64Tree returns the symmetric difference (XOR) as a PolyTree.
//
// Possible errors: ErrInvalidFillRule, ErrInvalidClipType
func Xor64Tree(subjects, clips Paths64, fillRule FillRule) (*PolyTree64, Paths64, error) {
	return BooleanOp64Tree(Xor, fillRule, subjects, clips)
}

// BooleanOp64Tree performs the specified boolean operation and returns a PolyTree.
// The PolyTree structure preserves the hierarchy of polygons and holes:
//   - Level 0: Outer polygons
//   - Level 1: Holes in outer polygons
//   - Level 2: Islands within holes
//   - Level 3: Holes in islands
//   - And so on...
//
// Use PolyPath.IsHole() to determine if a node represents a hole.
//
// Possible errors: ErrInvalidClipType, ErrInvalidFillRule
//
// Note: Degenerate paths (< 3 points) are automatically filtered out.
func BooleanOp64Tree(clipType ClipType, fillRule FillRule, subjects, clips Paths64) (*PolyTree64, Paths64, error) {
	// Validate clip type and fill rule
	if err := validateClipType(clipType); err != nil {
		return nil, nil, err
	}
	if err := validateFillRule(fillRule); err != nil {
		return nil, nil, err
	}

	// Filter out degenerate paths (< 3 points for closed polygons)
	subjects, _ = filterValidPaths(subjects, 3)
	clips, _ = filterValidPaths(clips, 3)

	return booleanOp64TreeImpl(clipType, fillRule, subjects, clips)
}

// ==============================================================================
// 32-bit Coordinate API Functions
// ==============================================================================
// These functions provide API compatibility with 32-bit graphics libraries.
// Internally, all operations are performed in 64-bit for numerical stability,
// with automatic overflow detection when converting results back to 32-bit.

// Union32 returns the union of subject and clip polygons (32-bit coordinates).
// Combines both sets of polygons into a single result where overlapping areas are merged.
//
// Possible errors: ErrInvalidFillRule, ErrInt32Overflow, ErrResultOverflow
func Union32(subjects, clips Paths32, fillRule FillRule) (Paths32, error) {
	// Convert to 64-bit
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	// Execute operation in 64-bit
	result64, err := Union64(subjects64, clips64, fillRule)
	if err != nil {
		return nil, err
	}

	// Convert back to 32-bit with overflow detection
	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// Intersect32 returns the intersection of subject and clip polygons (32-bit coordinates).
// Returns only the areas where subject and clip polygons overlap.
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Intersect32(subjects, clips Paths32, fillRule FillRule) (Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	result64, err := Intersect64(subjects64, clips64, fillRule)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// Difference32 returns the difference of subject and clip polygons (subject - clip) (32-bit coordinates).
// Subtracts the clip polygon areas from the subject polygons.
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Difference32(subjects, clips Paths32, fillRule FillRule) (Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	result64, err := Difference64(subjects64, clips64, fillRule)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// Xor32 returns the symmetric difference (XOR) of subject and clip polygons (32-bit coordinates).
// Returns areas that are in either subject or clip, but not in both.
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Xor32(subjects, clips Paths32, fillRule FillRule) (Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	result64, err := Xor64(subjects64, clips64, fillRule)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// BooleanOp32 performs the specified boolean operation on 32-bit coordinate polygons.
//
// Parameters:
//   - clipType: The boolean operation to perform (Intersection, Union, Difference, Xor)
//   - fillRule: How to determine polygon interiors (EvenOdd, NonZero, Positive, Negative)
//   - subjects: Subject paths for the operation (closed polygons)
//   - subjectsOpen: Optional open paths for clipping (can be nil)
//   - clips: Clip paths for the operation (closed polygons)
//
// Returns:
//   - solution: Resulting closed paths (32-bit)
//   - solutionOpen: Resulting open paths (32-bit)
//   - err: Error if validation fails, operation cannot be completed, or result overflows
//
// Possible errors: ErrInvalidClipType, ErrInvalidFillRule, ErrResultOverflow
func BooleanOp32(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths32) (solution, solutionOpen Paths32, err error) {
	// Convert to 64-bit
	subjects64 := Paths32ToPaths64(subjects)
	subjectsOpen64 := Paths32ToPaths64(subjectsOpen)
	clips64 := Paths32ToPaths64(clips)

	// Execute operation in 64-bit
	solution64, solutionOpen64, err := BooleanOp64(clipType, fillRule, subjects64, subjectsOpen64, clips64)
	if err != nil {
		return nil, nil, err
	}

	// Convert results back to 32-bit
	solution, err = Paths64ToPaths32(solution64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	solutionOpen, err = Paths64ToPaths32(solutionOpen64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return solution, solutionOpen, nil
}

// InflatePaths32 expands or contracts paths by the specified delta (32-bit coordinates).
//
// Possible errors: ErrInvalidJoinType, ErrInvalidEndType, ErrInvalidOptions, ErrNotImplemented, ErrResultOverflow
func InflatePaths32(paths Paths32, delta float64, joinType JoinType, endType EndType, opts ...OffsetOptions) (Paths32, error) {
	paths64 := Paths32ToPaths64(paths)

	result64, err := InflatePaths64(paths64, delta, joinType, endType, opts...)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// RectClip32 clips paths to a rectangular region (32-bit coordinates).
// The rectangle must be specified as exactly 4 points in order: (left,top), (right,top), (right,bottom), (left,bottom).
//
// Possible errors: ErrInvalidRectangle, ErrResultOverflow
func RectClip32(rect Path32, paths Paths32) (Paths32, error) {
	rect64 := Path32ToPath64(rect)
	paths64 := Paths32ToPaths64(paths)

	result64, err := RectClip64(rect64, paths64)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// RectClipLines32 clips open paths (lines) against a rectangular clipping region (32-bit coordinates).
// Unlike RectClip32 which is designed for closed polygons, RectClipLines32 handles
// line segments that can enter and exit the rectangle multiple times.
//
// Parameters:
//   - rect: Rectangle defining the clipping region (must have 4 points)
//   - paths: Open paths (lines) to clip
//
// Returns clipped line segments. Lines completely outside the rectangle are removed.
// Lines completely inside are returned unchanged. Lines that cross the rectangle
// boundaries are split into multiple segments.
//
// Possible errors: ErrInvalidRectangle, ErrResultOverflow
func RectClipLines32(rect Path32, paths Paths32) (Paths32, error) {
	rect64 := Path32ToPath64(rect)
	paths64 := Paths32ToPaths64(paths)

	result64, err := RectClipLines64(rect64, paths64)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// Area32 calculates the area of a polygon (32-bit coordinates).
// Positive area indicates counter-clockwise orientation, negative indicates clockwise.
func Area32(path Path32) float64 {
	path64 := Path32ToPath64(path)
	return Area64(path64)
}

// IsPositive32 determines if a path has positive (counter-clockwise) orientation (32-bit coordinates).
func IsPositive32(path Path32) bool {
	return Area32(path) > 0
}

// Reverse32 returns a reversed copy of the path (32-bit coordinates).
func Reverse32(path Path32) Path32 {
	if len(path) == 0 {
		return nil
	}
	result := make(Path32, len(path))
	for i, pt := range path {
		result[len(path)-1-i] = pt
	}
	return result
}

// ReversePaths32 returns reversed copies of all paths (32-bit coordinates).
func ReversePaths32(paths Paths32) Paths32 {
	if len(paths) == 0 {
		return nil
	}
	result := make(Paths32, len(paths))
	for i, path := range paths {
		result[i] = Reverse32(path)
	}
	return result
}

// Bounds32 calculates the bounding rectangle of a path (32-bit coordinates).
func Bounds32(path Path32) Rect32 {
	if len(path) == 0 {
		return Rect32{}
	}

	minX, maxX := path[0].X, path[0].X
	minY, maxY := path[0].Y, path[0].Y

	for _, pt := range path[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	return Rect32{Left: minX, Top: minY, Right: maxX, Bottom: maxY}
}

// BoundsPaths32 calculates the bounding rectangle of multiple paths (32-bit coordinates).
func BoundsPaths32(paths Paths32) Rect32 {
	if len(paths) == 0 {
		return Rect32{}
	}

	bounds := Bounds32(paths[0])
	for _, path := range paths[1:] {
		pathBounds := Bounds32(path)
		if pathBounds.Left < bounds.Left {
			bounds.Left = pathBounds.Left
		}
		if pathBounds.Top < bounds.Top {
			bounds.Top = pathBounds.Top
		}
		if pathBounds.Right > bounds.Right {
			bounds.Right = pathBounds.Right
		}
		if pathBounds.Bottom > bounds.Bottom {
			bounds.Bottom = pathBounds.Bottom
		}
	}

	return bounds
}

// SimplifyPath32 simplifies a path by removing points within epsilon distance of the line (32-bit coordinates).
//
// Possible errors: ErrInvalidParameter (epsilon <= 0), ErrResultOverflow
func SimplifyPath32(path Path32, epsilon float64, isClosedPath bool) (Path32, error) {
	path64 := Path32ToPath64(path)

	result64, err := SimplifyPath64(path64, epsilon, isClosedPath)
	if err != nil {
		return nil, err
	}

	result32, err := Path64ToPath32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// SimplifyPaths32 simplifies multiple paths (32-bit coordinates).
//
// Possible errors: ErrInvalidParameter (epsilon <= 0), ErrResultOverflow
func SimplifyPaths32(paths Paths32, epsilon float64, isClosedPath bool) (Paths32, error) {
	paths64 := Paths32ToPaths64(paths)

	result64, err := SimplifyPaths64(paths64, epsilon, isClosedPath)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// MinkowskiSum32 calculates the Minkowski sum of a pattern and a path (32-bit coordinates).
//
// Possible errors: ErrInvalidInput, ErrResultOverflow
func MinkowskiSum32(pattern, path Path32, isClosed bool) (Paths32, error) {
	pattern64 := Path32ToPath64(pattern)
	path64 := Path32ToPath64(path)

	result64, err := MinkowskiSum64(pattern64, path64, isClosed)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// MinkowskiDiff32 calculates the Minkowski difference of a pattern and a path (32-bit coordinates).
//
// Possible errors: ErrInvalidInput, ErrResultOverflow
func MinkowskiDiff32(pattern, path Path32, isClosed bool) (Paths32, error) {
	pattern64 := Path32ToPath64(pattern)
	path64 := Path32ToPath64(path)

	result64, err := MinkowskiDiff64(pattern64, path64, isClosed)
	if err != nil {
		return nil, err
	}

	result32, err := Paths64ToPaths32(result64)
	if err != nil {
		return nil, ErrResultOverflow
	}

	return result32, nil
}

// Union32Tree returns the union of subject and clip polygons as a hierarchical tree (32-bit coordinates).
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Union32Tree(subjects, clips Paths32, fillRule FillRule) (*PolyTree32, Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	tree64, openPaths64, err := Union64Tree(subjects64, clips64, fillRule)
	if err != nil {
		return nil, nil, err
	}

	tree32, err := PolyTree64To32(tree64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	openPaths32, err := Paths64ToPaths32(openPaths64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return tree32, openPaths32, nil
}

// Intersect32Tree returns the intersection of subject and clip polygons as a hierarchical tree (32-bit coordinates).
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Intersect32Tree(subjects, clips Paths32, fillRule FillRule) (*PolyTree32, Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	tree64, openPaths64, err := Intersect64Tree(subjects64, clips64, fillRule)
	if err != nil {
		return nil, nil, err
	}

	tree32, err := PolyTree64To32(tree64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	openPaths32, err := Paths64ToPaths32(openPaths64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return tree32, openPaths32, nil
}

// Difference32Tree returns the difference of subject and clip polygons as a hierarchical tree (32-bit coordinates).
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Difference32Tree(subjects, clips Paths32, fillRule FillRule) (*PolyTree32, Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	tree64, openPaths64, err := Difference64Tree(subjects64, clips64, fillRule)
	if err != nil {
		return nil, nil, err
	}

	tree32, err := PolyTree64To32(tree64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	openPaths32, err := Paths64ToPaths32(openPaths64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return tree32, openPaths32, nil
}

// Xor32Tree returns the symmetric difference (XOR) of subject and clip polygons as a hierarchical tree (32-bit coordinates).
//
// Possible errors: ErrInvalidFillRule, ErrResultOverflow
func Xor32Tree(subjects, clips Paths32, fillRule FillRule) (*PolyTree32, Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	tree64, openPaths64, err := Xor64Tree(subjects64, clips64, fillRule)
	if err != nil {
		return nil, nil, err
	}

	tree32, err := PolyTree64To32(tree64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	openPaths32, err := Paths64ToPaths32(openPaths64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return tree32, openPaths32, nil
}

// BooleanOp32Tree performs the specified boolean operation and returns a hierarchical tree (32-bit coordinates).
//
// Possible errors: ErrInvalidClipType, ErrInvalidFillRule, ErrResultOverflow
func BooleanOp32Tree(clipType ClipType, fillRule FillRule, subjects, clips Paths32) (*PolyTree32, Paths32, error) {
	subjects64 := Paths32ToPaths64(subjects)
	clips64 := Paths32ToPaths64(clips)

	tree64, openPaths64, err := BooleanOp64Tree(clipType, fillRule, subjects64, clips64)
	if err != nil {
		return nil, nil, err
	}

	tree32, err := PolyTree64To32(tree64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	openPaths32, err := Paths64ToPaths32(openPaths64)
	if err != nil {
		return nil, nil, ErrResultOverflow
	}

	return tree32, openPaths32, nil
}

// ==============================================================================
// Geometric Utility Functions
// ==============================================================================

// TranslatePath64 translates (moves) a path by the specified offset.
// All points in the path are shifted by (dx, dy).
//
// Parameters:
//   - path: The path to translate
//   - dx: Horizontal offset to add to all X coordinates
//   - dy: Vertical offset to add to all Y coordinates
//
// Returns a new path with all points translated. Returns empty path if input is empty.
func TranslatePath64(path Path64, dx, dy int64) Path64 {
	return translatePath64Impl(path, dx, dy)
}

// TranslatePaths64 translates multiple paths by the specified offset.
// All points in all paths are shifted by (dx, dy).
//
// Parameters:
//   - paths: The paths to translate
//   - dx: Horizontal offset to add to all X coordinates
//   - dy: Vertical offset to add to all Y coordinates
//
// Returns a new set of paths with all points translated. Returns empty Paths64 if input is empty.
func TranslatePaths64(paths Paths64, dx, dy int64) Paths64 {
	return translatePaths64Impl(paths, dx, dy)
}

// Ellipse64 generates an elliptical path (or circle if radiusX == radiusY).
//
// Parameters:
//   - center: Center point of the ellipse
//   - radiusX: Horizontal radius (must be > 0)
//   - radiusY: Vertical radius (if <= 0, defaults to radiusX for a circle)
//   - steps: Number of points to generate (if <= 2, uses default calculation based on radius)
//
// Returns a closed path forming an ellipse. Returns empty path if radiusX <= 0.
func Ellipse64(center Point64, radiusX, radiusY float64, steps int) Path64 {
	return ellipse64Impl(center, radiusX, radiusY, steps)
}

// ScalePath64 scales a path by the specified factor around the origin.
// All coordinates are multiplied by the scale factor.
//
// Parameters:
//   - path: The path to scale
//   - scale: Scale factor to apply
//
// Returns a new path with all points scaled. Returns empty path if input is empty.
func ScalePath64(path Path64, scale float64) Path64 {
	return scalePath64Impl(path, scale)
}

// RotatePath64 rotates a path by the specified angle around a center point.
//
// Parameters:
//   - path: The path to rotate
//   - angleRad: Rotation angle in radians (positive = counter-clockwise)
//   - center: Center point for rotation
//
// Returns a new path with all points rotated. Returns empty path if input is empty.
func RotatePath64(path Path64, angleRad float64, center Point64) Path64 {
	return rotatePath64Impl(path, angleRad, center)
}

// StarPolygon64 generates a star-shaped polygon path.
//
// Parameters:
//   - center: Center point of the star
//   - outerRadius: Radius of outer points (must be > 0)
//   - innerRadius: Radius of inner points (must be > 0)
//   - points: Number of star points (must be >= 3)
//
// Returns a closed path forming a star with alternating outer and inner points.
// Returns empty path if parameters are invalid.
func StarPolygon64(center Point64, outerRadius, innerRadius float64, points int) Path64 {
	return starPolygon64Impl(center, outerRadius, innerRadius, points)
}
