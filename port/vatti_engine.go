package clipper

import (
	"math"
	"sort"
)

// ==============================================================================
// Vatti Scanline Algorithm Implementation
// ==============================================================================

// VattiEngine implements the Vatti scanline algorithm for polygon boolean operations
type VattiEngine struct {
	clipType    ClipType
	fillRule    FillRule
	minimaList  []*LocalMinima // sorted list of local minima
	activeEdges *Edge          // head of active edge list (AEL)
	currentY    int64          // current scanline Y position
	outRecords  []*OutRec      // list of output records
	succeeded   bool           // algorithm execution status

	// Scanline processing
	scanlineSet map[int64]bool // set of Y coordinates to process
}

// NewVattiEngine creates a new Vatti algorithm engine
func NewVattiEngine(clipType ClipType, fillRule FillRule) *VattiEngine {
	return &VattiEngine{
		clipType:    clipType,
		fillRule:    fillRule,
		scanlineSet: make(map[int64]bool),
		succeeded:   true,
	}
}

// ExecuteClipping performs the complete boolean clipping operation
func (ve *VattiEngine) ExecuteClipping(subjects, subjectsOpen, clips Paths64) (solution, solutionOpen Paths64, err error) {
	// Phase 2: Path preprocessing - Convert paths to vertex chains and find local minima
	if err := ve.addPaths(subjects, PathTypeSubject, false); err != nil {
		return nil, nil, err
	}
	if err := ve.addPaths(clips, PathTypeClip, false); err != nil {
		return nil, nil, err
	}

	// Handle empty input case
	if len(ve.minimaList) == 0 {
		return Paths64{}, Paths64{}, nil
	}

	// Sort local minima by Y coordinate
	ve.sortLocalMinima()

	// Execute main scanline algorithm
	if !ve.executeScanlineAlgorithm() {
		return nil, nil, ErrClipperExecution
	}

	// Phase 6: Build output paths
	solution = ve.buildSolutionPaths()
	solutionOpen = Paths64{} // Open paths not yet supported

	return solution, solutionOpen, nil
}

// ==============================================================================
// Phase 2: Path Processing and Local Minima Detection
// ==============================================================================

// addPaths processes input paths and creates local minima
func (ve *VattiEngine) addPaths(paths Paths64, pathType PathType, isOpen bool) error {
	for _, path := range paths {
		if len(path) < 3 && !isOpen {
			continue // Skip degenerate closed paths
		}
		if len(path) < 2 && isOpen {
			continue // Skip degenerate open paths
		}

		if err := ve.addPath(path, pathType, isOpen); err != nil {
			return err
		}
	}
	return nil
}

// addPath processes a single path and identifies local minima
func (ve *VattiEngine) addPath(path Path64, pathType PathType, isOpen bool) error {
	// Convert path to vertex chain
	startVertex := createVertexFromPath(path, isOpen)
	if startVertex == nil {
		return nil // Skip invalid paths
	}

	// Validate vertex chain
	if !validateVertexChain(startVertex) {
		return ErrInvalidInput
	}

	// Find local minima in the vertex chain
	localMinima := findLocalMinima(startVertex, pathType, isOpen)

	// Add minima to the list and collect scanline Y coordinates
	for _, lm := range localMinima {
		ve.minimaList = append(ve.minimaList, lm)
		ve.scanlineSet[lm.Vertex.Pt.Y] = true

		// Also add top points of edges to scanlines
		if lm.Vertex.Next != nil {
			ve.scanlineSet[lm.Vertex.Next.Pt.Y] = true
		}
		if lm.Vertex.Prev != nil {
			ve.scanlineSet[lm.Vertex.Prev.Pt.Y] = true
		}
	}

	return nil
}

// sortLocalMinima sorts local minima by Y coordinate (bottom to top)
func (ve *VattiEngine) sortLocalMinima() {
	sort.Slice(ve.minimaList, func(i, j int) bool {
		if ve.minimaList[i].Vertex.Pt.Y != ve.minimaList[j].Vertex.Pt.Y {
			return ve.minimaList[i].Vertex.Pt.Y < ve.minimaList[j].Vertex.Pt.Y
		}
		// If Y coordinates are equal, sort by X coordinate
		return ve.minimaList[i].Vertex.Pt.X < ve.minimaList[j].Vertex.Pt.X
	})
}

// ==============================================================================
// Main Scanline Algorithm Execution
// ==============================================================================

// executeScanlineAlgorithm runs the main Vatti scanline algorithm
func (ve *VattiEngine) executeScanlineAlgorithm() bool {
	// Build sorted list of scanline Y coordinates
	scanlines := ve.getSortedScanlines()

	minimaIndex := 0 // Index into sorted minima list

	// Process each scanline from bottom to top
	for _, y := range scanlines {
		ve.currentY = y

		// Phase 3: Insert local minima into Active Edge List
		minimaIndex = ve.insertLocalMinimaIntoAEL(minimaIndex, y)

		// Phase 4: Process intersections between active edges
		if !ve.processIntersections(y) {
			return false
		}

		// Phase 3: Update active edges and remove completed edges
		ve.updateActiveEdges(y)

		if !ve.succeeded {
			break
		}
	}

	return ve.succeeded
}

// getSortedScanlines returns sorted list of Y coordinates to process
func (ve *VattiEngine) getSortedScanlines() []int64 {
	scanlines := make([]int64, 0, len(ve.scanlineSet))
	for y := range ve.scanlineSet {
		scanlines = append(scanlines, y)
	}
	sort.Slice(scanlines, func(i, j int) bool {
		return scanlines[i] < scanlines[j]
	})
	return scanlines
}

// ==============================================================================
// Phase 3: Active Edge List Management
// ==============================================================================

// insertLocalMinimaIntoAEL inserts local minima starting at current Y into AEL
func (ve *VattiEngine) insertLocalMinimaIntoAEL(startIndex int, y int64) int {
	index := startIndex

	// Process all local minima at this Y coordinate
	for index < len(ve.minimaList) && ve.minimaList[index].Vertex.Pt.Y == y {
		lm := ve.minimaList[index]

		// Create left and right bound edges from this local minimum
		leftEdge, rightEdge := ve.createEdgesFromLocalMinimum(lm)

		if leftEdge != nil {
			ve.insertEdgeIntoAEL(leftEdge)
		}
		if rightEdge != nil {
			ve.insertEdgeIntoAEL(rightEdge)
		}

		index++
	}

	return index
}

// createEdgesFromLocalMinimum creates left and right bound edges from a local minimum
func (ve *VattiEngine) createEdgesFromLocalMinimum(lm *LocalMinima) (*Edge, *Edge) {
	vertex := lm.Vertex

	// Find the edges going up from this local minimum
	var leftEdge, rightEdge *Edge

	// Check previous vertex (left bound) - exclude horizontal edges
	if vertex.Prev != nil && vertex.Prev.Pt.Y > vertex.Pt.Y {
		leftEdge = ve.createEdge(vertex, vertex.Prev, lm, true)
	}

	// Check next vertex (right bound) - exclude horizontal edges
	if vertex.Next != nil && vertex.Next.Pt.Y > vertex.Pt.Y {
		rightEdge = ve.createEdge(vertex, vertex.Next, lm, false)
	}

	return leftEdge, rightEdge
}

// createEdge creates an edge from two vertices
func (ve *VattiEngine) createEdge(botVertex, topVertex *Vertex, localMin *LocalMinima, isLeftBound bool) *Edge {
	edge := &Edge{
		Bot:         botVertex.Pt,
		Top:         topVertex.Pt,
		CurrX:       botVertex.Pt.X,
		VertexTop:   topVertex,
		LocalMin:    localMin,
		IsLeftBound: isLeftBound,
	}

	// Calculate slope (Dx)
	if topVertex.Pt.Y != botVertex.Pt.Y {
		edge.Dx = float64(topVertex.Pt.X-botVertex.Pt.X) / float64(topVertex.Pt.Y-botVertex.Pt.Y)
	} else {
		// Horizontal edge
		if topVertex.Pt.X > botVertex.Pt.X {
			edge.Dx = -math.Inf(1)
		} else {
			edge.Dx = math.Inf(1)
		}
	}

	// Set winding direction
	if isLeftBound {
		edge.WindDx = -1 // Left bounds contribute negative winding
	} else {
		edge.WindDx = 1 // Right bounds contribute positive winding
	}

	return edge
}

// insertEdgeIntoAEL inserts an edge into the Active Edge List in sorted X order
func (ve *VattiEngine) insertEdgeIntoAEL(edge *Edge) {
	if ve.activeEdges == nil || edge.CurrX < ve.activeEdges.CurrX {
		// Insert at beginning of list
		edge.NextInAEL = ve.activeEdges
		if ve.activeEdges != nil {
			ve.activeEdges.PrevInAEL = edge
		}
		ve.activeEdges = edge
		return
	}

	// Find insertion point in sorted order
	current := ve.activeEdges
	for current.NextInAEL != nil && current.NextInAEL.CurrX <= edge.CurrX {
		current = current.NextInAEL
	}

	// Insert after current
	edge.NextInAEL = current.NextInAEL
	edge.PrevInAEL = current
	if current.NextInAEL != nil {
		current.NextInAEL.PrevInAEL = edge
	}
	current.NextInAEL = edge
}

// updateActiveEdges updates current X positions and removes completed edges
func (ve *VattiEngine) updateActiveEdges(y int64) {
	edge := ve.activeEdges

	for edge != nil {
		nextEdge := edge.NextInAEL

		// Update current X position for this scanline
		ve.updateEdgeCurrentX(edge, y)

		// Check if edge has reached its top
		if edge.Top.Y == y {
			ve.removeEdgeFromAEL(edge)
		}

		edge = nextEdge
	}
}

// updateEdgeCurrentX updates an edge's current X position for the given Y
func (ve *VattiEngine) updateEdgeCurrentX(edge *Edge, y int64) {
	switch y {
	case edge.Bot.Y:
		edge.CurrX = edge.Bot.X
	case edge.Top.Y:
		edge.CurrX = edge.Top.X
	default:
		// Calculate X using slope
		deltaY := float64(y - edge.Bot.Y)
		edge.CurrX = edge.Bot.X + int64(edge.Dx*deltaY+0.5) // Round to nearest
	}
}

// removeEdgeFromAEL removes an edge from the Active Edge List
func (ve *VattiEngine) removeEdgeFromAEL(edge *Edge) {
	if edge.PrevInAEL != nil {
		edge.PrevInAEL.NextInAEL = edge.NextInAEL
	} else {
		ve.activeEdges = edge.NextInAEL
	}

	if edge.NextInAEL != nil {
		edge.NextInAEL.PrevInAEL = edge.PrevInAEL
	}

	// Clear pointers
	edge.PrevInAEL = nil
	edge.NextInAEL = nil
}

// ==============================================================================
// Phase 4: Intersection Processing (Simplified for now)
// ==============================================================================

// processIntersections handles edge intersections and updates winding counts
func (ve *VattiEngine) processIntersections(y int64) bool {
	// First, update winding counts for all edges
	ve.updateWindingCounts()

	// Process intersections and create output points based on contribution transitions
	edge := ve.activeEdges
	edgeCount := 0

	// Track contribution status before and after each edge
	prevContributing := false

	for edge != nil {
		edgeCount++
		currentContributing := ve.isContributingEdge(edge)

		// Add output point if contribution status changes or if we're in a contributing region
		if currentContributing {
			ve.addOutputPoint(edge, Point64{edge.CurrX, y})
		}

		// Also check if this edge represents a transition boundary for intersection
		if ve.clipType == Intersection && prevContributing != currentContributing {
			if !currentContributing && prevContributing {
				// Exiting intersection region - also add output point
				ve.addOutputPoint(edge, Point64{edge.CurrX, y})
			}
		}

		prevContributing = currentContributing

		// Check for intersections with next edge
		if edge.NextInAEL != nil && ve.edgesIntersect(edge, edge.NextInAEL) {
			// Swap edges in AEL
			ve.swapAdjacentEdges(edge, edge.NextInAEL)
		}

		edge = edge.NextInAEL
	}

	return true
}

// edgesIntersect checks if two edges intersect (simplified)
func (ve *VattiEngine) edgesIntersect(e1, e2 *Edge) bool {
	// Simple check: if current X positions are out of order, they likely intersected
	return e1.CurrX > e2.CurrX
}

// swapAdjacentEdges swaps two adjacent edges in the AEL
func (ve *VattiEngine) swapAdjacentEdges(e1, e2 *Edge) {
	if e1.NextInAEL != e2 {
		return // Not adjacent
	}

	// Update linked list pointers
	if e1.PrevInAEL != nil {
		e1.PrevInAEL.NextInAEL = e2
	} else {
		ve.activeEdges = e2
	}

	if e2.NextInAEL != nil {
		e2.NextInAEL.PrevInAEL = e1
	}

	e2.PrevInAEL = e1.PrevInAEL
	e1.NextInAEL = e2.NextInAEL
	e1.PrevInAEL = e2
	e2.NextInAEL = e1

	// Swap current X positions
	e1.CurrX, e2.CurrX = e2.CurrX, e1.CurrX
}

// ==============================================================================
// Phase 5: Winding Count Calculation and Fill Rules
// ==============================================================================

// updateWindingCounts calculates winding counts for all active edges
func (ve *VattiEngine) updateWindingCounts() {
	windCountSubject := 0
	windCountClip := 0

	edge := ve.activeEdges
	for edge != nil {
		// Update winding count based on this edge's path type
		if edge.LocalMin.PathType == PathTypeSubject {
			windCountSubject += edge.WindDx
		} else {
			windCountClip += edge.WindDx
		}

		// All edges get both counts (subject and clip winding at this X position)
		edge.WindCount = windCountSubject
		edge.WindCount2 = windCountClip

		edge = edge.NextInAEL
	}
}

// isContributingEdge determines if an edge contributes to the output based on fill rules and clip type
func (ve *VattiEngine) isContributingEdge(edge *Edge) bool {
	// Get winding counts
	windCnt := edge.WindCount
	windCnt2 := edge.WindCount2

	// Apply fill rule to determine if regions are filled
	pft := ve.fillRule
	var pftSubject, pftClip bool

	switch pft {
	case EvenOdd:
		pftSubject = (abs(windCnt) & 1) != 0
		pftClip = (abs(windCnt2) & 1) != 0
	case NonZero:
		pftSubject = windCnt != 0
		pftClip = windCnt2 != 0
	case Positive:
		pftSubject = windCnt > 0
		pftClip = windCnt2 > 0
	case Negative:
		pftSubject = windCnt < 0
		pftClip = windCnt2 < 0
	}

	var result bool
	// Determine if edge contributes based on clip type
	switch ve.clipType {
	case Union:
		result = pftSubject || pftClip
	case Intersection:
		result = pftSubject && pftClip
	case Difference:
		if edge.LocalMin.PathType == PathTypeSubject {
			result = pftSubject && !pftClip
		} else {
			result = pftClip && !pftSubject
		}
	case Xor:
		result = pftSubject != pftClip
	default:
		result = false
	}

	return result
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ==============================================================================
// Phase 6: Output Building
// ==============================================================================

// addOutputPoint adds a point to the output polygon for a contributing edge
func (ve *VattiEngine) addOutputPoint(edge *Edge, pt Point64) {
	// For intersection operations, use a single shared output record
	// This is a simplified approach - proper Vatti algorithm has more complex polygon linking
	var outRec *OutRec

	if ve.clipType == Intersection {
		// Use single shared output record for intersection polygon
		if len(ve.outRecords) == 0 {
			outRec = &OutRec{
				Idx:   0,
				State: OutRecStateOuter,
			}
			ve.outRecords = append(ve.outRecords, outRec)
		} else {
			outRec = ve.outRecords[0] // Use first (and only) output record
		}
		edge.OutRec = outRec
	} else {
		// For other operations, create separate records per edge (original logic)
		if edge.OutRec == nil {
			edge.OutRec = &OutRec{
				Idx:   len(ve.outRecords),
				State: OutRecStateOuter,
			}
			ve.outRecords = append(ve.outRecords, edge.OutRec)
		}
		outRec = edge.OutRec
	}

	// Create output point
	outPt := &OutPt{
		Pt:  pt,
		Idx: outRec.Idx,
	}

	// Link into polygon chain
	if outRec.Pts == nil {
		// First point in this polygon
		outRec.Pts = outPt
		outPt.Next = outPt
		outPt.Prev = outPt
	} else {
		// Insert at end of existing chain
		lastPt := outRec.Pts.Prev
		outPt.Next = outRec.Pts
		outPt.Prev = lastPt
		lastPt.Next = outPt
		outRec.Pts.Prev = outPt
	}
}

// buildSolutionPaths builds the final solution paths from output records
func (ve *VattiEngine) buildSolutionPaths() Paths64 {
	var solution Paths64

	for _, outRec := range ve.outRecords {
		if outRec.Pts == nil {
			continue
		}

		path := ve.buildPathFromOutRec(outRec)
		if len(path) >= 3 { // Valid polygon needs at least 3 points
			solution = append(solution, path)
		}
	}
	return solution
}

// buildPathFromOutRec converts an output record to a path
func (ve *VattiEngine) buildPathFromOutRec(outRec *OutRec) Path64 {
	if outRec.Pts == nil {
		return Path64{}
	}

	var path Path64
	start := outRec.Pts
	current := start

	// Traverse the circular linked list of points
	for {
		path = append(path, current.Pt)
		current = current.Next
		if current == start {
			break
		}
	}

	return path
}
