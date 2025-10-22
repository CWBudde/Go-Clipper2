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

	minimaIndex := 0 // Index into sorted minima list

	// Process each scanline from top to bottom (high Y to low Y, like C++)
	for i, y := range scanlines {
		ve.currentY = y

		// Phase 3: Insert local minima into Active Edge List
		minimaIndex = ve.insertLocalMinimaIntoAEL(minimaIndex, y)

		// Process horizontal edges (first time - after inserting minima)
		// C++ line 2142
		for {
			horz, ok := ve.PopHorz()
			if !ok {
				break
			}
			ve.DoHorizontal(horz)
		}

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

		// Phase 5: Process top of scanbeam (updates CurrX, handles maxima, advances edges)
		// C++ DoTopOfScanbeam line 2708
		ve.doTopOfScanbeam(y)

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

			// Push horizontal edges to queue (C++ line 1227)
			if IsHorizontal(leftEdge) {
				ve.PushHorz(leftEdge)
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

			// CRITICAL: After inserting right edge, check if it's out of order with edges to its right
			// C++ lines 1288-1293: Immediately fix AEL ordering by swapping out-of-order edges
			swapCount := 0
			for rightEdge.NextInAEL != nil && ve.isValidAELOrder(rightEdge.NextInAEL, rightEdge) {
				// Update winding counts at intersection
				ve.intersectEdges(rightEdge, rightEdge.NextInAEL, rightEdge.Bot)
				// Swap positions in AEL
				ve.swapPositionsInAEL(rightEdge, rightEdge.NextInAEL)
				swapCount++
			}

			// Push horizontal edges to queue (C++ line 1227)
			if IsHorizontal(rightEdge) {
				ve.PushHorz(rightEdge)
			}

			// If both edges exist and LEFT edge is contributing, create a local minimum polygon
			// This matches the C++ implementation at line 1283
			// NOTE: We only check if LEFT edge is contributing, not both
			if leftEdge != nil && ve.isContributingEdge(leftEdge) {
				pt := Point64{leftEdge.CurrX, y}
				ve.addLocalMinPoly(leftEdge, rightEdge, pt)
			}
		}

		index++
	}

	return index
}

// createEdgesFromLocalMinimum creates left and right bound edges from a local minimum
func (ve *VattiEngine) createEdgesFromLocalMinimum(lm *LocalMinima) (*Edge, *Edge) {
	vertex := lm.Vertex

	// Create edges from local minimum
	// In top-to-bottom processing, edges descend (Y decreases)
	// But horizontal edges (same Y) are also valid and handled specially
	var leftEdge, rightEdge *Edge

	// Create left bound from prev vertex (descending edge)
	if vertex.Prev != nil {
		leftEdge = ve.createEdge(vertex, vertex.Prev, lm, true)
	}

	// Create right bound from next vertex (descending edge)
	if vertex.Next != nil {
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

// doTopOfScanbeam processes the top of the current scanbeam
// C++ ClipperBase::DoTopOfScanbeam (line 2708)
func (ve *VattiEngine) doTopOfScanbeam(y int64) {
	edge := ve.activeEdges
	edgeCount := 0

	for edge != nil {
		// Note: edge will never be horizontal here (horizontals are processed separately)
		if edge.Top.Y == y {
			edge.CurrX = edge.Top.X

			if edge.VertexTop.isLocalMaximum() {
				// TOP OF BOUND (MAXIMA) - call doMaxima
				edge = ve.doMaxima(edge) // doMaxima returns next edge to process
				edgeCount++
				continue
			} else {
				// INTERMEDIATE VERTEX - edge continues to next vertex
				if edge.OutRec != nil {
					ve.addOutPt(edge, edge.Top, "doTopOfScanbeam:intermediate")
				}
				ve.UpdateEdgeIntoAEL(edge)
				if IsHorizontal(edge) {
					ve.PushHorz(edge) // Horizontals are processed later
				}
			}
		} else {
			// Not at top yet - update CurrX for this Y
			edge.CurrX = ve.topX(edge, y)
		}

		edge = edge.NextInAEL
		edgeCount++
	}
}

// doMaxima handles edge processing at local maxima
// C++ ClipperBase::DoMaxima (line 2735)
func (ve *VattiEngine) doMaxima(e *Edge) *Edge {
	nextE := e.NextInAEL
	prevE := e.PrevInAEL

	// Find the paired edge at this maximum
	maxPair := ve.getMaximaPair(e)
	if maxPair == nil {
		// No pair found - just remove this edge
		ve.removeEdgeFromAEL(e)
		return nextE
	}

	// If both edges are hot (contributing), join their OutRecs
	if e.OutRec != nil && maxPair.OutRec != nil {
		ve.addLocalMaxPoly(e, maxPair, e.Top)
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
	// Search forward first (standard C++ behavior)
	e2 := e.NextInAEL
	searchCount := 0
	for e2 != nil {
		if e2.VertexTop == e.VertexTop {
			return e2 // Found!
		}
		e2 = e2.NextInAEL
		searchCount++
	}

	// If forward search fails, search backward
	// This handles cases where UpdateEdgeIntoAEL advanced an edge to the same VertexTop
	// but it's positioned earlier in the AEL due to lower CurrX
	e2 = e.PrevInAEL
	searchCount = 0
	for e2 != nil {
		if e2.VertexTop == e.VertexTop {
			return e2 // Found!
		}
		e2 = e2.PrevInAEL
		searchCount++
	}

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
	// First check: if CurrX values differ, order by X position
	if newcomer.CurrX != resident.CurrX {
		result := newcomer.CurrX > resident.CurrX
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

	return outPt
}

// addLocalMaxPoly handles when edges stop contributing (local maximum)
// This joins two OutRec paths if they're different, closing the polygon
func (ve *VattiEngine) addLocalMaxPoly(e1, e2 *Edge, pt Point64) *OutPt {
	// Add the point to e1's OutRec
	result := ve.addOutPt(e1, pt, "addLocalMaxPoly")

	if e1.OutRec == e2.OutRec {
		// Same OutRec - just update the pts reference to close the loop
		e1.OutRec.Pts = result
	} else {
		// Different OutRecs - need to join them
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
		return outPt
	}

	opFront := outRec.Pts
	opBack := opFront.Next

	// Enhanced duplicate checking - check ALL points in the circular list
	// Add safety counter to prevent infinite loops if circular list is corrupted
	current := opFront
	safetyCount := 0
	maxPoints := 10000 // Reasonable upper limit for points in a single OutRec
	for {
		if current.Pt == pt {
			return current // Return existing point
		}
		current = current.Next
		safetyCount++

		if current == opFront {
			break // Completed the circle
		}

		if safetyCount > maxPoints {
			break
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
// PolyTree Building (Hierarchical Output)
// ==============================================================================

// BuildTree64 constructs a hierarchical PolyTree from the output records.
// Reference: C++ Clipper64::BuildTree64 in clipper.engine.cpp lines 3027-3053
func (ve *VattiEngine) BuildTree64(polytree *PolyTree64, openPaths *Paths64) {
	polytree.Clear()
	if openPaths != nil {
		*openPaths = make(Paths64, 0)
	}

	// Iterate through all output records
	for _, outRec := range ve.outRecords {
		if outRec == nil || outRec.Pts == nil {
			continue
		}

		// Build the path from the output record
		path := ve.buildPathFromOutRec(outRec)
		if len(path) < 3 {
			continue // Skip degenerate polygons
		}

		// Store the path in the output record for hierarchy building
		// (In C++ this is done in BuildPath64, but we need it here)
		// Create a copy to store
		pathCopy := make(Path64, len(path))
		copy(pathCopy, path)
		outRec.Pts = ve.pathToOutPtList(pathCopy) // Ensure consistent structure

		// Check bounds and build hierarchy
		if ve.checkBounds(outRec) {
			ve.recursiveCheckOwners(outRec, polytree)
		}
	}
}

// pathToOutPtList converts a path back to an OutPt circular linked list
// This is needed to ensure the OutRec has a valid point list
func (ve *VattiEngine) pathToOutPtList(path Path64) *OutPt {
	if len(path) == 0 {
		return nil
	}

	// Create first point
	first := &OutPt{Pt: path[0]}
	prev := first

	// Create remaining points and link them
	for i := 1; i < len(path); i++ {
		op := &OutPt{Pt: path[i]}
		prev.Next = op
		op.Prev = prev
		prev = op
	}

	// Close the circular list
	prev.Next = first
	first.Prev = prev

	return first
}

// checkBounds validates that an output record has valid bounds.
// Reference: C++ CheckBounds in clipper.engine.cpp
func (ve *VattiEngine) checkBounds(outRec *OutRec) bool {
	if outRec == nil || outRec.Pts == nil {
		return false
	}

	// Build path and check if it has valid area
	path := ve.buildPathFromOutRec(outRec)
	if len(path) < 3 {
		return false
	}

	// Calculate bounds
	bounds := Bounds64(path)
	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		return false
	}

	return true
}

// recursiveCheckOwners validates the ownership hierarchy and builds the PolyTree.
// Reference: C++ RecursiveCheckOwners in clipper.engine.cpp lines 2967-2990
func (ve *VattiEngine) recursiveCheckOwners(outRec *OutRec, polypath *PolyPath64) {
	// Pre-condition: outRec will have valid bounds
	// Post-condition: if a valid path, outRec will have a polypath

	if outRec.PolyPath != nil {
		return // Already processed
	}

	// Walk up the owner chain to find valid parent
	for outRec.Owner != nil {
		// Check if owner's polygon contains this polygon
		if outRec.Owner.Pts != nil && ve.checkBounds(outRec.Owner) {
			ownerPath := ve.buildPathFromOutRec(outRec.Owner)
			childPath := ve.buildPathFromOutRec(outRec)

			// Check if owner bounds contain child bounds
			ownerBounds := Bounds64(ownerPath)
			childBounds := Bounds64(childPath)

			if ve.boundsContains(ownerBounds, childBounds) &&
				ve.path2ContainsPath1(childPath, ownerPath) {
				break // Found valid owner
			}
		}
		outRec.Owner = outRec.Owner.Owner
	}

	// Build the path for this output record
	path := ve.buildPathFromOutRec(outRec)

	// Add to tree hierarchy
	if outRec.Owner != nil {
		// This polygon is owned by another
		if outRec.Owner.PolyPath == nil {
			// Recursively process owner first
			ve.recursiveCheckOwners(outRec.Owner, polypath)
		}
		outRec.PolyPath = outRec.Owner.PolyPath.AddChild(path)
	} else {
		// This is a top-level polygon
		outRec.PolyPath = polypath.AddChild(path)
	}
}

// boundsContains checks if bounds 'a' fully contains bounds 'b'
func (ve *VattiEngine) boundsContains(a, b Rect64) bool {
	return b.Left >= a.Left && b.Right <= a.Right &&
		b.Top >= a.Top && b.Bottom <= a.Bottom
}

// path2ContainsPath1 checks if path2 contains path1 using point-in-polygon tests.
// Reference: C++ Path2ContainsPath1 in clipper.h
func (ve *VattiEngine) path2ContainsPath1(path1, path2 Path64) bool {
	// Check if all points of path1 are inside path2
	// We need at least 2 consecutive points inside to be certain
	// (to account for touching edges)

	insideCount := 0
	outsideCount := 0

	for _, pt := range path1 {
		result := PointInPolygon(pt, path2, ve.fillRule)
		switch result {
		case Inside:
			insideCount++
			if insideCount > 1 {
				return true // At least 2 points inside = contained
			}
		case Outside:
			outsideCount++
			if outsideCount > 1 {
				return false // At least 2 points outside = not contained
			}
		}
	}

	// If we got here, most points are on the boundary
	// Check the midpoint of the bounds as a tiebreaker
	if insideCount > 0 && outsideCount == 0 {
		return true
	}

	bounds1 := Bounds64(path1)
	midPt := bounds1.MidPoint()
	return PointInPolygon(midPt, path2, ve.fillRule) == Inside
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
	if e.VertexTop == nil {
		return
	}
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
// CRITICAL: Must process consecutive horizontal segments in a single call via while(true) loop
// DO NOT push back to horizontal queue - this causes infinite loops
func (ve *VattiEngine) DoHorizontal(horz *Edge) {
	y := horz.Bot.Y

	// Find vertex_max - the vertex where this horizontal terminates
	// For now, use VertexTop as the maxima vertex (simplified from C++)
	vertexMax := horz.VertexTop

	// Add initial output point if this edge is hot
	isHot := (horz.OutRec != nil)
	if isHot {
		pt := Point64{horz.CurrX, y}
		ve.addOutPt(horz, pt, "DoHorizontal:start")
	}

	// Outer loop: process consecutive horizontal segments
	// C++ line 2581: while (true) // loop through consec. horizontal edges
	loopCount := 0
	for {
		loopCount++

		if loopCount > 100 {
			break
		}

		// Find the maximum pair edge (where this horizontal edge terminates)
		var maxPair *Edge

		// Search for an edge that ends at the same vertex as this horizontal
		if horz.IsLeftBound {
			// Left bound - look to the right for paired edge
			e := horz.NextInAEL
			for e != nil {
				if e.VertexTop == vertexMax {
					maxPair = e
					break
				}
				if e.CurrX > horz.Top.X {
					break
				}
				e = e.NextInAEL
			}
		} else {
			// Right bound - look to the left for paired edge
			e := horz.PrevInAEL
			for e != nil {
				if e.VertexTop == vertexMax {
					maxPair = e
					break
				}
				if e.CurrX < horz.Top.X {
					break
				}
				e = e.PrevInAEL
			}
		}

		// Set up direction markers
		horzLeft, horzRight := ve.ResetHorzDirection(horz, maxPair)
		isLeftToRight := horz.Dx > 0

		// Inner loop: process edges that intersect this horizontal segment
		// C++ line 2587: while (e)
		var e *Edge
		if isLeftToRight {
			e = horz.NextInAEL
		} else {
			e = horz.PrevInAEL
		}

		for e != nil {
			// C++ line 2589: if (e->vertex_top == vertex_max)
			if e.VertexTop == vertexMax {
				// Found the maxima pair - add local maximum and delete both edges

				if isHot {
					// Advance horz to vertex_max if needed
					for horz.VertexTop != vertexMax {
						ve.addOutPt(horz, horz.Top, "DoHorizontal:advance_to_max")
						ve.UpdateEdgeIntoAEL(horz)
					}

					// Add local maximum poly
					if isLeftToRight {
						ve.addLocalMaxPoly(horz, e, horz.Top)
					} else {
						ve.addLocalMaxPoly(e, horz, horz.Top)
					}
				}

				// Delete both edges from AEL
				ve.removeEdgeFromAEL(e)
				ve.removeEdgeFromAEL(horz)
				return
			}

			// Check if we should stop processing edges
			// C++ line 2616: if (vertex_max != horz.vertex_top || IsOpenEnd(horz))
			if vertexMax != horz.VertexTop {
				// Stop when beyond horizontal bounds
				// C++ line 2619: if ((is_left_to_right && e->curr_x > horz_right) || ...
				if (isLeftToRight && e.CurrX > horzRight) || (!isLeftToRight && e.CurrX < horzLeft) {
					break
				}

				// Additional slope checks would go here (C++ lines 2622-2643)
				// Simplified for now - just check if at boundary
				if e.CurrX == horz.Top.X && !IsHorizontal(e) {
					nextVertex := ve.getNextVertex(horz)
					if nextVertex != nil {
						pt := nextVertex.Pt
						// Check slope comparison (simplified)
						if isLeftToRight {
							if ve.topX(e, pt.Y) >= pt.X {
								break
							}
						} else {
							if ve.topX(e, pt.Y) <= pt.X {
								break
							}
						}
					}
				}
			}

			// Intersect edges
			// C++ lines 2646-2662
			pt := Point64{e.CurrX, y}
			if isLeftToRight {
				ve.intersectEdges(horz, e, pt)
				ve.swapPositionsInAEL(horz, e)
				horz.CurrX = e.CurrX
				e = horz.NextInAEL
			} else {
				ve.intersectEdges(e, horz, pt)
				ve.swapPositionsInAEL(e, horz)
				horz.CurrX = e.CurrX
				e = horz.PrevInAEL
			}
		}

		// Check if finished with consecutive horizontals
		// C++ line 2687: else if (NextVertex(horz)->pt.y != horz.top.y)
		nextVertex := ve.getNextVertex(horz)
		// CRITICAL: Check if next segment is at the SAME Y level as the current horizontal
		// We must compare against the original Y (from horz.Bot.Y at start), not horz.Top.Y which changes
		if nextVertex == nil || nextVertex.Pt.Y != y {
			// No more consecutive horizontals at this Y level
			break
		}

		// Still more horizontals in bound to process
		// C++ lines 2691-2696
		if isHot {
			ve.addOutPt(horz, horz.Top, "DoHorizontal:continue")
		}
		ve.UpdateEdgeIntoAEL(horz)

		// Continue loop with updated horizontal edge
		// DO NOT push to queue - we handle it in this loop
	}

	// End of intermediate horizontal
	// C++ lines 2699-2705
	if isHot {
		pt := Point64{horz.Top.X, horz.Top.Y}
		ve.addOutPt(horz, pt, "DoHorizontal:final")
	}

	ve.UpdateEdgeIntoAEL(horz)
}
