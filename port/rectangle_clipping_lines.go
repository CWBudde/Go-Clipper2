//go:build !clipper_cgo

package clipper

// This file implements line/open path clipping against rectangles
// Reference: clipper.rectclip.cpp RectClipLines64

// rectClipLinesImpl clips open paths (lines) against a rectangle
func rectClipLinesImpl(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) != 4 {
		return nil, ErrInvalidRectangle
	}

	// Extract rectangle bounds
	left := rect[0].X
	right := rect[0].X
	top := rect[0].Y
	bottom := rect[0].Y

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

	// Check for degenerate rectangle
	if left >= right || top >= bottom {
		return Paths64{}, nil
	}

	clipper := &rectLineClipper{
		left:   left,
		top:    top,
		right:  right,
		bottom: bottom,
	}

	var result Paths64
	for _, path := range paths {
		if len(path) < 2 {
			continue // Skip degenerate paths
		}

		clipped := clipper.clipPolyline(path)
		result = append(result, clipped...)
	}

	return result, nil
}

// rectLineClipper handles line clipping against a rectangle
type rectLineClipper struct {
	left, top, right, bottom int64
}

// outcode represents which region a point is in (Cohen-Sutherland style)
type outcode int

const (
	outcodeInside outcode = 0
	outcodeLeft   outcode = 1
	outcodeRight  outcode = 2
	outcodeBottom outcode = 4
	outcodeTop    outcode = 8
)

// computeOutcode returns the outcode for a point (Cohen-Sutherland)
func (rc *rectLineClipper) computeOutcode(pt Point64) outcode {
	code := outcodeInside

	if pt.X < rc.left {
		code |= outcodeLeft
	} else if pt.X > rc.right {
		code |= outcodeRight
	}

	if pt.Y < rc.top {
		code |= outcodeTop
	} else if pt.Y > rc.bottom {
		code |= outcodeBottom
	}

	return code
}

// clipPolyline clips an open path (polyline) and returns multiple clipped segments
func (rc *rectLineClipper) clipPolyline(path Path64) Paths64 {
	if len(path) < 2 {
		return nil
	}

	var result Paths64
	var currentSegment Path64

	// Process each line segment
	for i := 0; i < len(path)-1; i++ {
		p1 := path[i]
		p2 := path[i+1]

		p1out, p2out, accept := rc.clipSegment(p1, p2)

		if accept {
			// Segment (or part of it) is visible
			if len(currentSegment) == 0 {
				// Start new segment
				currentSegment = append(currentSegment, p1out)
				currentSegment = append(currentSegment, p2out)
			} else {
				// Check if we can continue the current segment
				lastPt := currentSegment[len(currentSegment)-1]
				if lastPt.X == p1out.X && lastPt.Y == p1out.Y {
					// Continue existing segment
					currentSegment = append(currentSegment, p2out)
				} else {
					// Gap detected - save current segment and start new one
					if len(currentSegment) >= 2 {
						result = append(result, currentSegment)
					}
					currentSegment = Path64{p1out, p2out}
				}
			}
		} else {
			// Segment completely outside - save current segment if any
			if len(currentSegment) >= 2 {
				result = append(result, currentSegment)
			}
			currentSegment = nil
		}
	}

	// Save final segment if any
	if len(currentSegment) >= 2 {
		result = append(result, currentSegment)
	}

	return result
}

// clipSegment clips a line segment using Cohen-Sutherland algorithm
// Returns: clipped p1, clipped p2, accept flag
func (rc *rectLineClipper) clipSegment(p1, p2 Point64) (Point64, Point64, bool) {
	code1 := rc.computeOutcode(p1)
	code2 := rc.computeOutcode(p2)

	for {
		if (code1 | code2) == outcodeInside {
			// Both points inside
			return p1, p2, true
		}

		if (code1 & code2) != outcodeInside {
			// Both points on same side of rectangle (outside)
			return p1, p2, false
		}

		// Line crosses rectangle boundary - clip it
		var x, y int64
		var codeOut outcode

		// Pick a point outside the rectangle
		if code1 != outcodeInside {
			codeOut = code1
		} else {
			codeOut = code2
		}

		// Find intersection point
		if (codeOut & outcodeTop) != 0 {
			// Point is above the clip rectangle
			x = p1.X + (p2.X-p1.X)*(rc.top-p1.Y)/(p2.Y-p1.Y)
			y = rc.top
		} else if (codeOut & outcodeBottom) != 0 {
			// Point is below the clip rectangle
			x = p1.X + (p2.X-p1.X)*(rc.bottom-p1.Y)/(p2.Y-p1.Y)
			y = rc.bottom
		} else if (codeOut & outcodeRight) != 0 {
			// Point is to the right of clip rectangle
			y = p1.Y + (p2.Y-p1.Y)*(rc.right-p1.X)/(p2.X-p1.X)
			x = rc.right
		} else if (codeOut & outcodeLeft) != 0 {
			// Point is to the left of clip rectangle
			y = p1.Y + (p2.Y-p1.Y)*(rc.left-p1.X)/(p2.X-p1.X)
			x = rc.left
		}

		// Replace outside point with intersection point
		if codeOut == code1 {
			p1 = Point64{X: x, Y: y}
			code1 = rc.computeOutcode(p1)
		} else {
			p2 = Point64{X: x, Y: y}
			code2 = rc.computeOutcode(p2)
		}
	}
}
