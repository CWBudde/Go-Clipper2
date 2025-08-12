//go:build !clipper_cgo

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

// ==============================================================================
// Vatti Algorithm Types
// ==============================================================================

// PathType represents the type of a path (subject or clip)
type PathType uint8

const (
	PathTypeSubject PathType = iota
	PathTypeClip
)

// Edge represents a polygon edge in the active edge list
type Edge struct {
	Bot        Point64      // bottom point of the edge
	Top        Point64      // top point of the edge
	Curr       Point64      // current position during scanline processing
	Dx         float64      // delta X per unit Y (horizontal slope)
	WindDelta  int          // +1 or -1 depending on edge direction
	WindCount  int          // accumulated winding count
	WindCount2 int          // accumulated winding count for clip polygons
	OutRec     *OutRec      // output record this edge contributes to
	Next       *Edge        // next edge in active edge list
	Prev       *Edge        // previous edge in active edge list
	NextInLML  *Edge        // next edge in local minima list
	PathType   PathType     // subject or clip path
	LocalMin   *LocalMinima // local minima this edge belongs to
}

// LocalMinima represents a local minimum point where edges start
type LocalMinima struct {
	Y          int64        // Y coordinate of the local minimum
	LeftBound  *Edge        // leftmost edge starting at this minimum
	RightBound *Edge        // rightmost edge starting at this minimum
	Next       *LocalMinima // next local minima (sorted by Y)
}

// OutRec represents an output polygon record
type OutRec struct {
	Idx      int     // index in the output record list
	Owner    *OutRec // parent polygon for holes
	State    OutRecState
	Pts      *OutPt    // linked list of output points
	BottomPt *OutPt    // bottommost point
	PolyPath *PolyPath // hierarchical path structure
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

// PolyPath represents a hierarchical polygon path structure
type PolyPath struct {
	Path     Path64      // the polygon path
	Children []*PolyPath // child paths (holes)
	Parent   *PolyPath   // parent path
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
