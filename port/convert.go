package clipper

import (
	"math"
)

// Conversion utilities between 32-bit and 64-bit coordinate types
// These functions provide safe conversions with overflow detection to ensure
// 32-bit operations maintain correctness when working with int32 coordinate ranges.

// Constants for int32 range validation
const (
	MaxInt32 = int64(math.MaxInt32) //  2147483647
	MinInt32 = int64(math.MinInt32) // -2147483648
)

// ==============================================================================
// 64-bit to 32-bit Conversions (with overflow detection)
// ==============================================================================

// ValidateInt32Range checks if a 64-bit value fits in int32 range
// Returns ErrInt32Overflow if the value exceeds int32 limits
func ValidateInt32Range(val int64) error {
	if val > MaxInt32 || val < MinInt32 {
		return ErrInt32Overflow
	}
	return nil
}

// Point64ToPoint32 converts a 64-bit point to 32-bit with overflow detection
// Returns error if either coordinate exceeds int32 range
func Point64ToPoint32(pt Point64) (Point32, error) {
	if err := ValidateInt32Range(pt.X); err != nil {
		return Point32{}, err
	}
	if err := ValidateInt32Range(pt.Y); err != nil {
		return Point32{}, err
	}
	return Point32{
		X: int32(pt.X),
		Y: int32(pt.Y),
	}, nil
}

// Path64ToPath32 converts a 64-bit path to 32-bit with overflow detection
// Returns error if any coordinate in any point exceeds int32 range
func Path64ToPath32(path Path64) (Path32, error) {
	if path == nil {
		return nil, nil
	}

	result := make(Path32, len(path))
	for i, pt := range path {
		converted, err := Point64ToPoint32(pt)
		if err != nil {
			return nil, err
		}
		result[i] = converted
	}
	return result, nil
}

// Paths64ToPaths32 converts multiple 64-bit paths to 32-bit with overflow detection
// Returns error if any coordinate in any path exceeds int32 range
func Paths64ToPaths32(paths Paths64) (Paths32, error) {
	if paths == nil {
		return nil, nil
	}

	result := make(Paths32, len(paths))
	for i, path := range paths {
		converted, err := Path64ToPath32(path)
		if err != nil {
			return nil, err
		}
		result[i] = converted
	}
	return result, nil
}

// Rect64ToRect32 converts a 64-bit rectangle to 32-bit with overflow detection
// Returns error if any coordinate exceeds int32 range
func Rect64ToRect32(rect Rect64) (Rect32, error) {
	if err := ValidateInt32Range(rect.Left); err != nil {
		return Rect32{}, err
	}
	if err := ValidateInt32Range(rect.Top); err != nil {
		return Rect32{}, err
	}
	if err := ValidateInt32Range(rect.Right); err != nil {
		return Rect32{}, err
	}
	if err := ValidateInt32Range(rect.Bottom); err != nil {
		return Rect32{}, err
	}
	return Rect32{
		Left:   int32(rect.Left),
		Top:    int32(rect.Top),
		Right:  int32(rect.Right),
		Bottom: int32(rect.Bottom),
	}, nil
}

// ==============================================================================
// 32-bit to 64-bit Conversions (always safe - promotion)
// ==============================================================================

// Point32ToPoint64 converts a 32-bit point to 64-bit (always safe)
// No overflow is possible since int32 range fits entirely within int64
func Point32ToPoint64(pt Point32) Point64 {
	return Point64{
		X: int64(pt.X),
		Y: int64(pt.Y),
	}
}

// Path32ToPath64 converts a 32-bit path to 64-bit (always safe)
func Path32ToPath64(path Path32) Path64 {
	if path == nil {
		return nil
	}

	result := make(Path64, len(path))
	for i, pt := range path {
		result[i] = Point32ToPoint64(pt)
	}
	return result
}

// Paths32ToPaths64 converts multiple 32-bit paths to 64-bit (always safe)
func Paths32ToPaths64(paths Paths32) Paths64 {
	if paths == nil {
		return nil
	}

	result := make(Paths64, len(paths))
	for i, path := range paths {
		result[i] = Path32ToPath64(path)
	}
	return result
}

// Rect32ToRect64 converts a 32-bit rectangle to 64-bit (always safe)
func Rect32ToRect64(rect Rect32) Rect64 {
	return Rect64{
		Left:   int64(rect.Left),
		Top:    int64(rect.Top),
		Right:  int64(rect.Right),
		Bottom: int64(rect.Bottom),
	}
}
