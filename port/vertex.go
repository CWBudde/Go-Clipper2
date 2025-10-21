package clipper

// ==============================================================================
// Vertex Structure and Management for Vatti Algorithm
// ==============================================================================

// VertexFlags represents various flags that can be set on a vertex
type VertexFlags uint8

const (
	VertexFlagsEmpty     VertexFlags = 0
	VertexFlagsOpenStart VertexFlags = 1 << iota // Start of an open path
	VertexFlagsOpenEnd                           // End of an open path
	VertexFlagsLocalMax                          // Local maximum vertex
	VertexFlagsLocalMin                          // Local minimum vertex
)

// Vertex represents a polygon vertex in the Vatti algorithm
// This is the fundamental building block of the scanline algorithm
type Vertex struct {
	Pt    Point64     // The vertex coordinates
	Next  *Vertex     // Next vertex in the polygon chain
	Prev  *Vertex     // Previous vertex in the polygon chain
	Flags VertexFlags // Vertex flags (local min/max, open start/end, etc.)
}

// JoinWith specifies how an edge joins with other edges
type JoinWith uint8

const (
	JoinWithNoJoin JoinWith = iota
	JoinWithLeft
	JoinWithRight
)

// createVertexFromPath converts a Path64 to a linked chain of vertices
func createVertexFromPath(path Path64, isOpen bool) *Vertex {
	if len(path) < 2 {
		return nil // Degenerate path
	}

	// Create vertices
	vertices := make([]*Vertex, len(path))
	for i, pt := range path {
		vertices[i] = &Vertex{
			Pt:    pt,
			Flags: VertexFlagsEmpty,
		}
	}

	// Link vertices into a chain
	for i := 0; i < len(vertices); i++ {
		if isOpen {
			// Open path: no wrap-around linking
			if i > 0 {
				vertices[i].Prev = vertices[i-1]
			}
			if i < len(vertices)-1 {
				vertices[i].Next = vertices[i+1]
			}
		} else {
			// Closed path: wrap-around linking
			vertices[i].Prev = vertices[(i-1+len(vertices))%len(vertices)]
			vertices[i].Next = vertices[(i+1)%len(vertices)]
		}
	}

	// Set open path flags
	if isOpen && len(vertices) >= 2 {
		vertices[0].Flags |= VertexFlagsOpenStart
		vertices[len(vertices)-1].Flags |= VertexFlagsOpenEnd
	}

	// Identify and mark local minima and maxima
	markLocalMinimaAndMaxima(vertices, isOpen)

	return vertices[0] // Return first vertex as chain head
}

// markLocalMinimaAndMaxima identifies and marks local minima and maxima in the vertex chain
// This mirrors the C++ AddPaths_ function (clipper.engine.cpp:616-707)
func markLocalMinimaAndMaxima(vertices []*Vertex, isOpen bool) {
	if len(vertices) < 2 {
		return
	}

	n := len(vertices)
	v0 := vertices[0]

	debugLog("markLocalMinimaAndMaxima: %d vertices, isOpen=%v, v0=%v", n, isOpen, v0.Pt)
	for i, v := range vertices {
		debugLog("  Vertex[%d]: %v", i, v.Pt)
	}

	// Determine initial direction (going_up)
	// For closed paths, skip over horizontal edges to find true direction
	var goingUp bool
	if isOpen {
		// For open paths, mark start vertex
		curr := vertices[1]
		idx := 1
		// Skip horizontal edges at start (C++ line 657-658)
		for idx < n && curr.Pt.Y == v0.Pt.Y {
			curr = vertices[idx]
			idx++
		}
		if idx >= n {
			return // Completely flat open path
		}
		goingUp = curr.Pt.Y <= v0.Pt.Y
		if goingUp {
			v0.Flags |= VertexFlagsOpenStart | VertexFlagsLocalMin
		} else {
			v0.Flags |= VertexFlagsOpenStart | VertexFlagsLocalMax
		}
	} else {
		// For closed paths (C++ line 670-677)
		prevV := vertices[n-1]
		prevIdx := n - 1
		// Skip over horizontal edges going backwards
		for prevIdx > 0 && prevV.Pt.Y == v0.Pt.Y {
			prevIdx--
			prevV = vertices[prevIdx]
		}
		if prevV.Pt.Y == v0.Pt.Y {
			return // Completely flat closed path - skip
		}
		goingUp = prevV.Pt.Y > v0.Pt.Y
	}

	goingUp0 := goingUp
	debugLog("  Initial goingUp=%v, goingUp0=%v", goingUp, goingUp0)

	// Traverse vertices and mark local minima/maxima (C++ line 681-693)
	for i := 1; i < n; i++ {
		curr := vertices[i]
		prev := vertices[i-1]

		if curr.Pt.Y > prev.Pt.Y && goingUp {
			// Direction change from up to down - mark local maximum
			debugLog("  Vertex[%d] %v: LocalMax (was going up, now going down)", i-1, prev.Pt)
			prev.Flags |= VertexFlagsLocalMax
			goingUp = false
		} else if curr.Pt.Y < prev.Pt.Y && !goingUp {
			// Direction change from down to up - mark local minimum
			debugLog("  Vertex[%d] %v: LocalMin (was going down, now going up)", i-1, prev.Pt)
			prev.Flags |= VertexFlagsLocalMin
			goingUp = true
		}
	}

	// Handle last vertex for open/closed paths (C++ line 695-707)
	lastV := vertices[n-1]
	if isOpen {
		lastV.Flags |= VertexFlagsOpenEnd
		if goingUp {
			lastV.Flags |= VertexFlagsLocalMax
		} else {
			lastV.Flags |= VertexFlagsLocalMin
		}
	} else if goingUp != goingUp0 {
		// Closed path - check if last vertex is min or max
		if goingUp0 {
			lastV.Flags |= VertexFlagsLocalMin
		} else {
			lastV.Flags |= VertexFlagsLocalMax
		}
	}
}

// isLocalMinimum checks if a vertex is a local minimum
func (v *Vertex) isLocalMinimum() bool {
	return (v.Flags & VertexFlagsLocalMin) != 0
}

// isLocalMaximum checks if a vertex is a local maximum
func (v *Vertex) isLocalMaximum() bool {
	return (v.Flags & VertexFlagsLocalMax) != 0
}

// isOpenStart checks if a vertex is the start of an open path
func (v *Vertex) isOpenStart() bool {
	return (v.Flags & VertexFlagsOpenStart) != 0
}

// isOpenEnd checks if a vertex is the end of an open path
func (v *Vertex) isOpenEnd() bool {
	return (v.Flags & VertexFlagsOpenEnd) != 0
}

// findLocalMinima finds all local minima in a vertex chain and creates LocalMinima structures
func findLocalMinima(startVertex *Vertex, pathType PathType, isOpen bool) []*LocalMinima {
	if startVertex == nil {
		return nil
	}

	var localMinima []*LocalMinima
	current := startVertex

	// Traverse the vertex chain
	for {
		if current.isLocalMinimum() {
			// Create local minimum entry
			lm := &LocalMinima{
				Vertex:   current,
				PathType: pathType,
				IsOpen:   isOpen,
			}
			localMinima = append(localMinima, lm)
		}

		current = current.Next
		if current == nil || current == startVertex {
			break // End of chain or completed loop
		}
	}

	return localMinima
}

// getVertexChainLength returns the number of vertices in the chain
func getVertexChainLength(startVertex *Vertex) int {
	if startVertex == nil {
		return 0
	}

	count := 1
	current := startVertex.Next

	for current != nil && current != startVertex {
		count++
		current = current.Next
	}

	return count
}

// validateVertexChain performs basic validation on a vertex chain
func validateVertexChain(startVertex *Vertex) bool {
	if startVertex == nil {
		return false
	}

	// Check for basic integrity
	current := startVertex
	visited := make(map[*Vertex]bool)

	for {
		if visited[current] {
			// We've seen this vertex before - check if it's the start (valid loop) or internal (invalid)
			return current == startVertex
		}
		visited[current] = true

		// Check bidirectional linking
		if current.Next != nil && current.Next.Prev != current {
			return false // Forward/backward link mismatch
		}
		if current.Prev != nil && current.Prev.Next != current {
			return false // Backward/forward link mismatch
		}

		current = current.Next
		if current == nil || current == startVertex {
			break
		}
	}

	return true
}
