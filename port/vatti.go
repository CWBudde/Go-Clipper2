//go:build !clipper_cgo

package clipper

// This file contains the Vatti polygon clipping algorithm implementation
// Extracted from impl_pure.go for better organization

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
	minimaList     *LocalMinima // sorted list of local minima
	activeEdgeList *Edge        // active edge list during scanline
	scanY          int64        // current scanline Y position
	outRecList     []*OutRec    // list of output records
	fillRule       FillRule     // fill rule for polygon interiors
	clipType       ClipType     // boolean operation type
}