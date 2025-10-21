package clipper

// simpleRectangleClipping implements basic boolean operations for axis-aligned rectangles
// This is a stepping stone implementation for M3 before full Vatti algorithm
func executeClippingOperationSimple(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution, solutionOpen Paths64, err error) {
	// Handle empty inputs
	if len(subjects) == 0 {
		if clipType == Union || clipType == Xor {
			return clips, Paths64{}, nil
		}
		return Paths64{}, Paths64{}, nil
	}
	if len(clips) == 0 {
		if clipType == Union || clipType == Difference || clipType == Xor {
			return subjects, Paths64{}, nil
		}
		return Paths64{}, Paths64{}, nil
	}

	// For now, implement basic rectangle operations
	switch clipType {
	case Union:
		return simpleUnionRects(subjects, clips), Paths64{}, nil
	case Intersection:
		return simpleIntersectionRects(subjects, clips), Paths64{}, nil
	case Difference:
		return simpleDifferenceRects(subjects, clips), Paths64{}, nil
	case Xor:
		return simpleXorRects(subjects, clips), Paths64{}, nil
	default:
		return nil, nil, ErrInvalidInput
	}
}

// Rectangle represents a bounding rectangle
type Rectangle struct {
	Left, Top, Right, Bottom int64
}

// pathToRectangle converts a path to a rectangle (assuming axis-aligned)
func pathToRectangle(path Path64) Rectangle {
	if len(path) == 0 {
		return Rectangle{}
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
	
	return Rectangle{Left: minX, Top: minY, Right: maxX, Bottom: maxY}
}

// rectangleToPath converts a rectangle back to a path
func rectangleToPath(r Rectangle) Path64 {
	if r.Left >= r.Right || r.Top >= r.Bottom {
		return Path64{} // Degenerate rectangle
	}
	
	return Path64{
		{r.Left, r.Top},
		{r.Right, r.Top},
		{r.Right, r.Bottom},
		{r.Left, r.Bottom},
	}
}

// intersectRectangles returns the intersection of two rectangles
func intersectRectangles(r1, r2 Rectangle) Rectangle {
	left := max64(r1.Left, r2.Left)
	right := min64(r1.Right, r2.Right)
	top := max64(r1.Top, r2.Top)
	bottom := min64(r1.Bottom, r2.Bottom)
	
	if left >= right || top >= bottom {
		return Rectangle{} // No intersection
	}
	
	return Rectangle{Left: left, Top: top, Right: right, Bottom: bottom}
}

// unionRectangles returns a list of rectangles representing the union
// For simplicity, this may return multiple rectangles instead of optimal merging
func unionRectangles(r1, r2 Rectangle) []Rectangle {
	// Check if rectangles overlap or are adjacent
	if r1.Right < r2.Left || r2.Right < r1.Left || 
	   r1.Bottom < r2.Top || r2.Bottom < r1.Top {
		// No overlap - return both rectangles
		result := []Rectangle{}
		if r1.Left < r1.Right && r1.Top < r1.Bottom {
			result = append(result, r1)
		}
		if r2.Left < r2.Right && r2.Top < r2.Bottom {
			result = append(result, r2)
		}
		return result
	}
	
	// For overlapping rectangles, return the bounding rectangle
	// This is simplified - a full implementation would handle more complex cases
	left := min64(r1.Left, r2.Left)
	right := max64(r1.Right, r2.Right)
	top := min64(r1.Top, r2.Top)
	bottom := max64(r1.Bottom, r2.Bottom)
	
	return []Rectangle{{Left: left, Top: top, Right: right, Bottom: bottom}}
}

// differenceRectangles computes r1 - r2
func differenceRectangles(r1, r2 Rectangle) []Rectangle {
	// Find intersection
	intersection := intersectRectangles(r1, r2)
	
	// If no intersection, return original rectangle
	if intersection.Left >= intersection.Right || intersection.Top >= intersection.Bottom {
		if r1.Left < r1.Right && r1.Top < r1.Bottom {
			return []Rectangle{r1}
		}
		return []Rectangle{}
	}
	
	// If intersection covers entire r1, return empty
	if intersection.Left <= r1.Left && intersection.Right >= r1.Right &&
	   intersection.Top <= r1.Top && intersection.Bottom >= r1.Bottom {
		return []Rectangle{}
	}
	
	// Split r1 into rectangles that don't intersect with r2
	var result []Rectangle
	
	// Left part
	if r1.Left < intersection.Left {
		result = append(result, Rectangle{
			Left: r1.Left, Top: r1.Top, 
			Right: intersection.Left, Bottom: r1.Bottom,
		})
	}
	
	// Right part  
	if intersection.Right < r1.Right {
		result = append(result, Rectangle{
			Left: intersection.Right, Top: r1.Top,
			Right: r1.Right, Bottom: r1.Bottom,
		})
	}
	
	// Top part (middle section only)
	if r1.Top < intersection.Top {
		result = append(result, Rectangle{
			Left: intersection.Left, Top: r1.Top,
			Right: intersection.Right, Bottom: intersection.Top,
		})
	}
	
	// Bottom part (middle section only)
	if intersection.Bottom < r1.Bottom {
		result = append(result, Rectangle{
			Left: intersection.Left, Top: intersection.Bottom,
			Right: intersection.Right, Bottom: r1.Bottom,
		})
	}
	
	return result
}

// xorRectangles computes the symmetric difference of two rectangles
func xorRectangles(r1, r2 Rectangle) []Rectangle {
	diff1 := differenceRectangles(r1, r2) // r1 - r2
	diff2 := differenceRectangles(r2, r1) // r2 - r1
	
	result := make([]Rectangle, 0, len(diff1)+len(diff2))
	result = append(result, diff1...)
	result = append(result, diff2...)
	
	return result
}

// simpleUnionRects implements union for rectangle-like paths
func simpleUnionRects(subjects, clips Paths64) Paths64 {
	// Convert all paths to rectangles
	var allRects []Rectangle
	
	for _, path := range subjects {
		if len(path) >= 3 {
			allRects = append(allRects, pathToRectangle(path))
		}
	}
	
	for _, path := range clips {
		if len(path) >= 3 {
			allRects = append(allRects, pathToRectangle(path))
		}
	}
	
	if len(allRects) == 0 {
		return Paths64{}
	}
	
	// Simple approach: merge overlapping rectangles
	// For now, just return all as union (not optimal but working)
	result := allRects
	
	// Convert back to paths
	var solution Paths64
	for _, rect := range result {
		path := rectangleToPath(rect)
		if len(path) >= 3 {
			solution = append(solution, path)
		}
	}
	
	return solution
}

// simpleIntersectionRects implements intersection for rectangle-like paths
func simpleIntersectionRects(subjects, clips Paths64) Paths64 {
	var solution Paths64
	
	// Find intersections between all subject and clip rectangles
	for _, subjectPath := range subjects {
		if len(subjectPath) < 3 {
			continue
		}
		subjectRect := pathToRectangle(subjectPath)
		
		for _, clipPath := range clips {
			if len(clipPath) < 3 {
				continue
			}
			clipRect := pathToRectangle(clipPath)
			
			intersection := intersectRectangles(subjectRect, clipRect)
			path := rectangleToPath(intersection)
			if len(path) >= 3 {
				solution = append(solution, path)
			}
		}
	}
	
	return solution
}

// simpleDifferenceRects implements difference for rectangle-like paths  
func simpleDifferenceRects(subjects, clips Paths64) Paths64 {
	var solution Paths64
	
	// For each subject, subtract all clips
	for _, subjectPath := range subjects {
		if len(subjectPath) < 3 {
			continue
		}
		
		currentRects := []Rectangle{pathToRectangle(subjectPath)}
		
		// Subtract each clip from all current rectangles
		for _, clipPath := range clips {
			if len(clipPath) < 3 {
				continue
			}
			clipRect := pathToRectangle(clipPath)
			
			var newRects []Rectangle
			for _, rect := range currentRects {
				diff := differenceRectangles(rect, clipRect)
				newRects = append(newRects, diff...)
			}
			currentRects = newRects
		}
		
		// Convert result rectangles to paths
		for _, rect := range currentRects {
			path := rectangleToPath(rect)
			if len(path) >= 3 {
				solution = append(solution, path)
			}
		}
	}
	
	return solution
}

// simpleXorRects implements XOR for rectangle-like paths
func simpleXorRects(subjects, clips Paths64) Paths64 {
	// XOR = (A - B) ∪ (B - A)
	diffAB := simpleDifferenceRects(subjects, clips)
	diffBA := simpleDifferenceRects(clips, subjects)
	
	result := make(Paths64, 0, len(diffAB)+len(diffBA))
	result = append(result, diffAB...)
	result = append(result, diffBA...)
	
	return result
}

// Helper functions max64 and min64 are defined in geometry.go