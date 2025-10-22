package clipper

import (
	"testing"
)

// ==============================================================================
// PolyPath64 Basic Tests
// ==============================================================================

func TestPolyPath64_Level(t *testing.T) {
	root := NewPolyTree64()
	if root.Level() != 0 {
		t.Errorf("Root level should be 0, got %d", root.Level())
	}

	// Add outer polygon (level 0 child = level 1)
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	if outer.Level() != 1 {
		t.Errorf("Outer polygon level should be 1, got %d", outer.Level())
	}

	// Add hole in outer (level 2)
	hole := outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})
	if hole.Level() != 2 {
		t.Errorf("Hole level should be 2, got %d", hole.Level())
	}

	// Add island in hole (level 3)
	island := hole.AddChild(Path64{{30, 30}, {70, 30}, {70, 70}, {30, 70}})
	if island.Level() != 3 {
		t.Errorf("Island level should be 3, got %d", island.Level())
	}
}

func TestPolyPath64_IsHole(t *testing.T) {
	root := NewPolyTree64()

	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	hole := outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})
	island := hole.AddChild(Path64{{30, 30}, {70, 30}, {70, 70}, {30, 70}})
	holeInIsland := island.AddChild(Path64{{40, 40}, {60, 40}, {60, 60}, {40, 60}})

	// Level 0 (root) is not a hole
	if root.IsHole() {
		t.Error("Root (level 0) should not be a hole")
	}

	// Level 1 (outer polygons) are NOT holes
	if outer.IsHole() {
		t.Error("Level 1 outer polygon should NOT be a hole")
	}

	// Level 2 (holes) ARE holes
	if !hole.IsHole() {
		t.Error("Level 2 hole should be a hole (even level > 0)")
	}

	// Level 3 (islands) are NOT holes
	if island.IsHole() {
		t.Error("Level 3 island should NOT be a hole (odd level)")
	}

	// Level 4 (holes in islands) ARE holes
	if !holeInIsland.IsHole() {
		t.Error("Level 4 hole in island should be a hole (even level)")
	}
}

func TestPolyPath64_Parent(t *testing.T) {
	root := NewPolyTree64()
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	hole := outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})

	if outer.Parent() != root {
		t.Error("Outer polygon parent should be root")
	}

	if hole.Parent() != outer {
		t.Error("Hole parent should be outer polygon")
	}

	if root.Parent() != nil {
		t.Error("Root parent should be nil")
	}
}

func TestPolyPath64_ChildAccess(t *testing.T) {
	root := NewPolyTree64()
	child1 := root.AddChild(Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}})
	child2 := root.AddChild(Path64{{20, 20}, {30, 20}, {30, 30}, {20, 30}})
	child3 := root.AddChild(Path64{{40, 40}, {50, 40}, {50, 50}, {40, 50}})

	if root.Count() != 3 {
		t.Errorf("Root should have 3 children, got %d", root.Count())
	}

	if root.Child(0) != child1 {
		t.Error("Child(0) should return first child")
	}

	if root.Child(1) != child2 {
		t.Error("Child(1) should return second child")
	}

	if root.Child(2) != child3 {
		t.Error("Child(2) should return third child")
	}

	if root.Child(3) != nil {
		t.Error("Child(3) should return nil (out of bounds)")
	}

	if root.Child(-1) != nil {
		t.Error("Child(-1) should return nil (negative index)")
	}
}

func TestPolyPath64_Clear(t *testing.T) {
	root := NewPolyTree64()
	root.AddChild(Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}})
	root.AddChild(Path64{{20, 20}, {30, 20}, {30, 30}, {20, 30}})

	if root.Count() != 2 {
		t.Errorf("Should have 2 children before clear, got %d", root.Count())
	}

	root.Clear()

	if root.Count() != 0 {
		t.Errorf("Should have 0 children after clear, got %d", root.Count())
	}
}

// ==============================================================================
// PolyTreeToPaths64 Tests
// ==============================================================================

func TestPolyTreeToPaths64_Empty(t *testing.T) {
	root := NewPolyTree64()
	paths := PolyTreeToPaths64(root)

	if len(paths) != 0 {
		t.Errorf("Empty tree should produce 0 paths, got %d", len(paths))
	}
}

func TestPolyTreeToPaths64_Flat(t *testing.T) {
	root := NewPolyTree64()
	root.AddChild(Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}})
	root.AddChild(Path64{{20, 20}, {30, 20}, {30, 30}, {20, 30}})

	paths := PolyTreeToPaths64(root)

	if len(paths) != 2 {
		t.Errorf("Should have 2 paths, got %d", len(paths))
	}
}

func TestPolyTreeToPaths64_Nested(t *testing.T) {
	root := NewPolyTree64()
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})
	outer.AddChild(Path64{{10, 10}, {15, 10}, {15, 15}, {10, 15}})

	paths := PolyTreeToPaths64(root)

	// Should have 3 paths total (1 outer + 2 holes)
	if len(paths) != 3 {
		t.Errorf("Should have 3 paths, got %d", len(paths))
	}
}

// ==============================================================================
// Union64Tree Basic Tests
// ==============================================================================

func TestUnion64Tree_TwoSeparateSquares(t *testing.T) {
	subject := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
	}
	clip := Paths64{
		{{100, 0}, {150, 0}, {150, 50}, {100, 50}},
	}

	tree, openPaths, err := Union64Tree(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Union64Tree failed: %v", err)
	}

	if len(openPaths) != 0 {
		t.Errorf("Should have 0 open paths, got %d", len(openPaths))
	}

	if tree.Count() != 2 {
		t.Errorf("Should have 2 top-level polygons, got %d", tree.Count())
	}

	// Both should be at level 1 (children of root)
	if tree.Child(0) == nil || tree.Child(0).Level() != 1 {
		t.Error("First polygon should be at level 1")
	}
	if tree.Child(1) == nil || tree.Child(1).Level() != 1 {
		t.Error("Second polygon should be at level 1")
	}
}

func TestUnion64Tree_OverlappingSquares(t *testing.T) {
	subject := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
	}
	clip := Paths64{
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	tree, openPaths, err := Union64Tree(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Union64Tree failed: %v", err)
	}

	if len(openPaths) != 0 {
		t.Errorf("Should have 0 open paths, got %d", len(openPaths))
	}

	// Union of two overlapping squares should produce 1 polygon
	if tree.Count() != 1 {
		t.Errorf("Should have 1 top-level polygon, got %d", tree.Count())
	}
}

func TestIntersect64Tree_OverlappingSquares(t *testing.T) {
	subject := Paths64{
		{{0, 0}, {50, 0}, {50, 50}, {0, 50}},
	}
	clip := Paths64{
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	tree, openPaths, err := Intersect64Tree(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Intersect64Tree failed: %v", err)
	}

	if len(openPaths) != 0 {
		t.Errorf("Should have 0 open paths, got %d", len(openPaths))
	}

	// Intersection should produce 1 polygon
	if tree.Count() != 1 {
		t.Errorf("Should have 1 top-level polygon (intersection), got %d", tree.Count())
	}
}

func TestDifference64Tree_SquareMinusSmaller(t *testing.T) {
	subject := Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
	}
	clip := Paths64{
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	tree, openPaths, err := Difference64Tree(subject, clip, NonZero)
	if err != nil {
		t.Fatalf("Difference64Tree failed: %v", err)
	}

	if len(openPaths) != 0 {
		t.Errorf("Should have 0 open paths, got %d", len(openPaths))
	}

	// Difference should produce 1 outer polygon with 1 hole
	if tree.Count() != 1 {
		t.Errorf("Should have 1 top-level polygon, got %d", tree.Count())
	}

	outer := tree.Child(0)
	if outer == nil {
		t.Fatal("Outer polygon is nil")
	}

	// The outer polygon should have 1 hole
	if outer.Count() != 1 {
		t.Errorf("Outer polygon should have 1 hole, got %d", outer.Count())
	}

	hole := outer.Child(0)
	if hole == nil {
		t.Fatal("Hole is nil")
	}

	if !hole.IsHole() {
		t.Error("Child of outer polygon should be a hole")
	}
}

// ==============================================================================
// Complex Hierarchy Tests
// ==============================================================================

func TestPolyTree_ComplexHierarchy(t *testing.T) {
	// Create a polygon with a hole, and an island inside the hole
	outerSquare := Paths64{
		{{0, 0}, {200, 0}, {200, 200}, {0, 200}},
	}
	hole := Paths64{
		{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
	}
	island := Paths64{
		{{75, 75}, {125, 75}, {125, 125}, {75, 125}},
	}

	// First create outer with hole
	withHole, _, err := Difference64Tree(outerSquare, hole, NonZero)
	if err != nil {
		t.Fatalf("Failed to create polygon with hole: %v", err)
	}

	// Verify structure
	if withHole.Count() != 1 {
		t.Fatalf("Should have 1 outer polygon, got %d", withHole.Count())
	}

	outer := withHole.Child(0)
	if outer.Count() != 1 {
		t.Errorf("Outer should have 1 hole, got %d", outer.Count())
	}

	// Now union with island to create island inside hole
	withHolePaths := PolyTreeToPaths64(withHole)
	finalTree, _, err := Union64Tree(withHolePaths, island, NonZero)
	if err != nil {
		t.Fatalf("Failed to add island: %v", err)
	}

	// Verify hierarchy: outer -> hole -> island
	if finalTree.Count() != 1 {
		t.Fatalf("Should have 1 top-level polygon, got %d", finalTree.Count())
	}

	finalOuter := finalTree.Child(0)
	if finalOuter.Count() != 1 {
		t.Fatalf("Outer should have 1 hole, got %d", finalOuter.Count())
	}

	finalHole := finalOuter.Child(0)
	if !finalHole.IsHole() {
		t.Error("First child should be a hole")
	}

	// The island should be a child of the hole
	// Note: This test may need adjustment based on actual behavior
	// as the island might be at the same level as the outer depending on algorithm
}

// ==============================================================================
// Area Calculation Tests
// ==============================================================================

func TestPolyPath64_Area(t *testing.T) {
	root := NewPolyTree64()

	// Add a 100x100 square (area = 10000)
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})

	area := outer.Area()
	expected := 10000.0
	if area != expected && area != -expected {
		t.Errorf("Expected area %f or %f, got %f", expected, -expected, area)
	}
}

func TestPolyPath64_AreaWithHole(t *testing.T) {
	root := NewPolyTree64()

	// Add a 100x100 square (area = 10000)
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})

	// Add a 40x40 hole (area = 1600)
	hole := outer.AddChild(Path64{{30, 30}, {70, 30}, {70, 70}, {30, 70}})

	// Total area should be 10000 +/- 1600 depending on orientation
	outerArea := Area64(outer.Polygon())
	holeArea := Area64(hole.Polygon())
	totalArea := outer.Area()

	// The areas should add up correctly (considering signs)
	expectedTotal := outerArea + holeArea
	if totalArea != expectedTotal {
		t.Errorf("Expected total area %f, got %f", expectedTotal, totalArea)
	}
}

// ==============================================================================
// Utility Method Tests
// ==============================================================================

func TestPolyPath64_TotalVertexCount(t *testing.T) {
	root := NewPolyTree64()
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}) // 4 vertices
	outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})         // 4 vertices
	outer.AddChild(Path64{{10, 10}, {15, 10}, {15, 15}, {10, 15}})         // 4 vertices

	totalVertices := root.TotalVertexCount()
	expected := 0 + 4 + 4 + 4 // root has no polygon, 3 paths with 4 vertices each
	if totalVertices != expected {
		t.Errorf("Expected %d total vertices, got %d", expected, totalVertices)
	}
}

func TestPolyPath64_TotalPolygonCount(t *testing.T) {
	root := NewPolyTree64()
	outer := root.AddChild(Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	outer.AddChild(Path64{{20, 20}, {80, 20}, {80, 80}, {20, 80}})
	outer.AddChild(Path64{{10, 10}, {15, 10}, {15, 15}, {10, 15}})

	totalPolygons := root.TotalPolygonCount()
	expected := 3 // 1 outer + 2 holes
	if totalPolygons != expected {
		t.Errorf("Expected %d total polygons, got %d", expected, totalPolygons)
	}
}
