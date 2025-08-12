//go:build !clipper_cgo

package clipper

// This file contains rectangle clipping implementation using Sutherland-Hodgman algorithm
// Separated for better code organization

// rectClipImpl pure Go implementation using Sutherland-Hodgman style clipping
func rectClipImpl(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) != 4 {
		return nil, ErrInvalidRectangle
	}

	// Extract all rectangle coordinates to find bounds
	left := rect[0].X
	right := rect[0].X
	top := rect[0].Y
	bottom := rect[0].Y

	// Find actual bounds from all 4 points (handles any orientation)
	for _, pt := range rect {
		if pt.X < left {
			left = pt.X
		}
		if pt.X > right {
			right = pt.X
		}
		if pt.Y < top {
			top = pt.Y
		}
		if pt.Y > bottom {
			bottom = pt.Y
		}
	}

	// Check for degenerate rectangle (zero area)
	if left >= right || top >= bottom {
		return Paths64{}, nil // Empty result for degenerate rectangle
	}

	clipper := &rectClipper{
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
	}

	result := make(Paths64, 0, len(paths))
	for _, path := range paths {
		if len(path) < 2 {
			continue // Skip degenerate paths
		}

		clipped := clipper.clipPath(path)
		cleaned := cleanPath(clipped)
		if len(cleaned) >= 2 { // Need at least 2 points for a valid path
			result = append(result, cleaned)
		}
	}

	return result, nil
}

// rectClipper implements rectangle clipping for a single rectangular bounds
type rectClipper struct {
	left, top, right, bottom int64
}

// location represents which side of the rectangle a point is on
type location uint8

const (
	locInside location = iota
	locLeft
	locTop
	locRight
	locBottom
)

// getLocation determines which region a point is in relative to the rectangle
func (rc *rectClipper) getLocation(pt Point64) location {
	if pt.X < rc.left {
		return locLeft
	}
	if pt.X > rc.right {
		return locRight
	}
	if pt.Y < rc.top {
		return locTop
	}
	if pt.Y > rc.bottom {
		return locBottom
	}
	return locInside
}

// clipPath clips a single path against the rectangle
func (rc *rectClipper) clipPath(path Path64) Path64 {
	if len(path) == 0 {
		return nil
	}

	// Check if entire path is inside rectangle for quick accept
	allInside := true
	for _, pt := range path {
		if rc.getLocation(pt) != locInside {
			allInside = false
			break
		}
	}
	if allInside {
		return path
	}

	// Apply Sutherland-Hodgman clipping against each edge
	clipped := path

	// Clip against left edge
	clipped = rc.clipAgainstEdge(clipped, locLeft)
	if len(clipped) == 0 {
		return nil
	}

	// Clip against top edge
	clipped = rc.clipAgainstEdge(clipped, locTop)
	if len(clipped) == 0 {
		return nil
	}

	// Clip against right edge
	clipped = rc.clipAgainstEdge(clipped, locRight)
	if len(clipped) == 0 {
		return nil
	}

	// Clip against bottom edge
	clipped = rc.clipAgainstEdge(clipped, locBottom)

	return clipped
}

// clipAgainstEdge clips a path against a single edge of the rectangle
func (rc *rectClipper) clipAgainstEdge(path Path64, edge location) Path64 {
	if len(path) == 0 {
		return nil
	}

	var result Path64

	for i := 0; i < len(path); i++ {
		curr := path[i]
		prev := path[(i+len(path)-1)%len(path)]

		currInside := rc.isInsideEdge(curr, edge)
		prevInside := rc.isInsideEdge(prev, edge)

		if currInside {
			if !prevInside {
				// Entering: add intersection then current point
				if intersection, ok := rc.getIntersection(prev, curr, edge); ok {
					result = append(result, intersection)
				}
			}
			result = append(result, curr)
		} else if prevInside {
			// Exiting: add intersection only
			if intersection, ok := rc.getIntersection(prev, curr, edge); ok {
				result = append(result, intersection)
			}
		}
		// Both outside: add nothing
	}

	return result
}

// isInsideEdge checks if a point is on the inside of a specific edge
func (rc *rectClipper) isInsideEdge(pt Point64, edge location) bool {
	switch edge {
	case locLeft:
		return pt.X >= rc.left
	case locTop:
		return pt.Y >= rc.top
	case locRight:
		return pt.X <= rc.right
	case locBottom:
		return pt.Y <= rc.bottom
	default:
		return true
	}
}

// getIntersection calculates intersection of line segment with rectangle edge
func (rc *rectClipper) getIntersection(p1, p2 Point64, edge location) (Point64, bool) {
	switch edge {
	case locLeft:
		return rc.intersectWithVerticalLine(p1, p2, rc.left)
	case locRight:
		return rc.intersectWithVerticalLine(p1, p2, rc.right)
	case locTop:
		return rc.intersectWithHorizontalLine(p1, p2, rc.top)
	case locBottom:
		return rc.intersectWithHorizontalLine(p1, p2, rc.bottom)
	default:
		return Point64{}, false
	}
}

// intersectWithVerticalLine finds intersection with vertical line x = lineX
func (rc *rectClipper) intersectWithVerticalLine(p1, p2 Point64, lineX int64) (Point64, bool) {
	if p1.X == p2.X {
		return Point64{}, false // Line is vertical, no intersection
	}

	// Use integer arithmetic when possible for better precision
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	// Calculate y = p1.Y + dy * (lineX - p1.X) / dx
	// Use 128-bit intermediate to avoid overflow
	numerator := NewInt128(dy).Mul64(lineX - p1.X)
	denominator := dx

	// Convert to float64 for final division (still more precise than original)
	t := numerator.ToFloat64() / float64(denominator)
	y := float64(p1.Y) + t

	return Point64{X: lineX, Y: int64(y + 0.5)}, true // Round to nearest integer
}

// intersectWithHorizontalLine finds intersection with horizontal line y = lineY
func (rc *rectClipper) intersectWithHorizontalLine(p1, p2 Point64, lineY int64) (Point64, bool) {
	if p1.Y == p2.Y {
		return Point64{}, false // Line is horizontal, no intersection
	}

	// Use integer arithmetic when possible for better precision
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	// Calculate x = p1.X + dx * (lineY - p1.Y) / dy
	// Use 128-bit intermediate to avoid overflow
	numerator := NewInt128(dx).Mul64(lineY - p1.Y)
	denominator := dy

	// Convert to float64 for final division (still more precise than original)
	t := numerator.ToFloat64() / float64(denominator)
	x := float64(p1.X) + t

	return Point64{X: int64(x + 0.5), Y: lineY}, true // Round to nearest integer
}

// cleanPath removes duplicate consecutive points and returns a clean path
func cleanPath(path Path64) Path64 {
	if len(path) <= 1 {
		return path
	}

	var cleaned Path64
	prev := path[0]
	cleaned = append(cleaned, prev)

	for i := 1; i < len(path); i++ {
		curr := path[i]
		if curr.X != prev.X || curr.Y != prev.Y { // Only add if different from previous
			cleaned = append(cleaned, curr)
			prev = curr
		}
	}

	// Remove final duplicate if path is closed and first/last points are same
	if len(cleaned) > 2 && cleaned[0].X == cleaned[len(cleaned)-1].X && cleaned[0].Y == cleaned[len(cleaned)-1].Y {
		cleaned = cleaned[:len(cleaned)-1]
	}

	return cleaned
}
