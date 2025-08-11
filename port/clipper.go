// Package clipper provides pure Go implementation of polygon clipping and offsetting operations.
// This is a port of the Clipper2 library (https://github.com/AngusJohnson/Clipper2).
package clipper

// Point64 represents a point with 64-bit integer coordinates
type Point64 struct {
	X, Y int64
}

// Path64 represents a sequence of points forming a path
type Path64 []Point64

// Paths64 represents a collection of paths
type Paths64 []Path64

// ClipType specifies the type of boolean operation
type ClipType uint8

const (
	Intersection ClipType = iota // intersect subject and clip polygons
	Union                        // union (OR) subject and clip polygons  
	Difference                   // subtract clip polygons from subject polygons
	Xor                          // exclusively or (XOR) subject and clip polygons
)

// FillRule specifies how to determine polygon interiors for self-intersecting polygons
type FillRule uint8

const (
	EvenOdd  FillRule = iota // odd numbered sub-regions are filled
	NonZero                  // non-zero sub-regions are filled  
	Positive                 // positive sub-regions are filled
	Negative                 // negative sub-regions are filled
)

// JoinType specifies how path segments are joined during offsetting
type JoinType uint8

const (
	Square JoinType = iota // squared-off join
	Round                  // rounded join
	Miter                  // mitered join
)

// EndType specifies how path ends are handled during offsetting
type EndType uint8

const (
	ClosedPolygon EndType = iota // end type for closed polygon paths
	ClosedLine                   // end type for closed line paths
	OpenSquare                   // end type for open paths - square end cap
	OpenRound                    // end type for open paths - round end cap
	OpenButt                     // end type for open paths - butt end cap
)

// OffsetOptions contains options for path offsetting
type OffsetOptions struct {
	MiterLimit   float64 // maximum allowed miter join length (default: 2.0)
	ArcTolerance float64 // maximum allowed deviation from true arc (default: 0.25)
}

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
func BooleanOp64(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution Paths64, solutionOpen Paths64, err error) {
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