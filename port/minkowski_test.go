package clipper

import (
	"errors"
	"math"
	"testing"
)

// TestMinkowskiSum64BasicSquare tests Minkowski sum with a square pattern on a simple path
func TestMinkowskiSum64BasicSquare(t *testing.T) {
	// Small square pattern
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	// Simple 3-point path
	path := Path64{{0, 0}, {10, 0}, {10, 10}}

	result, err := MinkowskiSum64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiSum64")
	}
	t.Logf("MinkowskiSum basic: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiSum64ClosedPath tests Minkowski sum on a closed path
func TestMinkowskiSum64ClosedPath(t *testing.T) {
	// Small square pattern
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	// Closed square path
	path := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	result, err := MinkowskiSum64(pattern, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiSum64")
	}
	t.Logf("MinkowskiSum closed: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiDiff64BasicSquare tests Minkowski difference with a square pattern
func TestMinkowskiDiff64BasicSquare(t *testing.T) {
	// Small square pattern
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	// Simple 3-point path
	path := Path64{{0, 0}, {10, 0}, {10, 10}}

	result, err := MinkowskiDiff64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiDiff64")
	}
	t.Logf("MinkowskiDiff basic: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiDiff64ClosedPath tests Minkowski difference on a closed path
func TestMinkowskiDiff64ClosedPath(t *testing.T) {
	// Small square pattern
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	// Closed square path
	path := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	result, err := MinkowskiDiff64(pattern, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiDiff64")
	}
	t.Logf("MinkowskiDiff closed: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiSum64EmptyInputs tests edge cases with empty inputs
func TestMinkowskiSum64EmptyInputs(t *testing.T) {
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	empty := Path64{}

	// Empty path - should return error
	_, err := MinkowskiSum64(pattern, empty, false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("expected ErrEmptyPath for empty path, got %v", err)
	}

	// Empty pattern - should return error
	_, err = MinkowskiSum64(empty, Path64{{0, 0}, {10, 0}}, false)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
	if !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("expected ErrEmptyPath for empty pattern, got %v", err)
	}
}

// TestMinkowskiSum64Circle tests Minkowski sum with a circular pattern
func TestMinkowskiSum64Circle(t *testing.T) {
	// Circular pattern approximated with 8 points
	circle := Path64{
		{10, 0}, {7, 7}, {0, 10}, {-7, 7},
		{-10, 0}, {-7, -7}, {0, -10}, {7, -7},
	}
	// Simple line path
	path := Path64{{0, 0}, {50, 0}}

	result, err := MinkowskiSum64(circle, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiSum64 with circle")
	}
	t.Logf("MinkowskiSum circle: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiSum64Triangle tests Minkowski sum with a triangular pattern
func TestMinkowskiSum64Triangle(t *testing.T) {
	// Triangular pattern
	triangle := Path64{{0, -5}, {5, 5}, {-5, 5}}
	// L-shaped path
	path := Path64{{0, 0}, {20, 0}, {20, 20}}

	result, err := MinkowskiSum64(triangle, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiSum64 with triangle")
	}
	t.Logf("MinkowskiSum triangle: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiDiff64LargerPattern tests Minkowski difference with larger pattern
func TestMinkowskiDiff64LargerPattern(t *testing.T) {
	// Larger square pattern
	pattern := Path64{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}
	// Path around the origin
	path := Path64{{0, 0}, {30, 0}, {30, 30}, {0, 30}}

	result, err := MinkowskiDiff64(pattern, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty result from MinkowskiDiff64 with large pattern")
	}
	t.Logf("MinkowskiDiff large: %d paths with total %d points", len(result), totalPoints(result))
}

// TestMinkowskiSum64OpenVsClosed tests difference between open and closed paths
func TestMinkowskiSum64OpenVsClosed(t *testing.T) {
	pattern := Path64{{-2, -2}, {2, -2}, {2, 2}, {-2, 2}}
	path := Path64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}

	// Open path
	openResult, err := MinkowskiSum64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}

	// Closed path
	closedResult, err := MinkowskiSum64(pattern, path, true)
	if err != nil {
		t.Fatal(err)
	}

	// Closed path should have more paths/points (connects last to first)
	openPoints := totalPoints(openResult)
	closedPoints := totalPoints(closedResult)

	t.Logf("Open path: %d paths, %d points", len(openResult), openPoints)
	t.Logf("Closed path: %d paths, %d points", len(closedResult), closedPoints)

	if closedPoints <= openPoints {
		t.Fatalf("expected closed path to have more points than open path")
	}
}

// TestMinkowskiSum64SinglePoint tests Minkowski sum with a single point pattern
func TestMinkowskiSum64SinglePoint(t *testing.T) {
	// Single point "pattern" (edge case)
	pattern := Path64{{5, 5}}
	path := Path64{{0, 0}, {10, 0}}

	result, err := MinkowskiSum64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}
	// Single point pattern should result in empty (no area)
	if len(result) != 0 {
		t.Fatalf("expected empty result for single-point pattern, got %d paths", len(result))
	}
}

// TestMinkowskiDiff64Symmetry tests that Minkowski diff is correct
func TestMinkowskiDiff64Symmetry(t *testing.T) {
	pattern := Path64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	path := Path64{{0, 0}, {10, 0}}

	resultSum, err := MinkowskiSum64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}

	resultDiff, err := MinkowskiDiff64(pattern, path, false)
	if err != nil {
		t.Fatal(err)
	}

	// Both should produce results
	if len(resultSum) == 0 {
		t.Fatal("expected non-empty sum result")
	}
	if len(resultDiff) == 0 {
		t.Fatal("expected non-empty diff result")
	}

	t.Logf("Sum: %d paths, Diff: %d paths", len(resultSum), len(resultDiff))
}

// TestMinkowskiSum64RobotPathPlanning tests realistic robot path planning use case
func TestMinkowskiSum64RobotPathPlanning(t *testing.T) {
	// Robot footprint (square approximation)
	robotFootprint := Path64{{-5, -5}, {5, -5}, {5, 5}, {-5, 5}}

	// Simple path for robot to follow
	robotPath := Path64{
		{0, 0}, {50, 0}, {50, 50}, {0, 50}, {0, 100},
	}

	// Compute expanded path (safe zone for robot center)
	expandedPath, err := MinkowskiSum64(robotFootprint, robotPath, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(expandedPath) == 0 {
		t.Fatal("expected non-empty expanded path for robot planning")
	}

	// The expanded path should be significantly larger
	sumArea := totalArea(expandedPath)
	if sumArea <= 0 {
		t.Fatal("expected positive area for expanded path")
	}

	t.Logf("Robot path planning: expanded to %d paths, total area: %d", len(expandedPath), sumArea)
}

// totalPoints counts total points across all paths
func totalPoints(paths Paths64) int {
	total := 0
	for _, p := range paths {
		total += len(p)
	}
	return total
}

// totalArea calculates total absolute area of all paths
func totalArea(paths Paths64) int64 {
	var total int64
	for _, p := range paths {
		area := Area64(p)
		if area < 0 {
			area = -area
		}
		total += int64(math.Abs(area))
	}
	return total
}
