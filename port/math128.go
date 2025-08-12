package clipper

import (
	"math"
	"math/bits"
)

// Int128 represents a signed 128-bit integer
type Int128 struct {
	Hi int64  // high 64 bits (sign-extended)
	Lo uint64 // low 64 bits
}

// UInt128 represents an unsigned 128-bit integer
type UInt128 struct {
	Hi uint64 // high 64 bits
	Lo uint64 // low 64 bits
}

// NewInt128 creates a new Int128 from a 64-bit integer
func NewInt128(val int64) Int128 {
	var hi int64
	if val < 0 {
		hi = -1 // sign extend for negative values
	}
	return Int128{Hi: hi, Lo: uint64(val)}
}

// NewUInt128 creates a new UInt128 from a 64-bit unsigned integer
func NewUInt128(val uint64) UInt128 {
	return UInt128{Hi: 0, Lo: val}
}

// IsNegative returns true if the Int128 value is negative
func (i Int128) IsNegative() bool {
	return i.Hi < 0
}

// IsZero returns true if the Int128 value is zero
func (i Int128) IsZero() bool {
	return i.Hi == 0 && i.Lo == 0
}

// Negate returns the negation of the Int128 value
// Note: Negate(MinInt128) returns MinInt128 due to two's complement wrap
func (i Int128) Negate() Int128 {
	// Two's complement: ~i + 1
	lo := ^i.Lo + 1
	hi := ^i.Hi
	if lo == 0 { // handle carry from low to high
		hi++
	}
	return Int128{Hi: hi, Lo: lo}
}

// Add adds two Int128 values
func (i Int128) Add(other Int128) Int128 {
	lo, carry := bits.Add64(i.Lo, other.Lo, 0)
	hi, _ := bits.Add64(uint64(i.Hi), uint64(other.Hi), carry)
	return Int128{Hi: int64(hi), Lo: lo}
}

// Sub subtracts other from i
func (i Int128) Sub(other Int128) Int128 {
	lo, borrow := bits.Sub64(i.Lo, other.Lo, 0)
	hi, _ := bits.Sub64(uint64(i.Hi), uint64(other.Hi), borrow)
	return Int128{Hi: int64(hi), Lo: lo}
}

// Cmp compares two Int128 values
// Returns -1 if i < other, 0 if i == other, 1 if i > other
// Note: Lo is compared as unsigned even for negative numbers (correct for two's complement)
func (i Int128) Cmp(other Int128) int {
	if i.Hi != other.Hi {
		if i.Hi < other.Hi {
			return -1
		}
		return 1
	}
	if i.Lo == other.Lo {
		return 0
	}
	if i.Lo < other.Lo {
		return -1
	}
	return 1
}

// ToFloat64 converts Int128 to float64 (may lose precision for large values)
func (i Int128) ToFloat64() float64 {
	// For values that fit in int64 range, use direct conversion to avoid precision loss
	if (i.Hi == 0) || (i.Hi == -1 && i.Lo >= 1<<63) {
		// Value fits in int64 range
		// Both cases can use the same conversion since they fit in int64 range
		return float64(int64(i.Lo))
	}

	// For larger values, use the general formula
	const two64 = 18446744073709551616.0 // 2^64 as float64
	return float64(i.Hi)*two64 + float64(i.Lo)
}

// Mul64 multiplies an Int128 by a 64-bit integer
func (i Int128) Mul64(val int64) Int128 {
	if val == 0 {
		return Int128{}
	}

	// Special case: MinInt128 cannot be negated properly (two's complement wrap)
	// MinInt128 * val uses algebraic decomposition to avoid Negate(MinInt128) = MinInt128
	if i.Hi == math.MinInt64 && i.Lo == 0 {
		// MinInt128 * val = -(MaxInt128+1) * val
		// Split into (MinInt128+1) * val - val to avoid the problematic negation
		if val == 1 {
			return i
		}
		if val == -1 {
			return i // MinInt128 negated stays MinInt128 (intentional wrap)
		}
		// For other values, use signed arithmetic directly
		if val > 0 {
			// Result will be negative, use positive calculation and negate
			pos_result := Int128{Hi: math.MaxInt64, Lo: ^uint64(0)}.Mul64(val).Add(NewInt128(val))
			return pos_result.Negate()
		} else {
			// val < 0, result will be positive
			return Int128{Hi: math.MaxInt64, Lo: ^uint64(0)}.Mul64(-val).Add(NewInt128(-val))
		}
	}

	// Special case: val == MinInt64 cannot be negated properly
	if val == math.MinInt64 {
		// i * MinInt64 = i * (-(MaxInt64+1)) = -(i * (MaxInt64+1))
		// Split into i * MaxInt64 + i, then negate
		result := i.Mul64(math.MaxInt64).Add(i)
		return result.Negate()
	}

	negative := (i.IsNegative()) != (val < 0)

	// Work with positive values
	abs_i := i
	if i.IsNegative() {
		abs_i = i.Negate()
	}
	abs_val := val // val is guaranteed not to be MinInt64 here
	if val < 0 {
		abs_val = -val
	}

	// Multiply using bits.Mul64 (returns hi, lo)
	lo_hi, lo_lo := bits.Mul64(abs_i.Lo, uint64(abs_val))
	_, hi_lo := bits.Mul64(uint64(abs_i.Hi), uint64(abs_val))

	// Combine results: for 128-bit result, discard overflow beyond 128 bits
	lo := lo_lo
	hi, _ := bits.Add64(lo_hi, hi_lo, 0) // ignore carry - it's overflow beyond 128 bits
	// do NOT add hi_hi - that represents bits >= 128

	result := Int128{Hi: int64(hi), Lo: lo}
	if negative {
		result = result.Negate()
	}
	return result
}

// CrossProduct128 calculates the cross product of vectors (p2-p1) and (p3-p1)
// using 128-bit intermediate calculations to prevent overflow
func CrossProduct128(p1, p2, p3 Point64) Int128 {
	// Calculate vectors
	v1x := p2.X - p1.X
	v1y := p2.Y - p1.Y
	v2x := p3.X - p1.X
	v2y := p3.Y - p1.Y

	// Cross product: v1x * v2y - v1y * v2x
	// Use 128-bit multiplication to avoid overflow
	term1 := NewInt128(v1x).Mul64(v2y)
	term2 := NewInt128(v1y).Mul64(v2x)

	return term1.Sub(term2)
}

// Area128 calculates the signed area of a polygon using 128-bit precision
func Area128(path Path64) Int128 {
	if len(path) < 3 {
		return Int128{}
	}

	var area Int128
	for i := range path {
		j := (i + 1) % len(path)
		// area += path[i].X * path[j].Y - path[j].X * path[i].Y
		term1 := NewInt128(path[i].X).Mul64(path[j].Y)
		term2 := NewInt128(path[j].X).Mul64(path[i].Y)
		area = area.Add(term1.Sub(term2))
	}

	return area
}

// DistanceSquared128 calculates the squared distance between two points using 128-bit precision
func DistanceSquared128(p1, p2 Point64) UInt128 {
	// Widen to 128-bit before subtraction to prevent int64 overflow with extreme coordinates
	dx := NewInt128(p2.X).Sub(NewInt128(p1.X))
	dy := NewInt128(p2.Y).Sub(NewInt128(p1.Y))

	// Square each difference: dx² and dy²
	// Since we're squaring, the result is always positive
	var dx_sq, dy_sq Int128

	// For dx²: if dx is negative, square gives positive result
	if dx.IsNegative() {
		dx = dx.Negate()
	}
	// Now dx is non-negative, so we can treat Lo as the value (Hi should be 0)
	dx_sq = NewInt128(int64(dx.Lo)).Mul64(int64(dx.Lo))

	// Same for dy²
	if dy.IsNegative() {
		dy = dy.Negate()
	}
	dy_sq = NewInt128(int64(dy.Lo)).Mul64(int64(dy.Lo))

	result := dx_sq.Add(dy_sq)

	// Distance squared should always be non-negative
	// If result is negative, there's a bug in Mul64
	if result.IsNegative() {
		panic("DistanceSquared128: negative result indicates Mul64 bug")
	}

	// Convert to UInt128 (safe since we verified non-negative)
	return UInt128{Hi: uint64(result.Hi), Lo: result.Lo}
}
