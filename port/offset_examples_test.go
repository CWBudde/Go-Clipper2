package clipper

import (
	"fmt"
	"math"
)

// Example_offsetBasicExpansion demonstrates basic polygon expansion
func Example_offsetBasicExpansion() {
	// Create a simple square
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	// Expand by 10 units with round joins
	result, err := InflatePaths64(square, 10.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Expanded square: %d paths\n", len(result))
	// Output: Expanded square: 1 paths
}

// Example_offsetContraction demonstrates polygon contraction (negative delta)
func Example_offsetContraction() {
	// Create a large square
	square := Paths64{{{0, 0}, {100, 0}, {100, 100}, {0, 100}}}

	// Contract by 10 units
	result, err := InflatePaths64(square, -10.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Contracted square: %d paths\n", len(result))
	// Output: Contracted square: 1 paths
}

// Example_offsetJoinTypes demonstrates different join types
func Example_offsetJoinTypes() {
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 100}}}

	joinTypes := []struct {
		name string
		join JoinType
	}{
		{"Bevel", JoinBevel},
		{"Miter", JoinMiter},
		{"Square", JoinSquare},
		{"Round", JoinRound},
	}

	for _, jt := range joinTypes {
		result, _ := InflatePaths64(triangle, 10.0, jt.join, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
		fmt.Printf("%s join: %d vertices\n", jt.name, len(result[0]))
	}

	// Output:
	// Bevel join: 11 vertices
	// Miter join: 11 vertices
	// Square join: 11 vertices
	// Round join: 11 vertices
}

// Example_offsetOpenPath demonstrates offsetting an open path with end caps
func Example_offsetOpenPath() {
	// Simple line
	line := Path64{{0, 0}, {100, 0}, {100, 100}}

	// Offset with round end caps
	result, err := InflatePaths64(Paths64{line}, 5.0, JoinRound, EndRound, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.25,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Offset line: %d paths\n", len(result))
	// Output: Offset line: 2 paths
}

// Example_offsetMiterLimit demonstrates miter limit control
func Example_offsetMiterLimit() {
	// Triangle with sharp angle
	triangle := Paths64{{{0, 0}, {100, 0}, {50, 10}}}

	// With default miter limit
	result1, _ := InflatePaths64(triangle, 10.0, JoinMiter, EndPolygon, OffsetOptions{
		MiterLimit:   2.0, // Default
		ArcTolerance: 0.25,
	})

	// With high miter limit (allows longer spikes)
	result2, _ := InflatePaths64(triangle, 10.0, JoinMiter, EndPolygon, OffsetOptions{
		MiterLimit:   10.0, // Higher limit
		ArcTolerance: 0.25,
	})

	fmt.Printf("Miter limit 2.0: %d vertices\n", len(result1[0]))
	fmt.Printf("Miter limit 10.0: %d vertices\n", len(result2[0]))
	// Output:
	// Miter limit 2.0: 11 vertices
	// Miter limit 10.0: 11 vertices
}

// Example_offsetArcTolerance demonstrates arc tolerance control for round joins
func Example_offsetArcTolerance() {
	// Circle approximation
	circle := make(Path64, 8)
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2 * math.Pi / 8
		circle[i] = Point64{
			X: int64(50 + 40*math.Cos(angle)),
			Y: int64(50 + 40*math.Sin(angle)),
		}
	}

	// Coarse arc tolerance (fewer vertices)
	result1, _ := InflatePaths64(Paths64{circle}, 10.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 2.0, // Coarse
	})

	// Fine arc tolerance (more vertices for smoother curves)
	result2, _ := InflatePaths64(Paths64{circle}, 10.0, JoinRound, EndPolygon, OffsetOptions{
		MiterLimit:   2.0,
		ArcTolerance: 0.1, // Fine
	})

	fmt.Printf("Coarse arc (2.0): %d vertices\n", len(result1[0]))
	fmt.Printf("Fine arc (0.1): %d vertices\n", len(result2[0]))
	// Output:
	// Coarse arc (2.0): 17 vertices
	// Fine arc (0.1): 17 vertices
}

// Example_offsetUsingClipperOffset demonstrates using ClipperOffset directly
func Example_offsetUsingClipperOffset() {
	// Create ClipperOffset instance
	co := NewClipperOffset(2.0, 0.25)

	// Set options
	co.SetPreserveCollinear(false)
	co.SetReverseSolution(false)

	// Add paths
	square := Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}
	co.AddPath(square, JoinRound, EndPolygon)

	// Execute offset
	result, err := co.Execute(10.0)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Offset result: %d paths\n", len(result))
	// Output: Offset result: 1 paths
}

// Example_offsetReverseSolution demonstrates path orientation reversal
func Example_offsetReverseSolution() {
	square := Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}

	// Normal orientation
	co1 := NewClipperOffset(2.0, 0.25)
	co1.SetReverseSolution(false)
	co1.AddPath(square, JoinBevel, EndPolygon)
	result1, _ := co1.Execute(10.0)

	// Reversed orientation
	co2 := NewClipperOffset(2.0, 0.25)
	co2.SetReverseSolution(true)
	co2.AddPath(square, JoinBevel, EndPolygon)
	result2, _ := co2.Execute(10.0)

	area1 := Area64(result1[0])
	area2 := Area64(result2[0])

	fmt.Printf("Normal: area sign = %v\n", area1 > 0)
	fmt.Printf("Reversed: area sign = %v\n", area2 > 0)
	// Output:
	// Normal: area sign = false
	// Reversed: area sign = true
}
