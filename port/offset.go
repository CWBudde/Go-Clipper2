package clipper

import (
	"math"
)

// offsetGroup represents a group of paths with the same join and end types
// Reference: clipper.offset.h lines 35-43, clipper.offset.cpp lines 139-162
type offsetGroup struct {
	pathsIn       Paths64
	lowestPathIdx *int
	isReversed    bool
	joinType      JoinType
	endType       EndType
}

// newOffsetGroup creates a new offset group and initializes it
// Reference: clipper.offset.cpp lines 139-162
func newOffsetGroup(paths Paths64, joinType JoinType, endType EndType) offsetGroup {
	group := offsetGroup{
		pathsIn:  make(Paths64, len(paths)),
		joinType: joinType,
		endType:  endType,
	}

	// Copy and strip duplicates from paths
	isJoined := endType == EndPolygon || endType == EndJoined
	for i, path := range paths {
		group.pathsIn[i] = stripDuplicates(path, isJoined)
	}

	// For closed polygons, determine orientation from lowest path
	if endType == EndPolygon {
		var isNegArea bool
		group.lowestPathIdx, isNegArea = getLowestClosedPathInfo(group.pathsIn)
		// If the lowermost path has negative area, the whole group is reversed
		// This is more efficient than reversing every path
		if group.lowestPathIdx != nil && isNegArea {
			group.isReversed = true
		}
	}

	return group
}

// ClipperOffset performs polygon offsetting (inflation/deflation)
// Reference: clipper.offset.h lines 32-122
type ClipperOffset struct {
	// Configuration
	miterLimit        float64
	arcTolerance      float64
	preserveCollinear bool
	reverseSolution   bool

	// Groups of paths to offset
	groups []offsetGroup

	// Working state (reused across offsetting operations)
	norms      []PointD
	pathOut    Path64
	delta      float64
	groupDelta float64
	tempLim    float64

	// For round joins (Phase 4)
	stepsPerRad float64
	stepSin     float64
	stepCos     float64
}

// NewClipperOffset creates a new ClipperOffset with the specified parameters
// Reference: clipper.offset.h lines 85-91
func NewClipperOffset(miterLimit, arcTolerance float64) *ClipperOffset {
	return &ClipperOffset{
		miterLimit:        miterLimit,
		arcTolerance:      arcTolerance,
		preserveCollinear: false,
		reverseSolution:   false,
		groups:            make([]offsetGroup, 0),
	}
}

// AddPath adds a single path to be offset
// Reference: clipper.offset.cpp lines 168-171
func (co *ClipperOffset) AddPath(path Path64, joinType JoinType, endType EndType) {
	paths := Paths64{path}
	co.groups = append(co.groups, newOffsetGroup(paths, joinType, endType))
}

// AddPaths adds multiple paths with the same join and end types
// Reference: clipper.offset.cpp lines 173-177
func (co *ClipperOffset) AddPaths(paths Paths64, joinType JoinType, endType EndType) {
	if len(paths) == 0 {
		return
	}
	co.groups = append(co.groups, newOffsetGroup(paths, joinType, endType))
}

// Clear removes all paths from the offset operation
// Reference: clipper.offset.h line 98
func (co *ClipperOffset) Clear() {
	co.groups = make([]offsetGroup, 0)
	co.norms = nil
}

// BuildNormals calculates perpendicular unit normals for each edge in the path
// Reference: clipper.offset.cpp lines 179-188
func (co *ClipperOffset) BuildNormals(path Path64) {
	co.norms = make([]PointD, 0, len(path))
	if len(path) == 0 {
		return
	}

	// Calculate normal for each edge
	for i := 0; i < len(path)-1; i++ {
		co.norms = append(co.norms, getUnitNormal(path[i], path[i+1]))
	}
	// Last edge wraps around to first vertex
	co.norms = append(co.norms, getUnitNormal(path[len(path)-1], path[0]))
}

// DoBevel creates a beveled join (simplest join type - just two offset points)
// Reference: clipper.offset.cpp lines 190-216
func (co *ClipperOffset) DoBevel(path Path64, j, k int) {
	var pt1, pt2 PointD

	if j == k {
		// Single point offset (for open path end caps)
		absDelta := math.Abs(co.groupDelta)
		pt1 = PointD{
			X: float64(path[j].X) - absDelta*co.norms[j].X,
			Y: float64(path[j].Y) - absDelta*co.norms[j].Y,
		}
		pt2 = PointD{
			X: float64(path[j].X) + absDelta*co.norms[j].X,
			Y: float64(path[j].Y) + absDelta*co.norms[j].Y,
		}
	} else {
		// Regular join between two edges
		pt1 = PointD{
			X: float64(path[j].X) + co.groupDelta*co.norms[k].X,
			Y: float64(path[j].Y) + co.groupDelta*co.norms[k].Y,
		}
		pt2 = PointD{
			X: float64(path[j].X) + co.groupDelta*co.norms[j].X,
			Y: float64(path[j].Y) + co.groupDelta*co.norms[j].Y,
		}
	}

	co.pathOut = append(co.pathOut,
		Point64{X: int64(pt1.X), Y: int64(pt1.Y)},
		Point64{X: int64(pt2.X), Y: int64(pt2.Y)})
}

// DoMiter creates a mitered join (sharp corner with miter limit control)
// Reference: clipper.offset.cpp lines 258-271
func (co *ClipperOffset) DoMiter(path Path64, j, k int, cosA float64) {
	// Calculate the miter point from averaged normals
	// q represents the distance along the averaged normal vector
	q := co.groupDelta / (cosA + 1)

	// Create miter point by averaging the two normals and scaling by q
	pt := Point64{
		X: int64(float64(path[j].X) + (co.norms[k].X+co.norms[j].X)*q),
		Y: int64(float64(path[j].Y) + (co.norms[k].Y+co.norms[j].Y)*q),
	}

	co.pathOut = append(co.pathOut, pt)
}

// DoSquare creates a squared join with perpendicular corners
// Reference: clipper.offset.cpp lines 218-256
func (co *ClipperOffset) DoSquare(path Path64, j, k int) {
	var vec PointD
	if j == k {
		// Single point - perpendicular to normal
		vec = PointD{co.norms[j].Y, -co.norms[j].X}
	} else {
		// Join - average of perpendicular vectors
		vec = getAvgUnitVector(
			PointD{-co.norms[k].Y, co.norms[k].X},
			PointD{co.norms[j].Y, -co.norms[j].X})
	}

	absDelta := math.Abs(co.groupDelta)

	// Offset the original vertex delta units along unit vector
	ptQ := PointD{float64(path[j].X), float64(path[j].Y)}
	ptQ = translatePoint(ptQ, absDelta*vec.X, absDelta*vec.Y)

	// Get perpendicular vertices
	pt1 := translatePoint(ptQ, co.groupDelta*vec.Y, co.groupDelta*-vec.X)
	pt2 := translatePoint(ptQ, co.groupDelta*-vec.Y, co.groupDelta*vec.X)

	// Get 2 vertices along one edge offset
	pt3 := getPerpendicD(path[k], co.norms[k], co.groupDelta)

	var pt PointD
	if j == k {
		// Single point case
		pt4 := PointD{pt3.X + vec.X*co.groupDelta, pt3.Y + vec.Y*co.groupDelta}
		pt = ptQ
		if ip, ok := getSegmentIntersectPtD(pt1, pt2, pt3, pt4); ok {
			pt = ip
		}
		// Get the second intersect point through reflection
		co.pathOut = append(co.pathOut,
			Point64{X: int64(reflectPoint(pt, ptQ).X), Y: int64(reflectPoint(pt, ptQ).Y)},
			Point64{X: int64(pt.X), Y: int64(pt.Y)})
	} else {
		// Regular join case
		pt4 := getPerpendicD(path[j], co.norms[k], co.groupDelta)
		pt = ptQ
		if ip, ok := getSegmentIntersectPtD(pt1, pt2, pt3, pt4); ok {
			pt = ip
		}
		co.pathOut = append(co.pathOut,
			Point64{X: int64(pt.X), Y: int64(pt.Y)},
			Point64{X: int64(reflectPoint(pt, ptQ).X), Y: int64(reflectPoint(pt, ptQ).Y)})
	}
}

// DoRound creates a rounded join with arc approximation
// Reference: clipper.offset.cpp lines 273-309
func (co *ClipperOffset) DoRound(path Path64, j, k int, angle float64) {
	pt := path[j]
	offsetVec := PointD{
		X: co.norms[k].X * co.groupDelta,
		Y: co.norms[k].Y * co.groupDelta,
	}

	// Special case: single point offset (for open path end caps)
	if j == k {
		offsetVec.Negate()
	}

	// Add first point
	co.pathOut = append(co.pathOut, Point64{
		X: pt.X + int64(offsetVec.X),
		Y: pt.Y + int64(offsetVec.Y),
	})

	// Calculate number of steps for arc approximation
	steps := int(math.Ceil(co.stepsPerRad * math.Abs(angle)))

	// Generate arc points using rotation matrix
	for i := 1; i < steps; i++ {
		// Rotate offset vector: new = (x*cos - sin*y, x*sin + y*cos)
		// Store old X to avoid using updated value
		oldX := offsetVec.X
		offsetVec.X = offsetVec.X*co.stepCos - co.stepSin*offsetVec.Y
		offsetVec.Y = oldX*co.stepSin + offsetVec.Y*co.stepCos
		co.pathOut = append(co.pathOut, Point64{
			X: pt.X + int64(offsetVec.X),
			Y: pt.Y + int64(offsetVec.Y),
		})
	}

	// Add final perpendicular point
	co.pathOut = append(co.pathOut, getPerpendic(path[j], co.norms[j], co.groupDelta))
}

// OffsetPoint processes a single vertex, determining the appropriate join geometry
// Phase 4: Supports Bevel, Miter, Square, and Round joins for closed polygons
// Reference: clipper.offset.cpp lines 311-370
func (co *ClipperOffset) OffsetPoint(group *offsetGroup, path Path64, j, k int) {
	// Skip if the two adjacent points are the same
	if path[j] == path[k] {
		return
	}

	// Calculate sin and cos of angle between edges
	sinA := crossProductD(co.norms[j], co.norms[k])
	cosA := dotProduct(co.norms[j], co.norms[k])

	// Clamp sinA to [-1, 1]
	if sinA > 1.0 {
		sinA = 1.0
	} else if sinA < -1.0 {
		sinA = -1.0
	}

	// Check for near-zero delta
	if math.Abs(co.groupDelta) <= floatingPointTolerance {
		co.pathOut = append(co.pathOut, path[j])
		return
	}

	// Test for concavity: cos(A) > -0.999 and sin(A) * delta < 0
	if cosA > -0.999 && (sinA*co.groupDelta < 0) {
		// Concave join - insert 3 points that create negative regions
		// These will be removed by the Union operation at the end
		co.pathOut = append(co.pathOut,
			getPerpendic(path[j], co.norms[k], co.groupDelta),
			path[j],
			getPerpendic(path[j], co.norms[j], co.groupDelta))
	} else if cosA > 0.999 && group.joinType != JoinRound {
		// Almost straight - less than 2.5 degrees
		// Use miter for near-collinear edges (all join types except Round)
		co.DoMiter(path, j, k, cosA)
	} else if group.joinType == JoinRound {
		// Round join - calculate angle and generate arc
		angle := math.Atan2(sinA, cosA)
		co.DoRound(path, j, k, angle)
	} else if group.joinType == JoinMiter {
		// Miter join - check if angle is within miter limit
		if cosA > co.tempLim-1 {
			// Angle is acceptable for miter
			co.DoMiter(path, j, k, cosA)
		} else {
			// Miter limit exceeded - fall back to square
			co.DoSquare(path, j, k)
		}
	} else if group.joinType == JoinBevel {
		co.DoBevel(path, j, k)
	} else {
		// Square join (default)
		co.DoSquare(path, j, k)
	}
}

// OffsetPolygon offsets a closed polygon path
// Reference: clipper.offset.cpp lines 372-378
func (co *ClipperOffset) OffsetPolygon(group *offsetGroup, path Path64) {
	co.pathOut = make(Path64, 0, len(path)*2)
	for j := 0; j < len(path); j++ {
		k := j - 1
		if k < 0 {
			k = len(path) - 1
		}
		co.OffsetPoint(group, path, j, k)
	}
}

// OffsetOpenJoined offsets an open path by treating it as a closed polygon
// Reference: clipper.offset.cpp lines 380-393
func (co *ClipperOffset) OffsetOpenJoined(group *offsetGroup, path Path64) Paths64 {
	solution := make(Paths64, 0, 1)

	// Offset as polygon
	co.OffsetPolygon(group, path)

	// Make a copy of the path and reverse it
	reversePath := make(Path64, len(path))
	for i := 0; i < len(path); i++ {
		reversePath[i] = path[len(path)-1-i]
	}

	// Rebuild normals: reverse the normals array
	for i, j := 0, len(co.norms)-1; i < j; i, j = i+1, j-1 {
		co.norms[i], co.norms[j] = co.norms[j], co.norms[i]
	}
	// Rotate: move last element to front
	if len(co.norms) > 0 {
		lastNorm := co.norms[len(co.norms)-1]
		copy(co.norms[1:], co.norms[:len(co.norms)-1])
		co.norms[0] = lastNorm
	}
	// Negate all normals
	negatePath(co.norms)

	// Offset the reversed path
	co.OffsetPolygon(group, reversePath)
	if len(co.pathOut) > 0 {
		solution = append(solution, co.pathOut)
	}

	return solution
}

// OffsetOpenPath offsets an open path with end caps
// Reference: clipper.offset.cpp lines 395-453
func (co *ClipperOffset) OffsetOpenPath(group *offsetGroup, path Path64) Paths64 {
	solution := make(Paths64, 0, 1)
	co.pathOut = make(Path64, 0, len(path)*2)

	highI := len(path) - 1

	// Do the line start cap
	if math.Abs(co.groupDelta) <= floatingPointTolerance {
		co.pathOut = append(co.pathOut, path[0])
	} else {
		switch group.endType {
		case EndButt:
			co.DoBevel(path, 0, 0)
		case EndRound:
			co.DoRound(path, 0, 0, math.Pi)
		default: // EndSquare
			co.DoSquare(path, 0, 0)
		}
	}

	// Offset the left side going forward
	for j := 1; j < highI; j++ {
		k := j - 1
		co.OffsetPoint(group, path, j, k)
	}

	// Reverse normals for return path
	for i := highI; i > 0; i-- {
		co.norms[i] = PointD{
			X: -co.norms[i-1].X,
			Y: -co.norms[i-1].Y,
		}
	}
	co.norms[0] = co.norms[highI]

	// Do the line end cap
	if math.Abs(co.groupDelta) <= floatingPointTolerance {
		co.pathOut = append(co.pathOut, path[highI])
	} else {
		switch group.endType {
		case EndButt:
			co.DoBevel(path, highI, highI)
		case EndRound:
			co.DoRound(path, highI, highI, math.Pi)
		default: // EndSquare
			co.DoSquare(path, highI, highI)
		}
	}

	// Offset the right side going backward
	for j := highI - 1; j > 0; j-- {
		k := j + 1
		co.OffsetPoint(group, path, j, k)
	}

	if len(co.pathOut) > 0 {
		solution = append(solution, co.pathOut)
	}

	return solution
}

// DoGroupOffset processes a single group of paths
// Reference: clipper.offset.cpp lines 455-540
func (co *ClipperOffset) DoGroupOffset(group *offsetGroup) Paths64 {
	solution := make(Paths64, 0)

	// Determine group delta based on end type
	if group.endType == EndPolygon {
		// For closed polygons, respect the sign of delta
		if group.lowestPathIdx == nil {
			co.delta = math.Abs(co.delta)
		}
		if group.isReversed {
			co.groupDelta = -co.delta
		} else {
			co.groupDelta = co.delta
		}
	} else {
		co.groupDelta = math.Abs(co.delta)
	}

	absDelta := math.Abs(co.groupDelta)

	// Calculate arc tolerance and rotation constants if using round joins or round end caps
	// Reference: clipper.offset.cpp lines 471-486
	if group.joinType == JoinRound || group.endType == EndRound {
		var arcTol float64
		if co.arcTolerance > floatingPointTolerance {
			arcTol = math.Min(absDelta, co.arcTolerance)
		} else {
			arcTol = absDelta * arcConst
		}

		stepsPerRad360 := math.Min(math.Pi/math.Acos(1-arcTol/absDelta), absDelta*math.Pi)
		co.stepSin = math.Sin(2 * math.Pi / stepsPerRad360)
		co.stepCos = math.Cos(2 * math.Pi / stepsPerRad360)
		if co.groupDelta < 0.0 {
			co.stepSin = -co.stepSin
		}
		co.stepsPerRad = stepsPerRad360 / (2 * math.Pi)
	}

	// Process each path in the group
	for _, path := range group.pathsIn {
		pathLen := len(path)

		if pathLen == 0 {
			continue
		}

		// Handle single-point paths (generate circle or square)
		// Reference: clipper.offset.cpp lines 495-528
		if pathLen == 1 {
			if co.groupDelta < 1 {
				continue
			}
			pt := path[0]

			// Single vertex - build a circle (Round) or square (other join types)
			if group.joinType == JoinRound {
				radius := absDelta
				var steps int
				if co.stepsPerRad > 0 {
					steps = int(math.Ceil(co.stepsPerRad * 2 * math.Pi))
				}
				co.pathOut = ellipse64(pt, radius, radius, steps)
			} else {
				d := int64(math.Ceil(absDelta))
				r := Rect64{
					Left:   pt.X - d,
					Top:    pt.Y - d,
					Right:  pt.X + d,
					Bottom: pt.Y + d,
				}
				co.pathOut = r.AsPath()
			}

			if len(co.pathOut) > 0 {
				solution = append(solution, co.pathOut)
			}
			continue
		}

		// Handle two-point paths with Joined end type
		// Reference: clipper.offset.cpp lines 530-533
		endType := group.endType
		if pathLen == 2 && group.endType == EndJoined {
			if group.joinType == JoinRound {
				endType = EndRound
			} else {
				endType = EndSquare
			}
		}

		// Build normals for this path
		co.BuildNormals(path)

		// Route to appropriate offsetting method
		if endType == EndPolygon {
			co.OffsetPolygon(group, path)
			if len(co.pathOut) > 0 {
				solution = append(solution, co.pathOut)
			}
		} else if endType == EndJoined {
			result := co.OffsetOpenJoined(group, path)
			solution = append(solution, result...)
		} else {
			// EndButt, EndSquare, EndRound
			result := co.OffsetOpenPath(group, path)
			solution = append(solution, result...)
		}
	}

	return solution
}

// ExecuteInternal performs the actual offset operation
// Reference: clipper.offset.cpp lines 574-634
func (co *ClipperOffset) ExecuteInternal(delta float64) (Paths64, error) {
	solution := make(Paths64, 0)

	if len(co.groups) == 0 {
		return solution, nil
	}

	co.delta = delta

	// Handle insignificant offset (< 0.5 pixels)
	if math.Abs(delta) < 0.5 {
		// Just return input paths unchanged
		for _, group := range co.groups {
			solution = append(solution, group.pathsIn...)
		}
		return solution, nil
	}

	// Calculate temp_lim for miter joins (Phase 2)
	if co.miterLimit <= 1 {
		co.tempLim = 2.0
	} else {
		co.tempLim = 2.0 / (co.miterLimit * co.miterLimit)
	}

	// Arc tolerance and rotation constants are now calculated per-group in DoGroupOffset
	// This allows each group to use appropriate settings based on its join/end types

	// Process each group
	for i := range co.groups {
		result := co.DoGroupOffset(&co.groups[i])
		solution = append(solution, result...)
	}

	if len(solution) == 0 {
		return solution, nil
	}

	// Determine if paths need to be reversed for the Union operation
	pathsReversed := false
	for _, group := range co.groups {
		if group.endType == EndPolygon {
			pathsReversed = group.isReversed
			break
		}
	}

	// Clean up self-intersections using Union operation
	// This removes negative regions created by concave joins
	var fillRule FillRule
	if pathsReversed {
		fillRule = Negative
	} else {
		fillRule = Positive
	}

	// TODO(Phase 6): Pass preserve_collinear to Union operation
	// The C++ implementation calls c.PreserveCollinear(preserve_collinear_) before Execute
	// This requires extending the Union64 API or using a lower-level interface that supports
	// preserve_collinear. For now, preserve_collinear is stored but not used in pure Go mode.
	// In oracle mode (-tags=clipper_cgo), this would need to be passed through the C bridge.

	// Use Union to clean up the offset paths
	// Use the public API so it routes to oracle when appropriate
	cleanedSolution, err := Union64(solution, nil, fillRule)
	if err != nil {
		return nil, err
	}

	// Apply reverse_solution if needed
	if co.reverseSolution != pathsReversed {
		for i := range cleanedSolution {
			reversePath(cleanedSolution[i])
		}
	}

	return cleanedSolution, nil
}

// Execute performs the offset operation with the specified delta
// Reference: clipper.offset.cpp lines 636-642
func (co *ClipperOffset) Execute(delta float64) (Paths64, error) {
	return co.ExecuteInternal(delta)
}

// MiterLimit returns the current miter limit value
func (co *ClipperOffset) MiterLimit() float64 {
	return co.miterLimit
}

// SetMiterLimit sets the miter limit with validation
// Values <= 1 will be clamped to 2.0 (minimum reasonable value)
func (co *ClipperOffset) SetMiterLimit(limit float64) {
	if limit <= 1.0 {
		co.miterLimit = 2.0
	} else {
		co.miterLimit = limit
	}
}

// ArcTolerance returns the current arc tolerance value
func (co *ClipperOffset) ArcTolerance() float64 {
	return co.arcTolerance
}

// SetArcTolerance sets the arc tolerance for round joins
// Values <= 0 will use a default based on offset delta
func (co *ClipperOffset) SetArcTolerance(tolerance float64) {
	co.arcTolerance = tolerance
}

// PreserveCollinear returns the current preserve_collinear setting
func (co *ClipperOffset) PreserveCollinear() bool {
	return co.preserveCollinear
}

// SetPreserveCollinear sets whether collinear edges should be preserved
// When true, the Union cleanup operation will retain collinear vertices
// Reference: clipper.offset.h line 111-112
func (co *ClipperOffset) SetPreserveCollinear(preserve bool) {
	co.preserveCollinear = preserve
}

// ReverseSolution returns the current reverse_solution setting
func (co *ClipperOffset) ReverseSolution() bool {
	return co.reverseSolution
}

// SetReverseSolution sets whether the output path orientation should be reversed
// Reference: clipper.offset.h line 114-115
func (co *ClipperOffset) SetReverseSolution(reverse bool) {
	co.reverseSolution = reverse
}

// Helper function to reverse a path in-place
func reversePath(path Path64) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}

// Helper function to strip duplicate points from a path
// Reference: Clipper2 uses StripDuplicates function
func stripDuplicates(path Path64, isClosed bool) Path64 {
	if len(path) == 0 {
		return path
	}

	result := make(Path64, 0, len(path))
	result = append(result, path[0])

	for i := 1; i < len(path); i++ {
		if path[i] != path[i-1] {
			result = append(result, path[i])
		}
	}

	// For closed paths, also check if last point equals first
	if isClosed && len(result) > 1 && result[len(result)-1] == result[0] {
		result = result[:len(result)-1]
	}

	return result
}
