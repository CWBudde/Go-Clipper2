package clipper

// This file contains the core type definitions for the Clipper2 polygon operations
// Includes basic types, enums, and complex algorithm-specific types

// ==============================================================================
// Core Types and Enums
// ==============================================================================

// Point64 represents a point with 64-bit integer coordinates
type Point64 struct {
	X, Y int64
}

// Path64 represents a sequence of points forming a path
type Path64 []Point64

// Paths64 represents a collection of paths
type Paths64 []Path64

// Rect64 represents a rectangle with 64-bit integer coordinates
// Matches Clipper2 C++ Rect<int64_t> structure
type Rect64 struct {
	Left, Top, Right, Bottom int64
}

// Width returns the width of the rectangle
func (r Rect64) Width() int64 {
	return r.Right - r.Left
}

// Height returns the height of the rectangle
func (r Rect64) Height() int64 {
	return r.Bottom - r.Top
}

// MidPoint returns the center point of the rectangle
func (r Rect64) MidPoint() Point64 {
	return Point64{
		X: (r.Left + r.Right) / 2,
		Y: (r.Top + r.Bottom) / 2,
	}
}

// AsPath converts a rectangle to a path (4 points in counter-clockwise order)
// Reference: clipper.core.h Rect::AsPath
func (r Rect64) AsPath() Path64 {
	return Path64{
		{X: r.Left, Y: r.Top},
		{X: r.Right, Y: r.Top},
		{X: r.Right, Y: r.Bottom},
		{X: r.Left, Y: r.Bottom},
	}
}

// Contains checks if a point is inside the rectangle (exclusive of boundaries)
func (r Rect64) Contains(pt Point64) bool {
	return pt.X > r.Left && pt.X < r.Right && pt.Y > r.Top && pt.Y < r.Bottom
}

// ContainsRect checks if another rectangle is fully contained within this rectangle
func (r Rect64) ContainsRect(other Rect64) bool {
	return other.Left >= r.Left && other.Right <= r.Right &&
		other.Top >= r.Top && other.Bottom <= r.Bottom
}

// IsEmpty returns true if the rectangle has zero or negative area
func (r Rect64) IsEmpty() bool {
	return r.Bottom <= r.Top || r.Right <= r.Left
}

// Intersects checks if this rectangle intersects with another rectangle
func (r Rect64) Intersects(other Rect64) bool {
	return max64(r.Left, other.Left) <= min64(r.Right, other.Right) &&
		max64(r.Top, other.Top) <= min64(r.Bottom, other.Bottom)
}

// ==============================================================================
// 32-bit Coordinate Types
// ==============================================================================
// These types provide API compatibility with 32-bit graphics libraries and systems.
// Internally, operations are performed in 64-bit for numerical stability, with
// automatic overflow detection when converting results back to 32-bit.

// Point32 represents a point with 32-bit integer coordinates
type Point32 struct {
	X, Y int32
}

// Path32 represents a sequence of points forming a path
type Path32 []Point32

// Paths32 represents a collection of paths
type Paths32 []Path32

// Rect32 represents a rectangle with 32-bit integer coordinates
type Rect32 struct {
	Left, Top, Right, Bottom int32
}

// Width returns the width of the rectangle
func (r Rect32) Width() int32 {
	return r.Right - r.Left
}

// Height returns the height of the rectangle
func (r Rect32) Height() int32 {
	return r.Bottom - r.Top
}

// MidPoint returns the center point of the rectangle
func (r Rect32) MidPoint() Point32 {
	return Point32{
		X: (r.Left + r.Right) / 2,
		Y: (r.Top + r.Bottom) / 2,
	}
}

// AsPath converts a rectangle to a path (4 points in counter-clockwise order)
func (r Rect32) AsPath() Path32 {
	return Path32{
		{X: r.Left, Y: r.Top},
		{X: r.Right, Y: r.Top},
		{X: r.Right, Y: r.Bottom},
		{X: r.Left, Y: r.Bottom},
	}
}

// Contains checks if a point is inside the rectangle (exclusive of boundaries)
func (r Rect32) Contains(pt Point32) bool {
	return pt.X > r.Left && pt.X < r.Right && pt.Y > r.Top && pt.Y < r.Bottom
}

// ContainsRect checks if another rectangle is fully contained within this rectangle
func (r Rect32) ContainsRect(other Rect32) bool {
	return other.Left >= r.Left && other.Right <= r.Right &&
		other.Top >= r.Top && other.Bottom <= r.Bottom
}

// IsEmpty returns true if the rectangle has zero or negative area
func (r Rect32) IsEmpty() bool {
	return r.Bottom <= r.Top || r.Right <= r.Left
}

// Intersects checks if this rectangle intersects with another rectangle
func (r Rect32) Intersects(other Rect32) bool {
	return max32(r.Left, other.Left) <= min32(r.Right, other.Right) &&
		max32(r.Top, other.Top) <= min32(r.Bottom, other.Bottom)
}

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
// Matches Clipper2 C++ enum exactly for oracle compatibility
type JoinType uint8

const (
	JoinSquare JoinType = iota // squared-off join at exactly offset distance
	JoinBevel                  // beveled join (simplest - just two offset points)
	JoinRound                  // rounded join with arc approximation
	JoinMiter                  // mitered join with miter limit
)

// EndType specifies how path ends are handled during offsetting
// Matches Clipper2 C++ enum exactly for oracle compatibility
type EndType uint8

const (
	EndPolygon EndType = iota // closed polygon (offsets only one side)
	EndJoined                 // open path with joined ends (treated as closed)
	EndButt                   // open path with square blunt ends
	EndSquare                 // open path with square extended ends
	EndRound                  // open path with round extended ends
)

// OffsetOptions contains options for path offsetting
type OffsetOptions struct {
	MiterLimit        float64 // maximum allowed miter join length (default: 2.0)
	ArcTolerance      float64 // maximum allowed deviation from true arc (default: 0.25)
	PreserveCollinear bool    // preserve collinear edges in Union cleanup (default: false)
	ReverseSolution   bool    // reverse output path orientation (default: false)
}

// ==============================================================================
// Vatti Algorithm Types
// ==============================================================================

// PathType represents the type of a path (subject or clip)
type PathType uint8

const (
	PathTypeSubject PathType = iota
	PathTypeClip
)

// Edge represents a polygon edge in the active edge list (based on Clipper2's Active struct)
type Edge struct {
	Bot        Point64 // bottom point of the edge
	Top        Point64 // top point of the edge
	CurrX      int64   // current X position (updated at every new scanline)
	Dx         float64 // delta X per unit Y (horizontal slope)
	WindDx     int     // +1 or -1 depending on winding direction
	WindCount  int     // accumulated winding count
	WindCount2 int     // accumulated winding count for clip polygons
	OutRec     *OutRec // output record this edge contributes to

	// Active Edge List (AEL) - Vatti's AET (active edge table)
	// Linked list of all edges (from left to right) that are present
	// (or 'active') within the current scanbeam
	PrevInAEL *Edge // previous edge in active edge list
	NextInAEL *Edge // next edge in active edge list

	// Sorted Edge List (SEL) - Vatti's ST (sorted table)
	// Linked list used when sorting edges into their new positions at the
	// top of scanbeams, but also (re)used to process horizontals
	PrevInSEL *Edge // previous edge in sorted edge list
	NextInSEL *Edge // next edge in sorted edge list

	Jump        *Edge        // jump pointer for efficiency
	VertexTop   *Vertex      // top vertex of this edge
	LocalMin    *LocalMinima // local minima this edge belongs to
	IsLeftBound bool         // true if this is a left bound edge
	JoinWith    JoinWith     // join specification
}

// IntersectNode represents an edge intersection that needs to be processed
// Corresponds to C++ IntersectNode (clipper.engine.h line 139)
type IntersectNode struct {
	Edge1 *Edge   // first edge in the intersection
	Edge2 *Edge   // second edge in the intersection
	Pt    Point64 // intersection point
}

// LocalMinima represents a local minimum point where edges start (aligned with Clipper2)
type LocalMinima struct {
	Vertex   *Vertex      // the vertex representing this local minimum
	PathType PathType     // subject or clip path type
	IsOpen   bool         // true if this is an open path
	Next     *LocalMinima // next local minima (sorted by Y)
}

// OutRec represents an output polygon record
type OutRec struct {
	Idx       int     // index in the output record list
	Owner     *OutRec // parent polygon for holes
	FrontEdge *Edge   // front edge (for tracking which side adds to front of list)
	BackEdge  *Edge   // back edge (for tracking which side adds to back of list)
	State     OutRecState
	Pts       *OutPt      // linked list of output points
	BottomPt  *OutPt      // bottommost point
	PolyPath  *PolyPath64 // hierarchical path structure (see polytree.go)
}

// OutRecState represents the state of an output record
type OutRecState uint8

const (
	OutRecStateUndefined OutRecState = iota
	OutRecStateOpen
	OutRecStateOuter
	OutRecStateHole
)

// OutPt represents a point in an output polygon
type OutPt struct {
	Pt   Point64 // the point coordinates
	Next *OutPt  // next point in the polygon
	Prev *OutPt  // previous point in the polygon
	Idx  int     // index for debugging
}

// Clipper64 implements the Vatti polygon clipping algorithm
type Clipper64 struct {
	_minimaList     *LocalMinima // sorted list of local minima
	_activeEdgeList *Edge        // active edge list during scanline
	_scanY          int64        // current scanline Y position
	_outRecList     []*OutRec    // list of output records
	_fillRule       FillRule     // fill rule for polygon interiors
	_clipType       ClipType     // boolean operation type
}

// ==============================================================================
// Validation Helper Functions
// ==============================================================================

// validateFillRule checks if a fill rule is valid
func validateFillRule(fillRule FillRule) error {
	if fillRule > Negative {
		return ErrInvalidFillRule
	}
	return nil
}

// validateClipType checks if a clip type is valid
func validateClipType(clipType ClipType) error {
	if clipType > Xor {
		return ErrInvalidClipType
	}
	return nil
}

// validateJoinType checks if a join type is valid
func validateJoinType(joinType JoinType) error {
	if joinType > JoinMiter {
		return ErrInvalidJoinType
	}
	return nil
}

// validateEndType checks if an end type is valid
func validateEndType(endType EndType) error {
	if endType > EndRound {
		return ErrInvalidEndType
	}
	return nil
}

// filterValidPaths removes nil and degenerate paths (< 3 points for closed paths)
// Returns filtered paths and a boolean indicating if any paths were filtered
func filterValidPaths(paths Paths64, minPoints int) (Paths64, bool) {
	if paths == nil {
		return Paths64{}, false
	}

	filtered := make(Paths64, 0, len(paths))
	hadInvalid := false

	for _, path := range paths {
		if path != nil && len(path) >= minPoints {
			filtered = append(filtered, path)
		} else if len(path) < minPoints && len(path) > 0 {
			hadInvalid = true
		}
	}

	return filtered, hadInvalid
}

// validatePath checks if a path is valid for operations requiring closed polygons
func validatePath(path Path64, minPoints int) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}
	if len(path) < minPoints {
		return ErrDegeneratePolygon
	}
	return nil
}

// filterValidPaths32 removes nil and degenerate paths (< 3 points for closed paths)
// Returns filtered paths and a boolean indicating if any paths were filtered
func filterValidPaths32(paths Paths32, minPoints int) (Paths32, bool) {
	if paths == nil {
		return Paths32{}, false
	}

	filtered := make(Paths32, 0, len(paths))
	hadInvalid := false

	for _, path := range paths {
		if path != nil && len(path) >= minPoints {
			filtered = append(filtered, path)
		} else if len(path) < minPoints && len(path) > 0 {
			hadInvalid = true
		}
	}

	return filtered, hadInvalid
}

// validatePath32 checks if a path is valid for operations requiring closed polygons
func validatePath32(path Path32, minPoints int) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}
	if len(path) < minPoints {
		return ErrDegeneratePolygon
	}
	return nil
}
