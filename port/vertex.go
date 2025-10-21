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
	Pt    Point64      // The vertex coordinates
	Next  *Vertex      // Next vertex in the polygon chain
	Prev  *Vertex      // Previous vertex in the polygon chain
	Flags VertexFlags  // Vertex flags (local min/max, open start/end, etc.)
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
func markLocalMinimaAndMaxima(vertices []*Vertex, isOpen bool) {
	if len(vertices) < 3 {
		return // Cannot have local min/max with less than 3 vertices
	}

	for i, v := range vertices {
		var prevV, nextV *Vertex

		if isOpen {
			// For open paths, don't check first and last vertices as min/max
			if i == 0 || i == len(vertices)-1 {
				continue
			}
			prevV = vertices[i-1]
			nextV = vertices[i+1]
		} else {
			// For closed paths, wrap around
			prevV = vertices[(i-1+len(vertices))%len(vertices)]
			nextV = vertices[(i+1)%len(vertices)]
		}

		// Check for local minimum - Y is smaller than OR equal to both neighbors
		// but at least one neighbor has a larger Y
		if v.Pt.Y <= prevV.Pt.Y && v.Pt.Y <= nextV.Pt.Y && (prevV.Pt.Y > v.Pt.Y || nextV.Pt.Y > v.Pt.Y) {
			v.Flags |= VertexFlagsLocalMin
		}
		
		// Check for local maximum - Y is larger than OR equal to both neighbors  
		// but at least one neighbor has a smaller Y
		if v.Pt.Y >= prevV.Pt.Y && v.Pt.Y >= nextV.Pt.Y && (prevV.Pt.Y < v.Pt.Y || nextV.Pt.Y < v.Pt.Y) {
			v.Flags |= VertexFlagsLocalMax
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