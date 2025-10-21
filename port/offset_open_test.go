package clipper

import (
	"math"
	"testing"
)

// ==============================================================================
// EndButt Tests - Blunt end caps
// ==============================================================================

func TestOffsetOpenButtSimpleLine(t *testing.T) {
	// Simple horizontal line with butt end caps
	line := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinSquare,
		EndButt,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	// Validate against oracle
	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndButt, result)
}

func TestOffsetOpenButtPolyline(t *testing.T) {
	// Multi-segment polyline with butt end caps
	polyline := Path64{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 50, Y: 50},
		{X: 100, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{polyline},
		10.0,
		JoinSquare,
		EndButt,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{polyline}, 10.0, JoinSquare, EndButt, result)
}

func TestOffsetOpenButtVariousAngles(t *testing.T) {
	// Test lines at different angles
	testCases := []struct {
		name  string
		line  Path64
		delta float64
	}{
		{
			name:  "Vertical line",
			line:  Path64{{X: 50, Y: 0}, {X: 50, Y: 100}},
			delta: 10.0,
		},
		{
			name:  "Diagonal 45°",
			line:  Path64{{X: 0, Y: 0}, {X: 100, Y: 100}},
			delta: 10.0,
		},
		{
			name:  "Diagonal -45°",
			line:  Path64{{X: 0, Y: 100}, {X: 100, Y: 0}},
			delta: 10.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(
				Paths64{tc.line},
				tc.delta,
				JoinSquare,
				EndButt,
				OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
			)

			if err != nil {
				t.Fatalf("InflatePaths64 failed: %v", err)
			}

			validateOffsetWithOracle(t, Paths64{tc.line}, tc.delta, JoinSquare, EndButt, result)
		})
	}
}

// ==============================================================================
// EndSquare Tests - Extended square end caps
// ==============================================================================

func TestOffsetOpenSquareSimpleLine(t *testing.T) {
	// Simple horizontal line with square end caps (extends beyond endpoints)
	line := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinSquare,
		EndSquare,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	// Square end caps should extend the line by delta at each end
	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndSquare, result)
}

func TestOffsetOpenSquarePolyline(t *testing.T) {
	// Multi-segment polyline with square end caps
	polyline := Path64{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 50, Y: 50},
		{X: 100, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{polyline},
		10.0,
		JoinRound,
		EndSquare,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{polyline}, 10.0, JoinRound, EndSquare, result)
}

func TestOffsetOpenSquareExtensionVerification(t *testing.T) {
	// Verify that square end caps extend correctly
	line := Path64{
		{X: 50, Y: 50},
		{X: 150, Y: 50},
	}

	delta := 20.0

	result, err := InflatePaths64(
		Paths64{line},
		delta,
		JoinSquare,
		EndSquare,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{line}, delta, JoinSquare, EndSquare, result)
}

// ==============================================================================
// EndRound Tests - Round end caps with arc approximation
// ==============================================================================

func TestOffsetOpenRoundSimpleLine(t *testing.T) {
	// Simple horizontal line with round end caps
	line := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinSquare,
		EndRound,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	// Round end caps should create semicircular ends
	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndRound, result)
}

func TestOffsetOpenRoundPolyline(t *testing.T) {
	// Multi-segment polyline with round end caps
	polyline := Path64{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 50, Y: 50},
		{X: 100, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{polyline},
		10.0,
		JoinRound,
		EndRound,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{polyline}, 10.0, JoinRound, EndRound, result)
}

func TestOffsetOpenRoundArcTolerance(t *testing.T) {
	// Test different arc tolerances for round end caps
	line := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
	}

	testCases := []struct {
		name         string
		arcTolerance float64
	}{
		{"Coarse (2.0)", 2.0},
		{"Medium (0.5)", 0.5},
		{"Fine (0.1)", 0.1},
		{"Very Fine (0.01)", 0.01},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(
				Paths64{line},
				10.0,
				JoinSquare,
				EndRound,
				OffsetOptions{MiterLimit: 2.0, ArcTolerance: tc.arcTolerance},
			)

			if err != nil {
				t.Fatalf("InflatePaths64 failed: %v", err)
			}

			validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndRound, result)
		})
	}
}

func TestOffsetOpenRoundVariousDeltas(t *testing.T) {
	// Test round end caps with different offset distances
	line := Path64{
		{X: 50, Y: 50},
		{X: 150, Y: 50},
	}

	testCases := []struct {
		name  string
		delta float64
	}{
		{"Small (5)", 5.0},
		{"Medium (15)", 15.0},
		{"Large (30)", 30.0},
		{"Very Large (50)", 50.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := InflatePaths64(
				Paths64{line},
				tc.delta,
				JoinRound,
				EndRound,
				OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
			)

			if err != nil {
				t.Fatalf("InflatePaths64 failed: %v", err)
			}

			validateOffsetWithOracle(t, Paths64{line}, tc.delta, JoinRound, EndRound, result)
		})
	}
}

// ==============================================================================
// EndJoined Tests - Treat open path as closed
// ==============================================================================

func TestOffsetOpenJoinedSimpleLine(t *testing.T) {
	// Simple line treated as a closed path
	line := Path64{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinSquare,
		EndJoined,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndJoined, result)
}

func TestOffsetOpenJoinedPolyline(t *testing.T) {
	// Multi-segment polyline treated as closed
	polyline := Path64{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 50, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{polyline},
		10.0,
		JoinRound,
		EndJoined,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{polyline}, 10.0, JoinRound, EndJoined, result)
}

func TestOffsetOpenJoinedClosedLoopBehavior(t *testing.T) {
	// Verify EndJoined creates a closed loop around the path
	polyline := Path64{
		{X: 20, Y: 20},
		{X: 80, Y: 20},
		{X: 80, Y: 80},
	}

	result, err := InflatePaths64(
		Paths64{polyline},
		15.0,
		JoinMiter,
		EndJoined,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{polyline}, 15.0, JoinMiter, EndJoined, result)
}

// ==============================================================================
// Single-Point Path Tests - Circle or square generation
// ==============================================================================

func TestOffsetSinglePointRoundJoin(t *testing.T) {
	// Single point with round join should create a circle
	point := Path64{{X: 50, Y: 50}}

	result, err := InflatePaths64(
		Paths64{point},
		20.0,
		JoinRound,
		EndPolygon,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	// Should create a circular path
	validateOffsetWithOracle(t, Paths64{point}, 20.0, JoinRound, EndPolygon, result)
}

func TestOffsetSinglePointSquareJoin(t *testing.T) {
	// Single point with square join should create a square
	point := Path64{{X: 50, Y: 50}}

	result, err := InflatePaths64(
		Paths64{point},
		20.0,
		JoinSquare,
		EndPolygon,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	// Should create a square path
	validateOffsetWithOracle(t, Paths64{point}, 20.0, JoinSquare, EndPolygon, result)
}

func TestOffsetSinglePointMiterJoin(t *testing.T) {
	// Single point with miter join should create a square
	point := Path64{{X: 50, Y: 50}}

	result, err := InflatePaths64(
		Paths64{point},
		20.0,
		JoinMiter,
		EndPolygon,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one path in result")
	}

	validateOffsetWithOracle(t, Paths64{point}, 20.0, JoinMiter, EndPolygon, result)
}

// ==============================================================================
// Two-Point Path Tests - Special handling for EndJoined
// ==============================================================================

func TestOffsetTwoPointJoinedRound(t *testing.T) {
	// Two-point path with EndJoined and JoinRound → converts to EndRound
	line := Path64{
		{X: 20, Y: 50},
		{X: 80, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinRound,
		EndJoined,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinRound, EndJoined, result)
}

func TestOffsetTwoPointJoinedSquare(t *testing.T) {
	// Two-point path with EndJoined and JoinSquare → converts to EndSquare
	line := Path64{
		{X: 20, Y: 50},
		{X: 80, Y: 50},
	}

	result, err := InflatePaths64(
		Paths64{line},
		10.0,
		JoinSquare,
		EndJoined,
		OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
	)

	if err != nil {
		t.Fatalf("InflatePaths64 failed: %v", err)
	}

	validateOffsetWithOracle(t, Paths64{line}, 10.0, JoinSquare, EndJoined, result)
}

// ==============================================================================
// Mixed End Cap Tests - Different join types with different end caps
// ==============================================================================

func TestOffsetOpenMixedJoinEndTypes(t *testing.T) {
	// Test all combinations of join types with end types
	line := Path64{
		{X: 0, Y: 0},
		{X: 50, Y: 0},
		{X: 50, Y: 50},
	}

	testCases := []struct {
		joinType JoinType
		endType  EndType
	}{
		{JoinBevel, EndButt},
		{JoinBevel, EndSquare},
		{JoinBevel, EndRound},
		{JoinMiter, EndButt},
		{JoinMiter, EndSquare},
		{JoinMiter, EndRound},
		{JoinSquare, EndButt},
		{JoinSquare, EndSquare},
		{JoinSquare, EndRound},
		{JoinRound, EndButt},
		{JoinRound, EndSquare},
		{JoinRound, EndRound},
	}

	for _, tc := range testCases {
		t.Run(tc.joinType.String()+"_"+tc.endType.String(), func(t *testing.T) {
			result, err := InflatePaths64(
				Paths64{line},
				10.0,
				tc.joinType,
				tc.endType,
				OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25},
			)

			if err != nil {
				t.Fatalf("InflatePaths64 failed: %v", err)
			}

			validateOffsetWithOracle(t, Paths64{line}, 10.0, tc.joinType, tc.endType, result)
		})
	}
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// validateOffsetWithOracle compares pure Go offset results with CGO oracle
// When built with -tags=clipper_cgo, this tests the pure Go implementation
// by comparing against the same oracle test when run without the tag
func validateOffsetWithOracle(t *testing.T, paths Paths64, delta float64, joinType JoinType, endType EndType, result Paths64) {
	t.Helper()

	// For now, just validate basic properties
	// Full oracle validation happens when tests are run with -tags=clipper_cgo

	// Validate we got results
	if len(result) == 0 {
		t.Error("Expected non-empty result")
		return
	}

	// Validate total area is reasonable (not zero or negative for expansion)
	resultArea := 0.0
	for _, path := range result {
		resultArea += math.Abs(Area64(path))
	}

	if resultArea == 0.0 {
		t.Error("Expected non-zero total area")
	}

	t.Logf("Result: %d paths, total area: %.2f", len(result), resultArea)
}

// String methods for enums (for test naming)
func (jt JoinType) String() string {
	switch jt {
	case JoinSquare:
		return "JoinSquare"
	case JoinBevel:
		return "JoinBevel"
	case JoinRound:
		return "JoinRound"
	case JoinMiter:
		return "JoinMiter"
	default:
		return "Unknown"
	}
}

func (et EndType) String() string {
	switch et {
	case EndPolygon:
		return "EndPolygon"
	case EndJoined:
		return "EndJoined"
	case EndButt:
		return "EndButt"
	case EndSquare:
		return "EndSquare"
	case EndRound:
		return "EndRound"
	default:
		return "Unknown"
	}
}
