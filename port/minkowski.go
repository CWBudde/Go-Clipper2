package clipper

// minkowskiInternal is the core Minkowski algorithm that generates quadrilaterals
// from pattern and path transformations.
//
// Algorithm (reference: clipper.minkowski.h lines 20-72):
// 1. For each point in path, create a transformed copy of pattern:
//   - For sum: add path point to each pattern point
//   - For diff: subtract pattern points from path point
//
// 2. Build quadrilaterals from adjacent transformed patterns:
//   - Connect corresponding vertices between adjacent patterns
//   - Ensure correct orientation (counter-clockwise for positive area)
//
// 3. Return all quadrilaterals (union will merge them later)
//
// Parameters:
// - pattern: The shape to apply at each path point
// - path: The path along which to apply the pattern
// - isSum: true for Minkowski sum (add), false for difference (subtract)
// - isClosed: true if path is closed (connects last to first point)
func minkowskiInternal(pattern, path Path64, isSum, isClosed bool) Paths64 {
	delta := 0
	if !isClosed {
		delta = 1
	}
	patLen := len(pattern)
	pathLen := len(path)

	if patLen == 0 || pathLen == 0 {
		return Paths64{}
	}

	// Step 1: Create transformed copies of pattern for each path point
	tmp := make(Paths64, pathLen)

	if isSum {
		// Minkowski Sum: pattern[j] + path[i] for all i, j
		for i, pathPt := range path {
			path2 := make(Path64, patLen)
			for j, patternPt := range pattern {
				path2[j] = Point64{
					X: pathPt.X + patternPt.X,
					Y: pathPt.Y + patternPt.Y,
				}
			}
			tmp[i] = path2
		}
	} else {
		// Minkowski Difference: path[i] - pattern[j] for all i, j
		for i, pathPt := range path {
			path2 := make(Path64, patLen)
			for j, patternPt := range pattern {
				path2[j] = Point64{
					X: pathPt.X - patternPt.X,
					Y: pathPt.Y - patternPt.Y,
				}
			}
			tmp[i] = path2
		}
	}

	// Step 2: Build quadrilaterals from adjacent transformed patterns
	result := make(Paths64, 0, (pathLen-delta)*patLen)

	// g is the "previous" path index
	g := 0
	if isClosed {
		g = pathLen - 1 // For closed paths, connect last point to first
	}

	h := patLen - 1 // Previous pattern point index

	// For each path segment (from point i-1 to point i)
	for i := delta; i < pathLen; i++ {
		// For each pattern edge (from point j-1 to point j)
		for j := 0; j < patLen; j++ {
			// Create quadrilateral from 4 vertices:
			// tmp[g][h] - previous path, previous pattern point
			// tmp[i][h] - current path, previous pattern point
			// tmp[i][j] - current path, current pattern point
			// tmp[g][j] - previous path, current pattern point
			quad := Path64{
				tmp[g][h],
				tmp[i][h],
				tmp[i][j],
				tmp[g][j],
			}

			// Ensure correct orientation (counter-clockwise for positive area)
			if !IsPositive64(quad) {
				// Reverse if negative orientation
				for k, l := 0, len(quad)-1; k < l; k, l = k+1, l-1 {
					quad[k], quad[l] = quad[l], quad[k]
				}
			}

			result = append(result, quad)
			h = j // Move to next pattern edge
		}
		g = i // Move to next path segment
	}

	return result
}
