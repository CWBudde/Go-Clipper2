package clipper

// This file implements the PolyTree/PolyPath hierarchical polygon structure
// Reference: clipper.engine.h lines 298-383

// PolyPath64 represents a node in a hierarchical polygon tree structure.
// Each node contains a polygon path and may have child nodes representing
// holes (odd levels) or nested islands (even levels).
//
// Hierarchy levels:
//   - Level 0: Root level (outer polygons)
//   - Level 1: Holes in outer polygons
//   - Level 2: Islands within holes
//   - Level 3: Holes in islands
//   - And so on...
//
// Use IsHole() to determine if a node represents a hole (odd levels).
type PolyPath64 struct {
	parent   *PolyPath64
	children []*PolyPath64
	polygon  Path64
}

// PolyTree64 is the root of a hierarchical polygon tree.
// It's a PolyPath64 node with no polygon data of its own.
type PolyTree64 = PolyPath64

// NewPolyTree64 creates a new empty polygon tree
func NewPolyTree64() *PolyTree64 {
	return &PolyTree64{
		parent:   nil,
		children: make([]*PolyPath64, 0),
		polygon:  nil,
	}
}

// Level returns the depth of this node in the tree.
// Root children are at level 0, their children at level 1, etc.
//   - Level 0, 2, 4, ... = outer polygons or islands
//   - Level 1, 3, 5, ... = holes
func (pp *PolyPath64) Level() int {
	level := 0
	p := pp.parent
	for p != nil {
		level++
		p = p.parent
	}
	return level
}

// IsHole returns true if this node represents a hole (even level > 0).
// Holes are at levels 2, 4, 6, etc.
// Reference: C++ PolyPath::IsHole() - returns lvl && !(lvl & 1)
func (pp *PolyPath64) IsHole() bool {
	level := pp.Level()
	// Even levels except level 0 are holes
	// Level 0 = root (no polygon)
	// Level 1 = outer polygons (NOT holes)
	// Level 2 = holes in outer polygons (holes)
	// Level 3 = islands in holes (NOT holes)
	// Level 4 = holes in islands (holes)
	return level > 0 && (level&1) == 0
}

// Parent returns the parent node, or nil if this is a root child
func (pp *PolyPath64) Parent() *PolyPath64 {
	return pp.parent
}

// Child returns the child node at the specified index.
// Returns nil if index is out of bounds.
func (pp *PolyPath64) Child(index int) *PolyPath64 {
	if index < 0 || index >= len(pp.children) {
		return nil
	}
	return pp.children[index]
}

// Count returns the number of child nodes
func (pp *PolyPath64) Count() int {
	return len(pp.children)
}

// Polygon returns the polygon path for this node.
// Returns empty path for root nodes.
func (pp *PolyPath64) Polygon() Path64 {
	return pp.polygon
}

// AddChild adds a new child node with the specified polygon path.
// This is used internally by the clipping algorithm.
func (pp *PolyPath64) AddChild(path Path64) *PolyPath64 {
	child := &PolyPath64{
		parent:   pp,
		children: make([]*PolyPath64, 0),
		polygon:  path,
	}
	pp.children = append(pp.children, child)
	return child
}

// Clear removes all child nodes from this tree
func (pp *PolyPath64) Clear() {
	pp.children = make([]*PolyPath64, 0)
}

// Area calculates the total signed area of this polygon and all its descendants.
// Holes (odd levels) subtract area, islands (even levels) add area.
func (pp *PolyPath64) Area() float64 {
	area := Area64(pp.polygon)
	for _, child := range pp.children {
		area += child.Area()
	}
	return area
}

// Children returns a slice of all child nodes for iteration
func (pp *PolyPath64) Children() []*PolyPath64 {
	return pp.children
}

// PolyTreeToPaths64 converts a hierarchical PolyTree to a flat Paths64 slice.
// All polygons at all levels are flattened into a single slice.
// The hierarchy information is lost in this conversion.
func PolyTreeToPaths64(polytree *PolyTree64) Paths64 {
	result := make(Paths64, 0)
	polyPathToPaths64(polytree, &result)
	return result
}

// polyPathToPaths64 recursively flattens a PolyPath to Paths64
func polyPathToPaths64(polypath *PolyPath64, paths *Paths64) {
	if len(polypath.polygon) > 0 {
		*paths = append(*paths, polypath.polygon)
	}
	for _, child := range polypath.children {
		polyPathToPaths64(child, paths)
	}
}

// TotalVertexCount returns the total number of vertices in the entire tree
func (pp *PolyPath64) TotalVertexCount() int {
	count := len(pp.polygon)
	for _, child := range pp.children {
		count += child.TotalVertexCount()
	}
	return count
}

// TotalPolygonCount returns the total number of polygons in the entire tree
func (pp *PolyPath64) TotalPolygonCount() int {
	count := 0
	if len(pp.polygon) > 0 {
		count = 1
	}
	for _, child := range pp.children {
		count += child.TotalPolygonCount()
	}
	return count
}

// ==============================================================================
// 32-bit PolyTree/PolyPath Implementation
// ==============================================================================

// PolyPath32 represents a node in a hierarchical polygon tree structure (32-bit coordinates).
// Each node contains a polygon path and may have child nodes representing
// holes (odd levels) or nested islands (even levels).
//
// Hierarchy levels:
//   - Level 0: Root level (outer polygons)
//   - Level 1: Holes in outer polygons
//   - Level 2: Islands within holes
//   - Level 3: Holes in islands
//   - And so on...
//
// Use IsHole() to determine if a node represents a hole (odd levels).
type PolyPath32 struct {
	parent   *PolyPath32
	children []*PolyPath32
	polygon  Path32
}

// PolyTree32 is the root of a hierarchical polygon tree (32-bit coordinates).
// It's a PolyPath32 node with no polygon data of its own.
type PolyTree32 = PolyPath32

// NewPolyTree32 creates a new empty polygon tree (32-bit coordinates)
func NewPolyTree32() *PolyTree32 {
	return &PolyTree32{
		parent:   nil,
		children: make([]*PolyPath32, 0),
		polygon:  nil,
	}
}

// Level returns the depth of this node in the tree.
// Root children are at level 0, their children at level 1, etc.
//   - Level 0, 2, 4, ... = outer polygons or islands
//   - Level 1, 3, 5, ... = holes
func (pp *PolyPath32) Level() int {
	level := 0
	p := pp.parent
	for p != nil {
		level++
		p = p.parent
	}
	return level
}

// IsHole returns true if this node represents a hole (even level > 0).
// Holes are at levels 2, 4, 6, etc.
func (pp *PolyPath32) IsHole() bool {
	level := pp.Level()
	return level > 0 && (level&1) == 0
}

// Parent returns the parent node, or nil if this is a root child
func (pp *PolyPath32) Parent() *PolyPath32 {
	return pp.parent
}

// Child returns the child node at the specified index.
// Returns nil if index is out of bounds.
func (pp *PolyPath32) Child(index int) *PolyPath32 {
	if index < 0 || index >= len(pp.children) {
		return nil
	}
	return pp.children[index]
}

// Count returns the number of child nodes
func (pp *PolyPath32) Count() int {
	return len(pp.children)
}

// Polygon returns the polygon path for this node.
// Returns empty path for root nodes.
func (pp *PolyPath32) Polygon() Path32 {
	return pp.polygon
}

// AddChild adds a new child node with the specified polygon path.
// This is used internally by the clipping algorithm.
func (pp *PolyPath32) AddChild(path Path32) *PolyPath32 {
	child := &PolyPath32{
		parent:   pp,
		children: make([]*PolyPath32, 0),
		polygon:  path,
	}
	pp.children = append(pp.children, child)
	return child
}

// Clear removes all child nodes from this tree
func (pp *PolyPath32) Clear() {
	pp.children = make([]*PolyPath32, 0)
}

// Area calculates the total signed area of this polygon and all its descendants.
// Holes (odd levels) subtract area, islands (even levels) add area.
func (pp *PolyPath32) Area() float64 {
	area := Area32(pp.polygon)
	for _, child := range pp.children {
		area += child.Area()
	}
	return area
}

// Children returns a slice of all child nodes for iteration
func (pp *PolyPath32) Children() []*PolyPath32 {
	return pp.children
}

// PolyTreeToPaths32 converts a hierarchical PolyTree to a flat Paths32 slice.
// All polygons at all levels are flattened into a single slice.
// The hierarchy information is lost in this conversion.
func PolyTreeToPaths32(polytree *PolyTree32) Paths32 {
	result := make(Paths32, 0)
	polyPathToPaths32(polytree, &result)
	return result
}

// polyPathToPaths32 recursively flattens a PolyPath to Paths32
func polyPathToPaths32(polypath *PolyPath32, paths *Paths32) {
	if len(polypath.polygon) > 0 {
		*paths = append(*paths, polypath.polygon)
	}
	for _, child := range polypath.children {
		polyPathToPaths32(child, paths)
	}
}

// TotalVertexCount returns the total number of vertices in the entire tree
func (pp *PolyPath32) TotalVertexCount() int {
	count := len(pp.polygon)
	for _, child := range pp.children {
		count += child.TotalVertexCount()
	}
	return count
}

// TotalPolygonCount returns the total number of polygons in the entire tree
func (pp *PolyPath32) TotalPolygonCount() int {
	count := 0
	if len(pp.polygon) > 0 {
		count = 1
	}
	for _, child := range pp.children {
		count += child.TotalPolygonCount()
	}
	return count
}

// PolyTree64To32 converts a 64-bit PolyTree to 32-bit with overflow detection
func PolyTree64To32(tree64 *PolyTree64) (*PolyTree32, error) {
	tree32 := NewPolyTree32()
	if err := polyPath64To32(tree64, tree32); err != nil {
		return nil, err
	}
	return tree32, nil
}

// polyPath64To32 recursively converts a 64-bit PolyPath to 32-bit
func polyPath64To32(src64 *PolyPath64, dst32 *PolyPath32) error {
	// Convert polygon if it exists
	if len(src64.Polygon()) > 0 {
		polygon32, err := Path64ToPath32(src64.Polygon())
		if err != nil {
			return err
		}
		dst32.polygon = polygon32
	}

	// Convert children
	for _, child64 := range src64.Children() {
		polygon32, err := Path64ToPath32(child64.Polygon())
		if err != nil {
			return err
		}
		child32 := dst32.AddChild(polygon32)

		// Recursively convert grandchildren
		if err := polyPath64To32(child64, child32); err != nil {
			return err
		}
	}

	return nil
}

// PolyTree32To64 converts a 32-bit PolyTree to 64-bit (always safe)
func PolyTree32To64(tree32 *PolyTree32) *PolyTree64 {
	tree64 := NewPolyTree64()
	polyPath32To64(tree32, tree64)
	return tree64
}

// polyPath32To64 recursively converts a 32-bit PolyPath to 64-bit
func polyPath32To64(src32 *PolyPath32, dst64 *PolyPath64) {
	// Convert polygon if it exists
	if len(src32.Polygon()) > 0 {
		dst64.polygon = Path32ToPath64(src32.Polygon())
	}

	// Convert children
	for _, child32 := range src32.Children() {
		polygon64 := Path32ToPath64(child32.Polygon())
		child64 := dst64.AddChild(polygon64)

		// Recursively convert grandchildren
		polyPath32To64(child32, child64)
	}
}
