//go:build clipper_cgo

package clipper

import (
	"math"

	"github.com/go-clipper/clipper2/capi"
)

// convertToCAPI converts port types to capi types
func pathsToCAPI(paths Paths64) capi.Paths64 {
	result := make(capi.Paths64, len(paths))
	for i, path := range paths {
		capiPath := make(capi.Path64, len(path))
		for j, pt := range path {
			capiPath[j] = capi.Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = capiPath
	}
	return result
}

// convertFromCAPI converts capi types to port types
func pathsFromCAPI(paths capi.Paths64) Paths64 {
	result := make(Paths64, len(paths))
	for i, path := range paths {
		portPath := make(Path64, len(path))
		for j, pt := range path {
			portPath[j] = Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = portPath
	}
	return result
}

// booleanOp64Impl delegates to the CGO oracle implementation
func booleanOp64Impl(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution Paths64, solutionOpen Paths64, err error) {
	capiSubjects := pathsToCAPI(subjects)
	capiSubjectsOpen := pathsToCAPI(subjectsOpen)
	capiClips := pathsToCAPI(clips)

	capiSolution, capiSolutionOpen, err := capi.BooleanOp64(
		uint8(clipType),
		uint8(fillRule),
		capiSubjects,
		capiSubjectsOpen,
		capiClips,
	)
	if err != nil {
		return nil, nil, err
	}

	solution = pathsFromCAPI(capiSolution)
	solutionOpen = pathsFromCAPI(capiSolutionOpen)
	return solution, solutionOpen, nil
}

// booleanOp64TreeImpl performs boolean operations and returns a hierarchical PolyTree
// For CGO oracle mode, we use the flat paths and construct a tree from them.
// TODO: Could be optimized by using C++ BuildTree64 directly via CGO
func booleanOp64TreeImpl(clipType ClipType, fillRule FillRule, subjects, clips Paths64) (*PolyTree64, Paths64, error) {
	// For now, use the pure Go tree building on top of CGO results
	// This is simpler than extending CGO bridge with full tree support
	engine := NewVattiEngine(clipType, fillRule)

	// Execute clipping to populate output records
	_, _, err := engine.ExecuteClipping(subjects, nil, clips)
	if err != nil {
		return nil, nil, err
	}

	// Build tree from output records using pure Go algorithm
	polytree := NewPolyTree64()
	var openPaths Paths64
	engine.BuildTree64(polytree, &openPaths)

	return polytree, openPaths, nil
}

// inflatePathsImpl delegates to the CGO oracle implementation
func inflatePathsImpl(paths Paths64, delta float64, joinType JoinType, endType EndType, opts OffsetOptions) (Paths64, error) {
	capiPaths := pathsToCAPI(paths)
	capiResult, err := capi.InflatePaths64(capiPaths, delta, uint8(joinType), uint8(endType), opts.MiterLimit, opts.ArcTolerance)
	if err != nil {
		return nil, err
	}
	return pathsFromCAPI(capiResult), nil
}

// rectClipImpl delegates to the CGO oracle implementation
func rectClipImpl(rect Path64, paths Paths64) (Paths64, error) {
	capiRect := make(capi.Path64, len(rect))
	for i, pt := range rect {
		capiRect[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	capiPaths := pathsToCAPI(paths)
	capiResult, err := capi.RectClip64(capiRect, capiPaths)
	if err != nil {
		return nil, err
	}
	return pathsFromCAPI(capiResult), nil
}

// rectClipLinesImpl delegates to the CGO oracle implementation
func rectClipLinesImpl(rect Path64, paths Paths64) (Paths64, error) {
	capiRect := make(capi.Path64, len(rect))
	for i, pt := range rect {
		capiRect[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	capiPaths := pathsToCAPI(paths)
	capiResult, err := capi.RectClipLines64(capiRect, capiPaths)
	if err != nil {
		return nil, err
	}
	return pathsFromCAPI(capiResult), nil
}

// areaImpl calculates area using basic polygon area formula
func areaImpl(path Path64) float64 {
	if len(path) < 3 {
		return 0.0
	}

	area := 0.0
	for i := 0; i < len(path); i++ {
		j := (i + 1) % len(path)
		area += float64(path[i].X * path[j].Y)
		area -= float64(path[j].X * path[i].Y)
	}
	return area / 2.0
}

// minkowskiSum64Impl delegates to the CGO oracle implementation
func minkowskiSum64Impl(pattern, path Path64, isClosed bool) (Paths64, error) {
	// Convert port types to capi types
	capiPattern := make(capi.Path64, len(pattern))
	for i, pt := range pattern {
		capiPattern[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	capiPath := make(capi.Path64, len(path))
	for i, pt := range path {
		capiPath[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}

	// Call CGO implementation
	capiResult, err := capi.MinkowskiSum64(capiPattern, capiPath, isClosed)
	if err != nil {
		return nil, err
	}

	// Convert back to port types
	result := make(Paths64, len(capiResult))
	for i, capiPath := range capiResult {
		portPath := make(Path64, len(capiPath))
		for j, pt := range capiPath {
			portPath[j] = Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = portPath
	}

	return result, nil
}

// minkowskiDiff64Impl delegates to the CGO oracle implementation
func minkowskiDiff64Impl(pattern, path Path64, isClosed bool) (Paths64, error) {
	// Convert port types to capi types
	capiPattern := make(capi.Path64, len(pattern))
	for i, pt := range pattern {
		capiPattern[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	capiPath := make(capi.Path64, len(path))
	for i, pt := range path {
		capiPath[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}

	// Call CGO implementation
	capiResult, err := capi.MinkowskiDiff64(capiPattern, capiPath, isClosed)
	if err != nil {
		return nil, err
	}

	// Convert back to port types
	result := make(Paths64, len(capiResult))
	for i, capiPath := range capiResult {
		portPath := make(Path64, len(capiPath))
		for j, pt := range capiPath {
			portPath[j] = Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = portPath
	}

	return result, nil
}

// bounds64Impl calculates the bounding rectangle of a single path
func bounds64Impl(path Path64) Rect64 {
	if len(path) == 0 {
		return Rect64{} // Empty rectangle
	}

	rect := Rect64{
		Left:   path[0].X,
		Top:    path[0].Y,
		Right:  path[0].X,
		Bottom: path[0].Y,
	}

	for _, pt := range path[1:] {
		if pt.X < rect.Left {
			rect.Left = pt.X
		}
		if pt.X > rect.Right {
			rect.Right = pt.X
		}
		if pt.Y < rect.Top {
			rect.Top = pt.Y
		}
		if pt.Y > rect.Bottom {
			rect.Bottom = pt.Y
		}
	}

	return rect
}

// boundsPaths64Impl calculates the bounding rectangle of multiple paths
func boundsPaths64Impl(paths Paths64) Rect64 {
	if len(paths) == 0 {
		return Rect64{} // Empty rectangle
	}

	const maxInt64 = int64(^uint64(0) >> 1)
	const minInt64 = -maxInt64 - 1

	rect := Rect64{
		Left:   maxInt64,
		Top:    maxInt64,
		Right:  minInt64,
		Bottom: minInt64,
	}

	hasPoints := false
	for _, path := range paths {
		for _, pt := range path {
			hasPoints = true
			if pt.X < rect.Left {
				rect.Left = pt.X
			}
			if pt.X > rect.Right {
				rect.Right = pt.X
			}
			if pt.Y < rect.Top {
				rect.Top = pt.Y
			}
			if pt.Y > rect.Bottom {
				rect.Bottom = pt.Y
			}
		}
	}

	if !hasPoints {
		return Rect64{} // No points found
	}

	return rect
}

// simplifyPath64Impl simplifies a path using perpendicular distance algorithm
func simplifyPath64Impl(path Path64, epsilon float64, isClosedPath bool) Path64 {
	pathLen := len(path)
	if pathLen < 4 {
		return append(Path64{}, path...) // Return copy of path
	}

	epsSqr := epsilon * epsilon
	flags := make([]bool, pathLen)
	distSqr := make([]float64, pathLen)

	high := pathLen - 1

	// Calculate initial perpendicular distances
	if isClosedPath {
		distSqr[0] = perpendicDistFromLineSqrd(path[0], path[high], path[1])
		distSqr[high] = perpendicDistFromLineSqrd(path[high], path[0], path[high-1])
	} else {
		// For open paths, endpoints are always kept
		const maxFloat64 = 1.7976931348623157e+308
		distSqr[0] = maxFloat64
		distSqr[high] = maxFloat64
	}

	for i := 1; i < high; i++ {
		distSqr[i] = perpendicDistFromLineSqrd(path[i], path[i-1], path[i+1])
	}

	// Iteratively remove points with smallest distances
	curr := 0
	for {
		// Find first point with distance <= epsilon (skip points > epsilon)
		if distSqr[curr] > epsSqr {
			start := curr
			for {
				curr = getNext(curr, high, flags)
				if curr == start {
					// All remaining points have distance > epsilon
					goto done
				}
				if distSqr[curr] <= epsSqr {
					break
				}
			}
		}

		prior := getPrior(curr, high, flags)
		next := getNext(curr, high, flags)

		if next == prior {
			break // Only two points left
		}

		// Flag the point with smaller distance for removal
		var prior2 int
		if distSqr[next] < distSqr[curr] {
			prior2 = prior
			prior = curr
			curr = next
			next = getNext(next, high, flags)
		} else {
			prior2 = getPrior(prior, high, flags)
		}

		flags[curr] = true
		curr = next
		next = getNext(next, high, flags)

		// Update distances for affected neighbors
		if isClosedPath || (curr != high && curr != 0) {
			distSqr[curr] = perpendicDistFromLineSqrd(path[curr], path[prior], path[next])
		}
		if isClosedPath || (prior != 0 && prior != high) {
			distSqr[prior] = perpendicDistFromLineSqrd(path[prior], path[prior2], path[curr])
		}
	}

done:

	// Build result path from non-flagged points
	result := make(Path64, 0, pathLen)
	for i := 0; i < pathLen; i++ {
		if !flags[i] {
			result = append(result, path[i])
		}
	}

	return result
}

// perpendicDistFromLineSqrd calculates squared perpendicular distance from point to line
func perpendicDistFromLineSqrd(pt, line1, line2 Point64) float64 {
	a := float64(pt.X - line1.X)
	b := float64(pt.Y - line1.Y)
	c := float64(line2.X - line1.X)
	d := float64(line2.Y - line1.Y)

	if c == 0 && d == 0 {
		return 0
	}

	return (a*d - c*b) * (a*d - c*b) / (c*c + d*d)
}

// getNext finds the next non-flagged index (wraps around for closed paths)
func getNext(curr, high int, flags []bool) int {
	next := curr + 1
	if next > high {
		next = 0
	}

	for flags[next] {
		next++
		if next > high {
			next = 0
		}
		if next == curr {
			break // Prevent infinite loop
		}
	}

	return next
}

// getPrior finds the previous non-flagged index (wraps around for closed paths)
func getPrior(curr, high int, flags []bool) int {
	prior := curr - 1
	if prior < 0 {
		prior = high
	}

	for flags[prior] {
		prior--
		if prior < 0 {
			prior = high
		}
		if prior == curr {
			break // Prevent infinite loop
		}
	}

	return prior
}

// ==============================================================================
// Geometric Utility Functions Implementation
// ==============================================================================

func translatePath64Impl(path Path64, dx, dy int64) Path64 {
	if len(path) == 0 {
		return Path64{}
	}

	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{X: pt.X + dx, Y: pt.Y + dy}
	}
	return result
}

func translatePaths64Impl(paths Paths64, dx, dy int64) Paths64 {
	if len(paths) == 0 {
		return Paths64{}
	}

	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = translatePath64Impl(path, dx, dy)
	}
	return result
}

func ellipse64Impl(center Point64, radiusX, radiusY float64, steps int) Path64 {
	return ellipse64(center, radiusX, radiusY, steps)
}

func scalePath64Impl(path Path64, scale float64) Path64 {
	if len(path) == 0 {
		return Path64{}
	}

	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point64{
			X: int64(float64(pt.X) * scale),
			Y: int64(float64(pt.Y) * scale),
		}
	}
	return result
}

func rotatePath64Impl(path Path64, angleRad float64, center Point64) Path64 {
	if len(path) == 0 {
		return Path64{}
	}

	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	result := make(Path64, len(path))
	for i, pt := range path {
		// Translate to origin
		dx := float64(pt.X - center.X)
		dy := float64(pt.Y - center.Y)

		// Rotate
		newX := dx*cosA - dy*sinA
		newY := dx*sinA + dy*cosA

		// Translate back
		result[i] = Point64{
			X: center.X + int64(newX+0.5),
			Y: center.Y + int64(newY+0.5),
		}
	}
	return result
}

func starPolygon64Impl(center Point64, outerRadius, innerRadius float64, points int) Path64 {
	if points < 3 || outerRadius <= 0 || innerRadius <= 0 {
		return Path64{}
	}

	result := make(Path64, points*2)
	angleStep := 2 * math.Pi / float64(points)

	for i := 0; i < points; i++ {
		angle := float64(i)*angleStep - math.Pi/2 // Start at top

		// Outer point
		result[i*2] = Point64{
			X: center.X + int64(outerRadius*math.Cos(angle)+0.5),
			Y: center.Y + int64(outerRadius*math.Sin(angle)+0.5),
		}

		// Inner point (halfway between current and next outer point)
		innerAngle := angle + angleStep/2
		result[i*2+1] = Point64{
			X: center.X + int64(innerRadius*math.Cos(innerAngle)+0.5),
			Y: center.Y + int64(innerRadius*math.Sin(innerAngle)+0.5),
		}
	}

	return result
}
