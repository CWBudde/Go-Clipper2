package clipper

import (
	"fmt"
	"math"
	"testing"
)

// TestNewInt128 tests the Int128 constructor
func TestNewInt128(t *testing.T) {
	tests := []struct {
		name  string
		input int64
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -42},
		{"max_int64", math.MaxInt64},
		{"min_int64", math.MinInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewInt128(tt.input)
			// For negative numbers, Hi should be -1 (sign extension)
			// For positive/zero, Hi should be 0
			expectedHi := int64(0)
			if tt.input < 0 {
				expectedHi = -1
			}
			expectedLo := uint64(tt.input)

			if result.Hi != expectedHi || result.Lo != expectedLo {
				t.Errorf("NewInt128(%d) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.input, result.Hi, result.Lo, expectedHi, expectedLo)
			}
		})
	}
}

// TestNewUInt128 tests the UInt128 constructor
func TestNewUInt128(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected UInt128
	}{
		{"zero", 0, UInt128{Hi: 0, Lo: 0}},
		{"positive", 42, UInt128{Hi: 0, Lo: 42}},
		{"max_uint64", math.MaxUint64, UInt128{Hi: 0, Lo: math.MaxUint64}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewUInt128(tt.input)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("NewUInt128(%d) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.input, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestInt128_IsNegative tests the IsNegative method
func TestInt128_IsNegative(t *testing.T) {
	tests := []struct {
		name     string
		input    Int128
		expected bool
	}{
		{"zero", NewInt128(0), false},
		{"positive", NewInt128(42), false},
		{"negative", NewInt128(-42), true},
		{"max_positive", Int128{Hi: math.MaxInt64, Lo: math.MaxUint64}, false},
		{"min_negative", Int128{Hi: math.MinInt64, Lo: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.IsNegative()
			if result != tt.expected {
				t.Errorf("%s.IsNegative() = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestInt128_IsZero tests the IsZero method
func TestInt128_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		input    Int128
		expected bool
	}{
		{"zero", NewInt128(0), true},
		{"positive", NewInt128(42), false},
		{"negative", NewInt128(-42), false},
		{"large_positive", Int128{Hi: 1, Lo: 0}, false},
		{"large_negative", Int128{Hi: -1, Lo: math.MaxUint64}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.IsZero()
			if result != tt.expected {
				t.Errorf("%s.IsZero() = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestInt128_Negate tests the Negate method
func TestInt128_Negate(t *testing.T) {
	tests := []struct {
		name     string
		input    Int128
		expected Int128
	}{
		{"zero", NewInt128(0), NewInt128(0)},
		{"positive", NewInt128(42), NewInt128(-42)},
		{"negative", NewInt128(-42), NewInt128(42)},
		{"one", NewInt128(1), NewInt128(-1)},
		{"minus_one", NewInt128(-1), NewInt128(1)},
		// MinInt128 negation wraps due to two's complement
		{"min_int128", Int128{Hi: math.MinInt64, Lo: 0}, Int128{Hi: math.MinInt64, Lo: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Negate()
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s.Negate() = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestInt128_Add tests the Add method
func TestInt128_Add(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Int128
		expected Int128
	}{
		{"zero_plus_zero", NewInt128(0), NewInt128(0), NewInt128(0)},
		{"positive_plus_positive", NewInt128(10), NewInt128(32), NewInt128(42)},
		{"negative_plus_negative", NewInt128(-10), NewInt128(-32), NewInt128(-42)},
		{"positive_plus_negative", NewInt128(50), NewInt128(-8), NewInt128(42)},
		{"negative_plus_positive", NewInt128(-50), NewInt128(92), NewInt128(42)},
		{"carry_from_low", Int128{Hi: 0, Lo: math.MaxUint64}, NewInt128(1), Int128{Hi: 1, Lo: 0}},
		{"carry_with_negative", Int128{Hi: -1, Lo: math.MaxUint64}, NewInt128(1), NewInt128(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Add(tt.b)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: %v + %v = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.a, tt.b, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestInt128_Sub tests the Sub method
func TestInt128_Sub(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Int128
		expected Int128
	}{
		{"zero_minus_zero", NewInt128(0), NewInt128(0), NewInt128(0)},
		{"positive_minus_positive", NewInt128(50), NewInt128(8), NewInt128(42)},
		{"negative_minus_negative", NewInt128(-8), NewInt128(-50), NewInt128(42)},
		{"positive_minus_negative", NewInt128(10), NewInt128(-32), NewInt128(42)},
		{"negative_minus_positive", NewInt128(-10), NewInt128(32), NewInt128(-42)},
		{"borrow_from_high", Int128{Hi: 1, Lo: 0}, NewInt128(1), Int128{Hi: 0, Lo: math.MaxUint64}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Sub(tt.b)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: %v - %v = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.a, tt.b, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestInt128_Cmp tests the Cmp method
func TestInt128_Cmp(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Int128
		expected int
	}{
		{"equal_zero", NewInt128(0), NewInt128(0), 0},
		{"equal_positive", NewInt128(42), NewInt128(42), 0},
		{"equal_negative", NewInt128(-42), NewInt128(-42), 0},
		{"positive_greater", NewInt128(50), NewInt128(42), 1},
		{"positive_less", NewInt128(42), NewInt128(50), -1},
		{"negative_greater", NewInt128(-42), NewInt128(-50), 1},
		{"negative_less", NewInt128(-50), NewInt128(-42), -1},
		{"positive_vs_negative", NewInt128(1), NewInt128(-1), 1},
		{"negative_vs_positive", NewInt128(-1), NewInt128(1), -1},
		{"large_positive_vs_small_positive", Int128{Hi: 1, Lo: 0}, NewInt128(1000000), 1},
		{"large_negative_vs_small_negative", Int128{Hi: -2, Lo: 0}, NewInt128(-1000000), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Cmp(tt.b)
			if result != tt.expected {
				t.Errorf("%s: %v.Cmp(%v) = %d, expected %d", tt.name, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestInt128_ToFloat64 tests the ToFloat64 method
func TestInt128_ToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    Int128
		expected float64
	}{
		{"zero", NewInt128(0), 0.0},
		{"positive", NewInt128(42), 42.0},
		{"negative", NewInt128(-42), -42.0},
		{"large_positive", Int128{Hi: 1, Lo: 0}, math.Pow(2, 64)},
		{"large_negative", Int128{Hi: -1, Lo: 0}, -math.Pow(2, 64)},
		{"max_int64", NewInt128(math.MaxInt64), float64(math.MaxInt64)},
		{"min_int64", NewInt128(math.MinInt64), float64(math.MinInt64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToFloat64()
			if result != tt.expected {
				t.Errorf("%s: %v.ToFloat64() = %g, expected %g", tt.name, tt.input, result, tt.expected)
			}
		})
	}
}

// TestInt128_Mul64 tests the Mul64 method
func TestInt128_Mul64(t *testing.T) {
	tests := []struct {
		name     string
		input    Int128
		val      int64
		expected Int128
	}{
		{"zero_times_anything", NewInt128(0), 42, NewInt128(0)},
		{"anything_times_zero", NewInt128(42), 0, NewInt128(0)},
		{"positive_times_positive", NewInt128(6), 7, NewInt128(42)},
		{"positive_times_negative", NewInt128(6), -7, NewInt128(-42)},
		{"negative_times_positive", NewInt128(-6), 7, NewInt128(-42)},
		{"negative_times_negative", NewInt128(-6), -7, NewInt128(42)},
		{"one_times_anything", NewInt128(1), 42, NewInt128(42)},
		{"anything_times_one", NewInt128(42), 1, NewInt128(42)},
		{"minus_one_times_anything", NewInt128(-1), 42, NewInt128(-42)},
		{"anything_times_minus_one", NewInt128(42), -1, NewInt128(-42)},
		// Special case: MinInt128
		{"min_int128_times_one", Int128{Hi: math.MinInt64, Lo: 0}, 1, Int128{Hi: math.MinInt64, Lo: 0}},
		{"min_int128_times_minus_one", Int128{Hi: math.MinInt64, Lo: 0}, -1, Int128{Hi: math.MinInt64, Lo: 0}},
		// Special case: multiplying by MinInt64
		{"one_times_min_int64", NewInt128(1), math.MinInt64, NewInt128(math.MinInt64)},
		{"two_times_min_int64", NewInt128(2), math.MinInt64, Int128{Hi: -1, Lo: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Mul64(tt.val)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: %v.Mul64(%d) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.input, tt.val, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestMath128_CrossProduct128 tests the CrossProduct128 function
func TestMath128_CrossProduct128(t *testing.T) {
	tests := []struct {
		name       string
		p1, p2, p3 Point64
		expected   Int128
	}{
		{
			name:     "collinear_points",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 1, Y: 1},
			p3:       Point64{X: 2, Y: 2},
			expected: NewInt128(0), // (1*2 - 1*2) = 0
		},
		{
			name:     "counter_clockwise",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 1, Y: 0},
			p3:       Point64{X: 0, Y: 1},
			expected: NewInt128(1), // (1*1 - 0*0) = 1
		},
		{
			name:     "clockwise",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 0, Y: 1},
			p3:       Point64{X: 1, Y: 0},
			expected: NewInt128(-1), // (0*0 - 1*1) = -1
		},
		{
			name:     "large_coordinates",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 1000000, Y: 0},
			p3:       Point64{X: 0, Y: 1000000},
			expected: NewInt128(1000000).Mul64(1000000), // 1000000 * 1000000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CrossProduct128(tt.p1, tt.p2, tt.p3)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: CrossProduct128(%v, %v, %v) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.p1, tt.p2, tt.p3, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestMath128_Area128 tests the Area128 function
func TestMath128_Area128(t *testing.T) {
	tests := []struct {
		name     string
		path     Path64
		expected Int128
	}{
		{
			name:     "empty_path",
			path:     Path64{},
			expected: NewInt128(0),
		},
		{
			name:     "single_point",
			path:     Path64{{X: 0, Y: 0}},
			expected: NewInt128(0),
		},
		{
			name:     "two_points",
			path:     Path64{{X: 0, Y: 0}, {X: 1, Y: 1}},
			expected: NewInt128(0),
		},
		{
			name: "unit_square_ccw",
			path: Path64{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0, Y: 1},
			},
			expected: NewInt128(2), // Area formula gives 2 * actual area
		},
		{
			name: "unit_square_cw",
			path: Path64{
				{X: 0, Y: 0},
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 1, Y: 0},
			},
			expected: NewInt128(-2), // Clockwise gives negative area
		},
		{
			name: "triangle",
			path: Path64{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 1, Y: 2},
			},
			expected: NewInt128(4), // Triangle with base=2, height=2, area=2, formula gives 2*2=4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Area128(tt.path)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: Area128(%v) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.path, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestDistanceSquared128 tests the DistanceSquared128 function
func TestDistanceSquared128(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Point64
		expected UInt128
	}{
		{
			name:     "same_point",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 0, Y: 0},
			expected: UInt128{Hi: 0, Lo: 0},
		},
		{
			name:     "unit_distance_x",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 1, Y: 0},
			expected: UInt128{Hi: 0, Lo: 1},
		},
		{
			name:     "unit_distance_y",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 0, Y: 1},
			expected: UInt128{Hi: 0, Lo: 1},
		},
		{
			name:     "pythagorean_3_4_5",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 3, Y: 4},
			expected: UInt128{Hi: 0, Lo: 25}, // 3² + 4² = 9 + 16 = 25
		},
		{
			name:     "negative_coordinates",
			p1:       Point64{X: -1, Y: -1},
			p2:       Point64{X: 1, Y: 1},
			expected: UInt128{Hi: 0, Lo: 8}, // (-2)² + (-2)² = 4 + 4 = 8
		},
		{
			name:     "large_coordinates",
			p1:       Point64{X: 0, Y: 0},
			p2:       Point64{X: 1000000, Y: 1000000},
			expected: UInt128{Hi: 0, Lo: 2000000000000}, // 2 * 10^12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DistanceSquared128(tt.p1, tt.p2)
			if result.Hi != tt.expected.Hi || result.Lo != tt.expected.Lo {
				t.Errorf("%s: DistanceSquared128(%v, %v) = {Hi: %d, Lo: %d}, expected {Hi: %d, Lo: %d}",
					tt.name, tt.p1, tt.p2, result.Hi, result.Lo, tt.expected.Hi, tt.expected.Lo)
			}
		})
	}
}

// TestInt128_AddSubInverse tests that Add and Sub are inverse operations
func TestInt128_AddSubInverse(t *testing.T) {
	values := []Int128{
		NewInt128(0),
		NewInt128(42),
		NewInt128(-42),
		NewInt128(math.MaxInt64),
		NewInt128(math.MinInt64),
		Int128{Hi: 123, Lo: 456},
		Int128{Hi: -123, Lo: 456},
	}

	for i, a := range values {
		for j, b := range values {
			name := fmt.Sprintf("val_%d_add_sub_val_%d", i, j)
			t.Run(name, func(t *testing.T) {
				sum := a.Add(b)
				diff := sum.Sub(b)
				if diff.Hi != a.Hi || diff.Lo != a.Lo {
					t.Errorf("values[%d] + values[%d] - values[%d] != values[%d]: %v + %v - %v = %v, expected %v",
						i, j, j, i, a, b, b, diff, a)
				}
			})
		}
	}
}

// TestInt128_NegateInverse tests that double negation returns original value
func TestInt128_NegateInverse(t *testing.T) {
	values := []Int128{
		NewInt128(0),
		NewInt128(42),
		NewInt128(-42),
		NewInt128(math.MaxInt64),
		Int128{Hi: 123, Lo: 456},
		Int128{Hi: -123, Lo: 456},
		// Note: MinInt128 is excluded because it's the only value where double negation doesn't work
	}

	for i, val := range values {
		name := fmt.Sprintf("val_%d_double_negate", i)
		t.Run(name, func(t *testing.T) {
			result := val.Negate().Negate()
			if result.Hi != val.Hi || result.Lo != val.Lo {
				t.Errorf("values[%d].Negate().Negate() != values[%d]: %v.Negate().Negate() = %v, expected %v",
					i, i, val, result, val)
			}
		})
	}
}

// Benchmark tests for performance measurement
func BenchmarkInt128_Add(b *testing.B) {
	a := NewInt128(123456789)
	c := NewInt128(987654321)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Add(c)
	}
}

func BenchmarkInt128_Mul64(b *testing.B) {
	a := NewInt128(123456789)
	val := int64(987654321)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Mul64(val)
	}
}

func BenchmarkCrossProduct128(b *testing.B) {
	p1 := Point64{X: 123, Y: 456}
	p2 := Point64{X: 789, Y: 0o12}
	p3 := Point64{X: 345, Y: 678}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CrossProduct128(p1, p2, p3)
	}
}
