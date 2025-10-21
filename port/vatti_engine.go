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
	sel         *Edge          // head of sorted edge list (SEL) for horizontal edges
	currentY    int64          // current scanline Y position
	botY        int64          // bottom of current scanbeam (for intersection detection)
	outRecords  []*OutRec      // list of output records
	succeeded   bool           // algorithm execution status

	// Intersection detection (C++ BuildIntersectList/ProcessIntersectList)
	intersectNodes []IntersectNode // list of edge intersections to process

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
	debugLogPhase("INITIALIZATION")
	debugLog("ClipType: %v, FillRule: %v", ve.clipType, ve.fillRule)
	debugLog("Subject paths: %v", subjects)
	debugLog("Clip paths: %v", clips)

	// Phase 2: Path preprocessing - Convert paths to vertex chains and find local minima
	debugLogPhase("PATH PREPROCESSING")
	if err := ve.addPaths(subjects, PathTypeSubject, false); err != nil {
		return nil, nil, err
	}
	if err := ve.addPaths(clips, PathTypeClip, false); err != nil {
		return nil, nil, err
	}

	debugLog("Found %d local minima", len(ve.minimaList))

	// Handle empty input case
	if len(ve.minimaList) == 0 {
		return Paths64{}, Paths64{}, nil
	}

	// Sort local minima by Y coordinate
	ve.sortLocalMinima()

	debugLog("Sorted local minima:")
	for i, lm := range ve.minimaList {
		debugLog("  LM[%d]: Y=%d, Point=%v", i, lm.Vertex.Pt.Y, lm.Vertex.Pt)
	}

	// Execute main scanline algorithm
	debugLogPhase("SCANLINE ALGORITHM")
	if !ve.executeScanlineAlgorithm() {
		return nil, nil, ErrClipperExecution
	}

	// Phase 6: Build output paths
	debugLogPhase("BUILD OUTPUT")
	solution = ve.buildSolutionPaths()
	solutionOpen = Paths64{} // Open paths not yet supported

	debugLog("Solution paths: %v", solution)

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

// sortLocalMinima sorts local minima by Y coordinate (top to bottom, like C++)
// C++ uses a priority queue that returns highest Y first, so we sort descending
func (ve *VattiEngine) sortLocalMinima() {
	sort.Slice(ve.minimaList, func(i, j int) bool {
		if ve.minimaList[i].Vertex.Pt.Y != ve.minimaList[j].Vertex.Pt.Y {
			return ve.minimaList[i].Vertex.Pt.Y > ve.minimaList[j].Vertex.Pt.Y // REVERSED: high Y first
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

	debugLog("Processing %d scanlines: %v", len(scanlines), scanlines)

	minimaIndex := 0 // Index into sorted minima list

	// Process each scanline from top to bottom (high Y to low Y, like C++)
	for i, y := range scanlines {
		ve.currentY = y

		debugLog("\n--- Scanline Y=%d ---", y)

		// Phase 3: Insert local minima into Active Edge List
		minimaIndex = ve.insertLocalMinimaIntoAEL(minimaIndex, y)

		debugLog("After inserting minima:")
		debugLogAEL(ve.activeEdges)

		// Process horizontal edges (first time - after inserting minima)
		// C++ line 2142
		for {
			horz, ok := ve.PopHorz()
			if !ok {
				break
			}
			ve.DoHorizontal(horz)
		}

		debugLog("After first horizontal pass:")
		debugLogAEL(ve.activeEdges)

		// Phase 4: Process intersections
		// Set botY to the bottom of the scanbeam (previous scanline Y, or current Y if first)
		if i > 0 {
			ve.botY = scanlines[i-1]
		} else {
			ve.botY = y
		}
		if !ve.processIntersections(y) {
			return false
		}

		debugLog("After processing intersections:")
		debugLogAEL(ve.activeEdges)

		// Phase 5: Process top of scanbeam (updates CurrX, handles maxima, advances edges)
		// C++ DoTopOfScanbeam line 2708
		ve.doTopOfScanbeam(y)

		debugLog("After doTopOfScanbeam:")
		debugLogAEL(ve.activeEdges)

		// Process horizontal edges (second time - after top of scanbeam)
		// C++ line 2152
		for {
			horz, ok := ve.PopHorz()
			if !ok {
				break
			}
			ve.DoHorizontal(horz)
		}

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
	// Sort descending (high Y to low Y) to match C++ top-to-bottom processing
	sort.Slice(scanlines, func(i, j int) bool {
		return scanlines[i] > scanlines[j] // REVERSED: high Y first
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
			// Update winding counts for the left edge
			ve.setWindingCount(leftEdge)
			debugLog("Inserted LEFT edge at X=%d, WC=%d/%d, Contributing=%v",
				leftEdge.CurrX, leftEdge.WindCount, leftEdge.WindCount2, ve.isContributingEdge(leftEdge))

			// Push horizontal edges to queue (C++ line 1227)
			if IsHorizontal(leftEdge) {
				ve.PushHorz(leftEdge)
				debugLog("  Pushed LEFT horizontal edge to queue")
			}
		}
		if rightEdge != nil {
			// C++ InsertRightEdge (line 1189-1195): Insert right edge IMMEDIATELY after left edge
			if leftEdge != nil {
				// Insert right edge right after left edge
				rightEdge.PrevInAEL = leftEdge
				rightEdge.NextInAEL = leftEdge.NextInAEL
				if leftEdge.NextInAEL != nil {
					leftEdge.NextInAEL.PrevInAEL = rightEdge
				}
				leftEdge.NextInAEL = rightEdge
				debugLog("Inserted RIGHT edge immediately after LEFT edge")
			} else {
				// No left edge, use normal insertion
				ve.insertEdgeIntoAEL(rightEdge)
			}
			// Update winding counts for the right edge
			// C++ line 1271-1273: right_bound inherits left_bound's winding
			if leftEdge != nil {
				rightEdge.WindCount = leftEdge.WindCount
				rightEdge.WindCount2 = leftEdge.WindCount2
			} else {
				ve.setWindingCount(rightEdge)
			}
			debugLog("Inserted RIGHT edge at X=%d, WC=%d/%d, Contributing=%v",
				rightEdge.CurrX, rightEdge.WindCount, rightEdge.WindCount2, ve.isContributingEdge(rightEdge))

			// CRITICAL: After inserting right edge, check if it's out of order with edges to its right
			// C++ lines 1288-1293: Immediately fix AEL ordering by swapping out-of-order edges
			swapCount := 0
			for rightEdge.NextInAEL != nil && ve.isValidAELOrder(rightEdge.NextInAEL, rightEdge) {
				debugLog("  RIGHT edge out of order with next edge - swapping (swap #%d)", swapCount+1)
				// Update winding counts at intersection
				ve.intersectEdges(rightEdge, rightEdge.NextInAEL, rightEdge.Bot)
				// Swap positions in AEL
				ve.swapPositionsInAEL(rightEdge, rightEdge.NextInAEL)
				swapCount++
			}
			if swapCount > 0 {
				debugLog("  Performed %d swaps to fix AEL order for RIGHT edge", swapCount)
			}

			// Push horizontal edges to queue (C++ line 1227)
			if IsHorizontal(rightEdge) {
				ve.PushHorz(rightEdge)
				debugLog("  Pushed RIGHT horizontal edge to queue")
			}

			// If both edges exist and LEFT edge is contributing, create a local minimum polygon
			// This matches the C++ implementation at line 1283
			// NOTE: We only check if LEFT edge is contributing, not both
			debugLog("Checking if should create local min poly: leftEdge=%v, leftContributing=%v",
				leftEdge != nil, leftEdge != nil && ve.isContributingEdge(leftEdge))
			if leftEdge != nil && ve.isContributingEdge(leftEdge) {
				debugLog("LEFT edge is contributing - creating local minimum polygon")
				pt := Point64{leftEdge.CurrX, y}
				ve.addLocalMinPoly(leftEdge, rightEdge, pt)
			} else if leftEdge != nil {
				debugLog("LEFT edge is NOT contributing (WC=%d/%d) - skipping local min poly creation",
					leftEdge.WindCount, leftEdge.WindCount2)
			} else {
				debugLog("No left edge - skipping local min poly creation")
			}
		}

		index++
	}

	return index
}

// createEdgesFromLocalMinimum creates left and right bound edges from a local minimum
func (ve *VattiEngine) createEdgesFromLocalMinimum(lm *LocalMinima) (*Edge, *Edge) {
	vertex := lm.Vertex

	debugLog("Creating edges from LM at %v, Prev=%v, Next=%v",
		vertex.Pt, vertex.Prev != nil, vertex.Next != nil)
	if vertex.Prev != nil {
		debugLog("  Prev.Pt=%v (Y>curr: %v)", vertex.Prev.Pt, vertex.Prev.Pt.Y > vertex.Pt.Y)
	}
	if vertex.Next != nil {
		debugLog("  Next.Pt=%v (Y>curr: %v)", vertex.Next.Pt, vertex.Next.Pt.Y > vertex.Pt.Y)
	}

	// Create edges from local minimum
	// In top-to-bottom processing, edges descend (Y decreases)
	// But horizontal edges (same Y) are also valid and handled specially
	var leftEdge, rightEdge *Edge

	// Create left bound from prev vertex (descending edge)
	if vertex.Prev != nil {
		leftEdge = ve.createEdge(vertex, vertex.Prev, lm, true)
		debugLog("  Created LEFT edge from %v to %v", vertex.Pt, vertex.Prev.Pt)
	}

	// Create right bound from next vertex (descending edge)
	if vertex.Next != nil {
		rightEdge = ve.createEdge(vertex, vertex.Next, lm, false)
		debugLog("  Created RIGHT edge from %v to %v", vertex.Pt, vertex.Next.Pt)
	}

	debugLog("  Returning: leftEdge=%v, rightEdge=%v", leftEdge != nil, rightEdge != nil)
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

// doTopOfScanbeam processes the top of the current scanbeam
// C++ ClipperBase::DoTopOfScanbeam (line 2708)
func (ve *VattiEngine) doTopOfScanbeam(y int64) {
	edge := ve.activeEdges
	edgeCount := 0

	debugLog("[doTopOfScanbeam] Starting at Y=%d", y)
	for edge != nil {
		debugLog("[doTopOfScanbeam]   Processing edge[%d] ptr=%p: Top=(%d,%d), CurrX=%d, NextInAEL=%p",
			edgeCount, edge, edge.Top.X, edge.Top.Y, edge.CurrX, edge.NextInAEL)

		// Note: edge will never be horizontal here (horizontals are processed separately)
		if edge.Top.Y == y {
			edge.CurrX = edge.Top.X
			debugLog("[doTopOfScanbeam]     Edge is at top Y=%d, CurrX updated to %d", y, edge.CurrX)

			if edge.VertexTop.isLocalMaximum() {
				// TOP OF BOUND (MAXIMA) - call doMaxima
				debugLog("[doTopOfScanbeam]     Edge is LOCAL MAXIMUM - calling doMaxima")
				edge = ve.doMaxima(edge) // doMaxima returns next edge to process
				debugLog("[doTopOfScanbeam]     doMaxima returned edge=%p", edge)
				edgeCount++
				continue
			} else {
				// INTERMEDIATE VERTEX - edge continues to next vertex
				debugLog("[doTopOfScanbeam]     Edge is INTERMEDIATE - adding point and updating")
				if edge.OutRec != nil {
					ve.addOutPt(edge, edge.Top, "doTopOfScanbeam:intermediate")
				}
				ve.UpdateEdgeIntoAEL(edge)
				if IsHorizontal(edge) {
					debugLog("[doTopOfScanbeam]     Edge became horizontal - pushing to horz queue")
					ve.PushHorz(edge) // Horizontals are processed later
				}
			}
		} else {
			// Not at top yet - update CurrX for this Y
			edge.CurrX = ve.topX(edge, y)
			debugLog("[doTopOfScanbeam]     Edge not at top yet, CurrX updated to %d", edge.CurrX)
		}

		edge = edge.NextInAEL
		edgeCount++
	}
	debugLog("[doTopOfScanbeam] Finished processing %d edges", edgeCount)
}

// doMaxima handles edge processing at local maxima
// C++ ClipperBase::DoMaxima (line 2735)
func (ve *VattiEngine) doMaxima(e *Edge) *Edge {
	nextE := e.NextInAEL
	prevE := e.PrevInAEL

	debugLog("[doMaxima] Processing edge at Top=(%d,%d), VertexTop=%p, OutRec=%v",
		e.Top.X, e.Top.Y, e.VertexTop, e.OutRec != nil)

	// Find the paired edge at this maximum
	maxPair := ve.getMaximaPair(e)
	if maxPair == nil {
		debugLog("[doMaxima]   NO PAIR FOUND - removing single edge")
		// No pair found - just remove this edge
		ve.removeEdgeFromAEL(e)
		return nextE
	}

	debugLog("[doMaxima]   PAIR FOUND at Top=(%d,%d), VertexTop=%p, OutRec=%v",
		maxPair.Top.X, maxPair.Top.Y, maxPair.VertexTop, maxPair.OutRec != nil)

	// If both edges are hot (contributing), join their OutRecs
	if e.OutRec != nil && maxPair.OutRec != nil {
		debugLog("[doMaxima]   Both edges hot - joining OutRec #%d and #%d",
			e.OutRec.Idx, maxPair.OutRec.Idx)
		ve.addLocalMaxPoly(e, maxPair, e.Top)
	} else {
		debugLog("[doMaxima]   Not joining - e.OutRec=%v, maxPair.OutRec=%v",
			e.OutRec != nil, maxPair.OutRec != nil)
	}

	// Remove both edges from AEL
	ve.removeEdgeFromAEL(e)
	ve.removeEdgeFromAEL(maxPair)

	// Return next edge to process (may have changed after removals)
	if prevE != nil {
		return prevE.NextInAEL
	}
	return ve.activeEdges
}

// getMaximaPair finds the paired edge at a local maximum
// C++ GetMaximaPair (line 254) + backward search extension
func (ve *VattiEngine) getMaximaPair(e *Edge) *Edge {
	debugLog("[getMaximaPair] Searching for pair of edge=%p at Top=(%d,%d), VertexTop=%p",
		e, e.Top.X, e.Top.Y, e.VertexTop)
	debugLog("[getMaximaPair]   e.NextInAEL=%p, e.PrevInAEL=%p", e.NextInAEL, e.PrevInAEL)

	// Search forward first (standard C++ behavior)
	e2 := e.NextInAEL
	debugLog("[getMaximaPair]   Starting FORWARD search from e2=%p", e2)
	searchCount := 0
	for e2 != nil {
		debugLog("[getMaximaPair]   Checking forward edge[%d] ptr=%p: VertexTop=%p, match=%v",
			searchCount, e2, e2.VertexTop, e2.VertexTop == e.VertexTop)
		if e2.VertexTop == e.VertexTop {
			debugLog("[getMaximaPair]   FORWARD MATCH FOUND at position %d!", searchCount)
			return e2 // Found!
		}
		e2 = e2.NextInAEL
		searchCount++
	}
	debugLog("[getMaximaPair]   Forward search: NO MATCH - searched %d edges", searchCount)

	// If forward search fails, search backward
	// This handles cases where UpdateEdgeIntoAEL advanced an edge to the same VertexTop
	// but it's positioned earlier in the AEL due to lower CurrX
	e2 = e.PrevInAEL
	debugLog("[getMaximaPair]   Starting BACKWARD search from e2=%p", e2)
	searchCount = 0
	for e2 != nil {
		debugLog("[getMaximaPair]   Checking backward edge[%d] ptr=%p: VertexTop=%p, match=%v",
			searchCount, e2, e2.VertexTop, e2.VertexTop == e.VertexTop)
		if e2.VertexTop == e.VertexTop {
			debugLog("[getMaximaPair]   BACKWARD MATCH FOUND at position %d!", searchCount)
			return e2 // Found!
		}
		e2 = e2.PrevInAEL
		searchCount++
	}
	debugLog("[getMaximaPair]   Backward search: NO MATCH - searched %d edges", searchCount)

	return nil
}

// topX calculates the X coordinate of an edge at a given Y
func (ve *VattiEngine) topX(e *Edge, y int64) int64 {
	if y == e.Top.Y {
		return e.Top.X
	}
	if y == e.Bot.Y {
		return e.Bot.X
	}
	// For horizontal edges (Dx = ±Inf), X doesn't change with Y
	if math.IsInf(e.Dx, 0) {
		return e.Bot.X // or e.Top.X, they're the same for horizontal edges
	}
	// Calculate using slope
	deltaY := float64(y - e.Bot.Y)
	return e.Bot.X + int64(e.Dx*deltaY+0.5)
}

// updateEdgePositions updates current X positions for all active edges
func (ve *VattiEngine) updateEdgePositions(y int64) {
	edge := ve.activeEdges

	for edge != nil {
		// Update current X position for this scanline
		ve.updateEdgeCurrentX(edge, y)
		edge = edge.NextInAEL
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
// Phase 4: Intersection Processing (C++ BuildIntersectList/ProcessIntersectList)
// ==============================================================================

// adjustCurrXAndCopyToSEL updates edge positions to top_y and copies AEL to SEL
// C++ ClipperBase::AdjustCurrXAndCopyToSEL (line 2113)
func (ve *VattiEngine) adjustCurrXAndCopyToSEL(topY int64) {
	edge := ve.activeEdges
	ve.sel = edge

	for edge != nil {
		// Copy AEL pointers to SEL
		edge.PrevInSEL = edge.PrevInAEL
		edge.NextInSEL = edge.NextInAEL
		edge.Jump = edge.NextInSEL

		// Update CurrX to the position at topY
		edge.CurrX = ve.topX(edge, topY)

		edge = edge.NextInAEL
	}
}

// addNewIntersectNode creates an intersection node for two edges
// C++ ClipperBase::AddNewIntersectNode (line 2356)
func (ve *VattiEngine) addNewIntersectNode(e1, e2 *Edge, topY int64) {
	// Calculate intersection point
	ip, ok := GetSegmentIntersectPt(e1.Bot, e1.Top, e2.Bot, e2.Top)
	if !ok {
		// Parallel edges - use e1's CurrX position
		ip = Point64{e1.CurrX, topY}
	}

	// Rounding errors can occasionally place the calculated intersection
	// point either below or above the scanbeam, so check and correct
	if ip.Y > ve.botY || ip.Y < topY {
		absDx1 := absFloat(e1.Dx)
		absDx2 := absFloat(e2.Dx)

		if absDx1 > 100 && absDx2 > 100 {
			if absDx1 > absDx2 {
				ip = GetClosestPointOnSegment(ip, e1.Bot, e1.Top)
			} else {
				ip = GetClosestPointOnSegment(ip, e2.Bot, e2.Top)
			}
		} else if absDx1 > 100 {
			ip = GetClosestPointOnSegment(ip, e1.Bot, e1.Top)
		} else if absDx2 > 100 {
			ip = GetClosestPointOnSegment(ip, e2.Bot, e2.Top)
		} else {
			// Clamp Y to scanbeam bounds
			if ip.Y < topY {
				ip.Y = topY
			} else {
				ip.Y = ve.botY
			}
			// Recalculate X using the edge with smaller slope
			if absDx1 < absDx2 {
				ip.X = ve.topX(e1, ip.Y)
			} else {
				ip.X = ve.topX(e2, ip.Y)
			}
		}
	}

	// Add to intersection list
	ve.intersectNodes = append(ve.intersectNodes, IntersectNode{
		Edge1: e1,
		Edge2: e2,
		Pt:    ip,
	})
}

// absFloat returns absolute value of a float64
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// buildIntersectList finds all edge intersections in the current scanbeam using merge sort
// C++ ClipperBase::BuildIntersectList (line 2390)
func (ve *VattiEngine) buildIntersectList(topY int64) bool {
	if ve.activeEdges == nil || ve.activeEdges.NextInAEL == nil {
		return false // Need at least 2 edges
	}

	// Calculate edge positions at the top of the current scanbeam, and from this
	// we will determine the intersections required to reach these new positions.
	ve.adjustCurrXAndCopyToSEL(topY)

	// Find all edge intersections in the current scanbeam using a stable merge
	// sort that ensures only adjacent edges are intersecting. Intersect info is
	// stored in intersectNodes ready to be processed in ProcessIntersectList.
	// Re merge sorts see https://stackoverflow.com/a/46319131/359538

	left := ve.sel
	var right, lEnd, rEnd, currBase, tmp *Edge

	// Perform iterative merge sort on SEL using jump pointers
	for left != nil && left.Jump != nil {
		var prevBase *Edge = nil

		for left != nil && left.Jump != nil {
			currBase = left
			right = left.Jump
			lEnd = right
			rEnd = right.Jump
			left.Jump = rEnd

			// Merge left and right sublists
			for left != lEnd && right != rEnd {
				if right.CurrX < left.CurrX {
					// Right edge has crossed left edge(s) - they intersected!
					// Add intersection nodes for all edges from right.prev back to left
					tmp = right.PrevInSEL
					for {
						ve.addNewIntersectNode(tmp, right, topY)
						if tmp == left {
							break
						}
						tmp = tmp.PrevInSEL
					}

					// Extract right edge from SEL
					tmp = right
					right = ve.extractFromSEL(tmp)
					lEnd = right

					// Insert extracted edge before left
					ve.insert1Before2InSEL(tmp, left)

					// Update currBase if needed
					if left == currBase {
						currBase = tmp
						currBase.Jump = rEnd
						if prevBase == nil {
							ve.sel = currBase
						} else {
							prevBase.Jump = currBase
						}
					}
				} else {
					left = left.NextInSEL
				}
			}

			prevBase = currBase
			left = rEnd
		}

		left = ve.sel
	}

	return len(ve.intersectNodes) > 0
}

// extractFromSEL removes an edge from the SEL and returns the next edge
func (ve *VattiEngine) extractFromSEL(e *Edge) *Edge {
	next := e.NextInSEL

	if e.PrevInSEL != nil {
		e.PrevInSEL.NextInSEL = next
	}
	if next != nil {
		next.PrevInSEL = e.PrevInSEL
	}

	e.PrevInSEL = nil
	e.NextInSEL = nil

	return next
}

// insert1Before2InSEL inserts edge1 before edge2 in the SEL
func (ve *VattiEngine) insert1Before2InSEL(e1, e2 *Edge) {
	e1.PrevInSEL = e2.PrevInSEL
	if e1.PrevInSEL != nil {
		e1.PrevInSEL.NextInSEL = e1
	} else {
		ve.sel = e1
	}
	e1.NextInSEL = e2
	e2.PrevInSEL = e1
}

// processIntersectList processes all intersection nodes bottom-up
// C++ ClipperBase::ProcessIntersectList (line 2448)
func (ve *VattiEngine) processIntersectList() {
	// We now have a list of intersections required so that edges will be
	// correctly positioned at the top of the scanbeam. However, it's important
	// that edge intersections are processed from the bottom up, but it's also
	// crucial that intersections only occur between adjacent edges.

	// First we do a sort so intersections proceed in a bottom up order
	sort.Slice(ve.intersectNodes, func(i, j int) bool {
		// Sort by Y (descending - higher Y = lower on screen), then by X
		if ve.intersectNodes[i].Pt.Y != ve.intersectNodes[j].Pt.Y {
			return ve.intersectNodes[i].Pt.Y > ve.intersectNodes[j].Pt.Y
		}
		return ve.intersectNodes[i].Pt.X < ve.intersectNodes[j].Pt.X
	})

	// Now as we process these intersections, we must sometimes adjust the order
	// to ensure that intersecting edges are always adjacent
	for i := 0; i < len(ve.intersectNodes); i++ {
		node := &ve.intersectNodes[i]

		// If edges are not adjacent, find a later intersection that makes them adjacent
		if !ve.edgesAdjacentInAEL(node.Edge1, node.Edge2) {
			j := i + 1
			for j < len(ve.intersectNodes) && !ve.edgesAdjacentInAEL(ve.intersectNodes[j].Edge1, ve.intersectNodes[j].Edge2) {
				j++
			}
			// Swap intersections to process the one with adjacent edges first
			ve.intersectNodes[i], ve.intersectNodes[j] = ve.intersectNodes[j], ve.intersectNodes[i]
			node = &ve.intersectNodes[i]
		}

		// Process the intersection
		ve.intersectEdges(node.Edge1, node.Edge2, node.Pt)
		ve.swapPositionsInAEL(node.Edge1, node.Edge2)

		// Update CurrX to intersection point
		node.Edge1.CurrX = node.Pt.X
		node.Edge2.CurrX = node.Pt.X

		// Check for horizontal joins (C++ lines 2477-2478)
		// TODO: Implement CheckJoinLeft/CheckJoinRight when we add horizontal join support
	}
}

// edgesAdjacentInAEL checks if two edges are adjacent in the AEL
func (ve *VattiEngine) edgesAdjacentInAEL(e1, e2 *Edge) bool {
	return e1.NextInAEL == e2 || e2.NextInAEL == e1
}

// processIntersections handles edge intersections using C++ BuildIntersectList/ProcessIntersectList
// C++ ClipperBase::DoIntersections (line 2347)
func (ve *VattiEngine) processIntersections(topY int64) bool {
	// Build list of all edge intersections in the current scanbeam
	if ve.buildIntersectList(topY) {
		// Process intersections bottom-up, ensuring edges are adjacent
		ve.processIntersectList()
		// Clear intersection list for next scanbeam
		ve.intersectNodes = ve.intersectNodes[:0]
	}
	return true
}

// isValidAELOrder checks if two edges are in valid AEL order
// C++ ClipperBase::IsValidAelOrder (line 1119)
// Returns true if newcomer should be AFTER resident in the AEL
func (ve *VattiEngine) isValidAELOrder(resident, newcomer *Edge) bool {
	debugLog("[isValidAELOrder] Checking: resident.CurrX=%d, newcomer.CurrX=%d",
		resident.CurrX, newcomer.CurrX)

	// First check: if CurrX values differ, order by X position
	if newcomer.CurrX != resident.CurrX {
		result := newcomer.CurrX > resident.CurrX
		debugLog("[isValidAELOrder]   Different X - result=%v", result)
		return result
	}

	// If CurrX is the same, use cross product to determine geometric order
	// Get the turning direction: resident.top, newcomer.bot, newcomer.top
	crossSign := CrossProductSign(resident.Top, newcomer.Bot, newcomer.Top)
	if crossSign != 0 {
		return crossSign < 0
	}

	// Edges are collinear - use additional heuristics
	// For simplicity, maintain current order (this is a simplified version)
	// The full C++ implementation has more complex collinear handling
	return true
}

// swapPositionsInAEL swaps two edges in the AEL (must be adjacent)
// C++ ClipperBase::SwapPositionsInAEL
func (ve *VattiEngine) swapPositionsInAEL(e1, e2 *Edge) {
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
	e2.NextInAEL = e1
	e1.PrevInAEL = e2
}

// intersectEdges handles edge intersection with winding count updates
// C++ ClipperBase::IntersectEdges (simplified for closed paths)
func (ve *VattiEngine) intersectEdges(e1, e2 *Edge, pt Point64) {
	// UPDATE WINDING COUNTS (C++ lines 1882-1917)
	if e1.LocalMin.PathType != e2.LocalMin.PathType {
		// Different polygon types - update WindCount2
		if ve.fillRule == EvenOdd {
			e1.WindCount2 = 1 - e1.WindCount2
			e2.WindCount2 = 1 - e2.WindCount2
		} else {
			e1.WindCount2 += e2.WindDx
			e2.WindCount2 -= e1.WindDx
		}
	} else {
		// Same polygon type - update WindCount
		if ve.fillRule == EvenOdd {
			oldE1WindCnt := e1.WindCount
			e1.WindCount = e2.WindCount
			e2.WindCount = oldE1WindCnt
		} else {
			if e1.WindCount+e2.WindDx == 0 {
				e1.WindCount = -e1.WindCount
			} else {
				e1.WindCount += e2.WindDx
			}
			if e2.WindCount-e1.WindDx == 0 {
				e2.WindCount = -e2.WindCount
			} else {
				e2.WindCount -= e1.WindDx
			}
		}
	}

	debugLog("[intersectEdges] Updated winding: e1: WC=%d WC2=%d, e2: WC=%d WC2=%d",
		e1.WindCount, e1.WindCount2, e2.WindCount, e2.WindCount2)
}

// ==============================================================================
// Phase 5: Winding Count Calculation and Fill Rules
// ==============================================================================

// setWindingCount sets winding count for a single edge based on its neighbors
// This mirrors the C++ SetWindCountForClosedPathEdge function
func (ve *VattiEngine) setWindingCount(edge *Edge) {
	// Find the nearest closed path edge of the same PolyType in AEL (heading left)
	pathType := edge.LocalMin.PathType
	prev := edge.PrevInAEL

	debugLog("[setWindingCount] Edge at Bot=(%d,%d), pathType=%v, WindDx=%d",
		edge.Bot.X, edge.Bot.Y, pathType, edge.WindDx)

	// Skip edges of different path type or open paths
	for prev != nil && prev.LocalMin.PathType != pathType {
		prev = prev.PrevInAEL
	}

	if prev == nil {
		// No previous edge of same type - initialize based on winding direction
		edge.WindCount = edge.WindDx
		edge.WindCount2 = 0
	} else if ve.fillRule == EvenOdd {
		// EvenOdd: winding is just the direction
		edge.WindCount = edge.WindDx
		edge.WindCount2 = prev.WindCount2
	} else {
		// NonZero, Positive, or Negative filling
		if prev.WindCount*prev.WindDx < 0 {
			// Previous edge has opposite winding direction
			if abs(prev.WindCount) > 1 {
				if prev.WindDx*edge.WindDx < 0 {
					edge.WindCount = prev.WindCount
				} else {
					edge.WindCount = prev.WindCount + edge.WindDx
				}
			} else {
				edge.WindCount = edge.WindDx
			}
		} else {
			// Previous edge has same winding direction
			if abs(prev.WindCount) > 1 && prev.WindDx*edge.WindDx < 0 {
				edge.WindCount = prev.WindCount
			} else if prev.WindCount+edge.WindDx == 0 {
				edge.WindCount = prev.WindCount
			} else {
				edge.WindCount = prev.WindCount + edge.WindDx
			}
		}
		edge.WindCount2 = prev.WindCount2
	}

	// Update WindCount2 for the opposite path type
	// C++ lines 1071-1086: Loop through ALL edges from e2 to current edge,
	// accumulating winding from edges of the opposite polygon type
	if ve.fillRule == EvenOdd {
		// EvenOdd: toggle for each edge of opposite type
		e2 := prev
		if e2 == nil {
			e2 = ve.activeEdges
		} else {
			e2 = e2.NextInAEL
		}
		for e2 != nil && e2 != edge {
			if e2.LocalMin.PathType != pathType {
				edge.WindCount2 = 1 - edge.WindCount2 // Toggle 0↔1
			}
			e2 = e2.NextInAEL
		}
	} else {
		// NonZero/Positive/Negative: accumulate all winding contributions
		e2 := prev
		if e2 == nil {
			e2 = ve.activeEdges
		} else {
			e2 = e2.NextInAEL
		}
		for e2 != nil && e2 != edge {
			if e2.LocalMin.PathType != pathType {
				edge.WindCount2 += e2.WindDx
			}
			e2 = e2.NextInAEL
		}
	}

	debugLog("[setWindingCount]   Result: WindCount=%d, WindCount2=%d, Contributing=%v",
		edge.WindCount, edge.WindCount2, ve.isContributingEdge(edge))
}

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
		// Positive fill rule: use absolute value for simple polygons (non-self-intersecting).
		// This works for both CW (positive winding) and CCW (negative winding) polygon orientations.
		pftSubject = abs(windCnt) > 0
		pftClip = abs(windCnt2) > 0
	case Negative:
		// Negative fill rule: use absolute value for simple polygons (non-self-intersecting).
		// Note: Both Positive and Negative use the same logic (abs(windCnt) > 0) because they
		// differ only in their interpretation of "filled" regions, but in the Vatti algorithm's
		// edge contribution logic, both simply need to check if the winding count is non-zero,
		// regardless of its sign. The actual difference between Positive/Negative is handled
		// by the polygon orientation during input processing.
		pftSubject = abs(windCnt) > 0
		pftClip = abs(windCnt2) > 0
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
// Local Minima/Maxima Polygon Management (from Clipper2 C++)
// ==============================================================================

// addLocalMinPoly creates a new OutRec when edges start contributing (local minimum)
// This is called when entering a filled region - both edges get assigned to the same OutRec
func (ve *VattiEngine) addLocalMinPoly(e1, e2 *Edge, pt Point64) *OutPt {
	// Create new output record
	outRec := &OutRec{
		Idx:   len(ve.outRecords),
		State: OutRecStateOuter,
	}
	ve.outRecords = append(ve.outRecords, outRec)

	// Assign both edges to this OutRec
	e1.OutRec = outRec
	e2.OutRec = outRec

	// Set front/back edges based on which is left bound
	// Left bound edge is the "front" (adds to front of circular list)
	// Right bound edge is the "back" (adds to back of circular list)
	if e1.IsLeftBound {
		outRec.FrontEdge = e1
		outRec.BackEdge = e2
	} else {
		outRec.FrontEdge = e2
		outRec.BackEdge = e1
	}

	// Create first output point
	outPt := &OutPt{
		Pt:   pt,
		Idx:  outRec.Idx,
		Next: nil,
		Prev: nil,
	}
	outPt.Next = outPt
	outPt.Prev = outPt
	outRec.Pts = outPt

	debugLog("addLocalMinPoly: Created OutRec #%d for edges at (%d,%d)", outRec.Idx, pt.X, pt.Y)

	return outPt
}

// addLocalMaxPoly handles when edges stop contributing (local maximum)
// This joins two OutRec paths if they're different, closing the polygon
func (ve *VattiEngine) addLocalMaxPoly(e1, e2 *Edge, pt Point64) *OutPt {
	debugLog("addLocalMaxPoly: Processing edges at (%d,%d)", pt.X, pt.Y)

	// Add the point to e1's OutRec
	result := ve.addOutPt(e1, pt, "addLocalMaxPoly")

	if e1.OutRec == e2.OutRec {
		// Same OutRec - just update the pts reference to close the loop
		e1.OutRec.Pts = result
		debugLog("  Same OutRec #%d - closing polygon", e1.OutRec.Idx)
	} else {
		// Different OutRecs - need to join them
		debugLog("  Different OutRecs (#%d and #%d) - joining", e1.OutRec.Idx, e2.OutRec.Idx)
		if e1.OutRec.Idx < e2.OutRec.Idx {
			ve.joinOutrecPaths(e1, e2)
		} else {
			ve.joinOutrecPaths(e2, e1)
		}
	}

	return result
}

// joinOutrecPaths merges e2's OutRec path into e1's OutRec path
// This is the critical function that creates properly ordered polygons
func (ve *VattiEngine) joinOutrecPaths(e1, e2 *Edge) {
	debugLog("joinOutrecPaths: Joining OutRec #%d into OutRec #%d", e2.OutRec.Idx, e1.OutRec.Idx)

	// Get the start and end points of each circular list
	p1St := e1.OutRec.Pts
	p2St := e2.OutRec.Pts
	p1End := p1St.Next
	p2End := p2St.Next

	// Determine if we're joining front-to-front or back-to-back
	isFront := (e1 == e1.OutRec.FrontEdge)

	if isFront {
		// Join front to front
		// Link: p2End <- p1St -> p2End and p2St -> p1End <- p2St
		p2End.Prev = p1St
		p1St.Next = p2End
		p2St.Next = p1End
		p1End.Prev = p2St
		e1.OutRec.Pts = p2St // Update pts to new front
		e1.OutRec.FrontEdge = e2.OutRec.FrontEdge
		if e1.OutRec.FrontEdge != nil {
			e1.OutRec.FrontEdge.OutRec = e1.OutRec
		}
		debugLog("  Joined front-to-front")
	} else {
		// Join back to back
		// Link: p1End <- p2St -> p1End and p1St -> p2End <- p1St
		p1End.Prev = p2St
		p2St.Next = p1End
		p1St.Next = p2End
		p2End.Prev = p1St
		e1.OutRec.BackEdge = e2.OutRec.BackEdge
		if e1.OutRec.BackEdge != nil {
			e1.OutRec.BackEdge.OutRec = e1.OutRec
		}
		debugLog("  Joined back-to-back")
	}

	// Mark e2's OutRec as deleted (set pts to nil)
	e2.OutRec.Pts = nil

	// Clear the edge pointers
	e2.OutRec = nil
	e1.OutRec = nil
}

// addOutPt adds a point to an edge's OutRec (front or back based on FrontEdge/BackEdge)
// This is the C++-aligned version that properly tracks point ordering
// The caller parameter helps debug duplicate point issues
func (ve *VattiEngine) addOutPt(edge *Edge, pt Point64, caller string) *OutPt {
	if edge.OutRec == nil {
		debugLog("WARNING: addOutPt[%s] called on edge without OutRec at (%d,%d)", caller, pt.X, pt.Y)
		return nil
	}

	outRec := edge.OutRec
	toFront := (edge == outRec.FrontEdge)

	// If this is the first point, just create it
	if outRec.Pts == nil {
		outPt := &OutPt{
			Pt:   pt,
			Idx:  outRec.Idx,
			Next: nil,
			Prev: nil,
		}
		outPt.Next = outPt
		outPt.Prev = outPt
		outRec.Pts = outPt
		debugLog("addOutPt[%s]: Added FIRST point (%d,%d) to OutRec #%d", caller, pt.X, pt.Y, outRec.Idx)
		return outPt
	}

	opFront := outRec.Pts
	opBack := opFront.Next

	// Enhanced duplicate checking - check ALL points in the circular list
	current := opFront
	for {
		if current.Pt == pt {
			debugLog("addOutPt[%s]: DUPLICATE DETECTED! Point (%d,%d) already exists in OutRec #%d (skipping)",
				caller, pt.X, pt.Y, outRec.Idx)
			return current // Return existing point
		}
		current = current.Next
		if current == opFront {
			break // Completed the circle
		}
	}

	// Create new point
	newOp := &OutPt{
		Pt:   pt,
		Idx:  outRec.Idx,
		Next: nil,
		Prev: nil,
	}

	// Insert into circular list
	opBack.Prev = newOp
	newOp.Prev = opFront
	newOp.Next = opBack
	opFront.Next = newOp

	if toFront {
		outRec.Pts = newOp
		debugLog("addOutPt[%s]: Added point (%d,%d) to FRONT of OutRec #%d", caller, pt.X, pt.Y, outRec.Idx)
	} else {
		debugLog("addOutPt[%s]: Added point (%d,%d) to BACK of OutRec #%d", caller, pt.X, pt.Y, outRec.Idx)
	}

	return newOp
}

// ==============================================================================
// Phase 6: Output Building
// ==============================================================================

// addOutputPoint adds a point to the output polygon for a contributing edge
// DEPRECATED: This is the old transition-based approach, kept for Intersection compatibility
// New code should use addLocalMinPoly/addLocalMaxPoly/addOutPt
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

	// Link into polygon chain with proper ordering
	// Key insight: For intersection, we're building a polygon by adding points
	// as we scan upward. We need to maintain proper counter-clockwise order:
	// bottom-left → bottom-right → top-right → top-left
	//
	// The trick is that RIGHT edges should append (normal order going up)
	// but LEFT edges should prepend to the start (reverse order going down)
	if outRec.Pts == nil {
		// First point in this polygon
		outRec.Pts = outPt
		outPt.Next = outPt
		outPt.Prev = outPt
	} else {
		// For RIGHT bound edges: append at end (keeps going up the right side)
		// For LEFT bound edges: prepend at start (prepares for going down the left side)
		if !edge.IsLeftBound {
			// Right bound: append to end
			// Maintains: ... → last → outPt → Pts
			lastPt := outRec.Pts.Prev
			outPt.Next = outRec.Pts
			outPt.Prev = lastPt
			lastPt.Next = outPt
			outRec.Pts.Prev = outPt
		} else {
			// Left bound: prepend before Pts (but DON'T change Pts reference)
			// This builds the return path down the left side in reverse
			// Maintains: ... → prev → outPt → Pts → ...
			outPt.Next = outRec.Pts
			outPt.Prev = outRec.Pts.Prev
			outRec.Pts.Prev.Next = outPt
			outRec.Pts.Prev = outPt
			// DON'T update outRec.Pts - keep the first point as the start.
			// This is critical: outRec.Pts always points to the first point added to the polygon,
			// ensuring that when we later traverse the circular linked list to build the final path,
			// we start from the original starting point. Changing outRec.Pts here would break
			// the correct order and could result in duplicate points or incorrect polygon orientation.
		}
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

// ==============================================================================
// Horizontal Edge Processing (C++ DoHorizontal subsystem)
// ==============================================================================

// PushHorz adds an edge to the horizontal edge queue (SEL)
// C++ line 1316-1319
func (ve *VattiEngine) PushHorz(e *Edge) {
	if ve.sel != nil {
		e.NextInSEL = ve.sel
	} else {
		e.NextInSEL = nil
	}
	ve.sel = e
}

// PopHorz removes and returns the next horizontal edge from the queue
// C++ line 1321-1325
func (ve *VattiEngine) PopHorz() (*Edge, bool) {
	e := ve.sel
	if e == nil {
		return nil, false
	}
	ve.sel = e.NextInSEL
	return e, true
}

// IsHorizontal checks if an edge is horizontal (same Y for top and bottom)
func IsHorizontal(e *Edge) bool {
	return e.Top.Y == e.Bot.Y
}

// UpdateEdgeIntoAEL advances an edge to its next vertex
// C++ line 1742-1791
func (ve *VattiEngine) UpdateEdgeIntoAEL(e *Edge) {
	// Advance edge to next vertex
	e.Bot = e.Top
	e.VertexTop = ve.getNextVertex(e)
	e.Top = e.VertexTop.Pt
	e.CurrX = e.Bot.X

	// Recalculate slope (Dx)
	if e.Top.Y != e.Bot.Y {
		e.Dx = float64(e.Top.X-e.Bot.X) / float64(e.Top.Y-e.Bot.Y)
	} else {
		// Horizontal edge
		if e.Top.X > e.Bot.X {
			e.Dx = -math.Inf(1)
		} else {
			e.Dx = math.Inf(1)
		}
	}

	// If edge is hot (contributing), add the bottom point
	if e.OutRec != nil {
		ve.addOutPt(e, e.Bot, "UpdateEdgeIntoAEL")
	}

	// Note: We don't add horizontal edges to scanline set here
	// They will be processed by DoHorizontal
}

// getNextVertex returns the next vertex in the chain
func (ve *VattiEngine) getNextVertex(e *Edge) *Vertex {
	if e.IsLeftBound {
		// Left bound edge follows Prev chain
		return e.VertexTop.Prev
	}
	// Right bound edge follows Next chain
	return e.VertexTop.Next
}

// ResetHorzDirection sets up direction markers for horizontal edge processing
// C++ line 2511-2535
func (ve *VattiEngine) ResetHorzDirection(horz *Edge, maxPair *Edge) (horzLeft, horzRight int64) {
	if horz.Bot.X == horz.Top.X {
		// Horizontal line is actually vertical (degenerate case)
		horzLeft = horz.CurrX
		horzRight = horz.CurrX

		// Check if maxPair is to the left or right
		e := horz.NextInAEL
		for e != nil && e != maxPair {
			e = e.NextInAEL
		}
		// If maxPair not found to the right, flip direction
		if e == nil {
			horz.Dx = -horz.Dx
		}
	} else if horz.CurrX < horz.Top.X {
		// Horizontal edge goes left to right
		horzLeft = horz.CurrX
		horzRight = horz.Top.X
		horz.Dx = -math.Inf(1) // Horizontal marker (going right)
	} else {
		// Horizontal edge goes right to left
		horzLeft = horz.Top.X
		horzRight = horz.CurrX
		horz.Dx = math.Inf(1) // Horizontal marker (going left)
	}

	return horzLeft, horzRight
}

// DoHorizontal processes a horizontal edge
// C++ line 2537-2708
func (ve *VattiEngine) DoHorizontal(horz *Edge) {
	debugLog("DoHorizontal: Processing horizontal edge at Y=%d from X=%d to X=%d",
		horz.Bot.Y, horz.Bot.X, horz.Top.X)

	// Find the maximum pair edge (where this horizontal edge terminates)
	// Look for an edge that shares the same Top point (not just same vertex object)
	var maxPair *Edge

	// Search for an edge that ends at the same point as this horizontal edge's top
	if horz.IsLeftBound {
		// Left bound - look to the right for paired edge
		e := horz.NextInAEL
		for e != nil {
			// Check if this edge's top matches our horizontal's top
			if e.CurrX == horz.Top.X && e.Top.Y == horz.Top.Y && e.Top == horz.Top {
				maxPair = e
				break
			}
			// Stop searching beyond the horizontal bounds
			if e.CurrX > horz.Top.X {
				break
			}
			e = e.NextInAEL
		}
	} else {
		// Right bound - look to the left for paired edge
		e := horz.PrevInAEL
		for e != nil {
			// Check if this edge's top matches our horizontal's top
			if e.CurrX == horz.Top.X && e.Top.Y == horz.Top.Y && e.Top == horz.Top {
				maxPair = e
				break
			}
			// Stop searching beyond the horizontal bounds
			if e.CurrX < horz.Top.X {
				break
			}
			e = e.PrevInAEL
		}
	}

	// Also check if vertex is marked as local maximum
	if maxPair == nil && horz.VertexTop != nil && horz.VertexTop.isLocalMaximum() {
		debugLog("  VertexTop is marked as LocalMax but no paired edge found in AEL")
	}

	// Set up direction markers (left and right X bounds)
	horzLeft, horzRight := ve.ResetHorzDirection(horz, maxPair)

	debugLog("  Horizontal bounds: left=%d, right=%d, maxPair=%v",
		horzLeft, horzRight, maxPair != nil)

	isHot := (horz.OutRec != nil)

	// Process edges that intersect this horizontal edge
	// We need to add intermediate points for all contributing edges we cross
	e := horz.NextInAEL
	for e != nil {
		// Stop when we reach the maximum pair or go beyond the horizontal bounds
		if e == maxPair {
			break
		}
		if e.CurrX > horzRight {
			break
		}

		// If both edges are hot (contributing), add intermediate point
		if isHot && e.OutRec != nil {
			pt := Point64{e.CurrX, horz.Bot.Y}

			// Add point to horizontal edge's OutRec
			if horz.OutRec != nil {
				ve.addOutPt(horz, pt, "DoHorizontal:intermediate")
			}

			debugLog("  Added intermediate point at (%d,%d) to OutRec #%d",
				pt.X, pt.Y, horz.OutRec.Idx)
		}

		e = e.NextInAEL
	}

	// Handle the end of the horizontal edge
	if maxPair != nil {
		// This horizontal edge ends at a local maximum
		debugLog("  Horizontal edge ends at local maximum")

		if isHot {
			// Add the final point at the maximum
			pt := Point64{horz.Top.X, horz.Top.Y}
			ve.addLocalMaxPoly(horz, maxPair, pt)
		}

		// Remove the maximum pair from AEL if it has reached its top
		if maxPair.Top.Y == horz.Top.Y && !IsHorizontal(maxPair) {
			ve.removeEdgeFromAEL(maxPair)
		}
	} else {
		// Horizontal edge continues to next vertex
		debugLog("  Horizontal edge continues (no maximum pair)")

		if isHot {
			pt := Point64{horz.Top.X, horz.Top.Y}
			ve.addOutPt(horz, pt, "DoHorizontal:end")
		}
	}

	// Check if the edge continues beyond this horizontal segment
	if !IsHorizontal(horz) {
		// Edge has become non-horizontal after advancing
		// This shouldn't happen in DoHorizontal since we're processing an already-horizontal edge
		// But we keep this check for safety
		return
	}

	// If the horizontal edge needs to continue to the next vertex, update it
	if maxPair == nil && horz.Top.Y == horz.Bot.Y {
		// Try to advance to next vertex
		nextVertex := ve.getNextVertex(horz)
		if nextVertex != nil && nextVertex.Pt.Y == horz.Top.Y {
			// Next vertex is also at the same Y - another horizontal segment
			ve.UpdateEdgeIntoAEL(horz)
			if IsHorizontal(horz) {
				// Push back to horizontal queue for further processing
				ve.PushHorz(horz)
			}
		}
	}
}
