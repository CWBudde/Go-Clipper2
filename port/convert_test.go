package clipper

import (
	"math"
	"testing"
)

// ==============================================================================
// Point Conversion Tests
// ==============================================================================

func TestPoint64ToPoint32_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    Point64
		expected Point32
	}{
		{"Zero", Point64{0, 0}, Point32{0, 0}},
		{"Positive", Point64{100, 200}, Point32{100, 200}},
		{"Negative", Point64{-100, -200}, Point32{-100, -200}},
		{"MaxInt32", Point64{MaxInt32, MaxInt32}, Point32{math.MaxInt32, math.MaxInt32}},
		{"MinInt32", Point64{MinInt32, MinInt32}, Point32{math.MinInt32, math.MinInt32}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Point64ToPoint32(tt.input)
			if err != nil {
				t.Errorf("Point64ToPoint32(%v) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("Point64ToPoint32(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPoint64ToPoint32_Overflow(t *testing.T) {
	tests := []struct {
		name  string
		input Point64
	}{
		{"X overflow positive", Point64{MaxInt32 + 1, 0}},
		{"X overflow negative", Point64{MinInt32 - 1, 0}},
		{"Y overflow positive", Point64{0, MaxInt32 + 1}},
		{"Y overflow negative", Point64{0, MinInt32 - 1}},
		{"Both overflow", Point64{MaxInt32 + 1, MaxInt32 + 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Point64ToPoint32(tt.input)
			if err != ErrInt32Overflow {
				t.Errorf("Point64ToPoint32(%v) expected ErrInt32Overflow, got %v", tt.input, err)
			}
		})
	}
}

func TestPoint32ToPoint64(t *testing.T) {
	tests := []struct {
		name     string
		input    Point32
		expected Point64
	}{
		{"Zero", Point32{0, 0}, Point64{0, 0}},
		{"Positive", Point32{100, 200}, Point64{100, 200}},
		{"Negative", Point32{-100, -200}, Point64{-100, -200}},
		{"MaxInt32", Point32{math.MaxInt32, math.MaxInt32}, Point64{MaxInt32, MaxInt32}},
		{"MinInt32", Point32{math.MinInt32, math.MinInt32}, Point64{MinInt32, MinInt32}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Point32ToPoint64(tt.input)
			if result != tt.expected {
				t.Errorf("Point32ToPoint64(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ==============================================================================
// Path Conversion Tests
// ==============================================================================

func TestPath64ToPath32_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    Path64
		expected Path32
	}{
		{"Nil path", nil, nil},
		{"Empty path", Path64{}, Path32{}},
		{"Single point", Path64{{10, 20}}, Path32{{10, 20}}},
		{"Triangle", Path64{{0, 0}, {100, 0}, {50, 100}}, Path32{{0, 0}, {100, 0}, {50, 100}}},
		{"At boundaries", Path64{{MaxInt32, MinInt32}, {MinInt32, MaxInt32}}, Path32{{math.MaxInt32, math.MinInt32}, {math.MinInt32, math.MaxInt32}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Path64ToPath32(tt.input)
			if err != nil {
				t.Errorf("Path64ToPath32(%v) unexpected error: %v", tt.input, err)
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Path64ToPath32(%v) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Path64ToPath32(%v)[%d] = %v, want %v", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestPath64ToPath32_Overflow(t *testing.T) {
	tests := []struct {
		name  string
		input Path64
	}{
		{"First point overflows", Path64{{MaxInt32 + 1, 0}, {0, 0}}},
		{"Middle point overflows", Path64{{0, 0}, {MinInt32 - 1, 0}, {0, 0}}},
		{"Last point overflows", Path64{{0, 0}, {0, MaxInt32 + 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Path64ToPath32(tt.input)
			if err != ErrInt32Overflow {
				t.Errorf("Path64ToPath32(%v) expected ErrInt32Overflow, got %v", tt.input, err)
			}
		})
	}
}

func TestPath32ToPath64(t *testing.T) {
	tests := []struct {
		name     string
		input    Path32
		expected Path64
	}{
		{"Nil path", nil, nil},
		{"Empty path", Path32{}, Path64{}},
		{"Single point", Path32{{10, 20}}, Path64{{10, 20}}},
		{"Rectangle", Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}, Path64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Path32ToPath64(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Path32ToPath64(%v) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Path32ToPath64(%v)[%d] = %v, want %v", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// ==============================================================================
// Paths Conversion Tests
// ==============================================================================

func TestPaths64ToPaths32_Success(t *testing.T) {
	input := Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}
	expected := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	result, err := Paths64ToPaths32(input)
	if err != nil {
		t.Errorf("Paths64ToPaths32() unexpected error: %v", err)
	}
	if len(result) != len(expected) {
		t.Errorf("Paths64ToPaths32() length = %d, want %d", len(result), len(expected))
		return
	}

	for i := range result {
		if len(result[i]) != len(expected[i]) {
			t.Errorf("Paths64ToPaths32()[%d] length = %d, want %d", i, len(result[i]), len(expected[i]))
			continue
		}
		for j := range result[i] {
			if result[i][j] != expected[i][j] {
				t.Errorf("Paths64ToPaths32()[%d][%d] = %v, want %v", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestPaths64ToPaths32_Overflow(t *testing.T) {
	input := Paths64{
		{{0, 0}, {100, 0}, {100, 100}},
		{{MaxInt32 + 1, 0}, {0, 0}}, // Second path has overflow
	}

	_, err := Paths64ToPaths32(input)
	if err != ErrInt32Overflow {
		t.Errorf("Paths64ToPaths32() expected ErrInt32Overflow, got %v", err)
	}
}

func TestPaths32ToPaths64(t *testing.T) {
	input := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}
	expected := Paths64{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	result := Paths32ToPaths64(input)
	if len(result) != len(expected) {
		t.Errorf("Paths32ToPaths64() length = %d, want %d", len(result), len(expected))
		return
	}

	for i := range result {
		if len(result[i]) != len(expected[i]) {
			t.Errorf("Paths32ToPaths64()[%d] length = %d, want %d", i, len(result[i]), len(expected[i]))
			continue
		}
		for j := range result[i] {
			if result[i][j] != expected[i][j] {
				t.Errorf("Paths32ToPaths64()[%d][%d] = %v, want %v", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

// ==============================================================================
// Rectangle Conversion Tests
// ==============================================================================

func TestRect64ToRect32_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    Rect64
		expected Rect32
	}{
		{"Zero rect", Rect64{0, 0, 0, 0}, Rect32{0, 0, 0, 0}},
		{"Positive rect", Rect64{10, 20, 100, 200}, Rect32{10, 20, 100, 200}},
		{"Negative rect", Rect64{-100, -200, -10, -20}, Rect32{-100, -200, -10, -20}},
		{"At boundaries", Rect64{MinInt32, MinInt32, MaxInt32, MaxInt32}, Rect32{math.MinInt32, math.MinInt32, math.MaxInt32, math.MaxInt32}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Rect64ToRect32(tt.input)
			if err != nil {
				t.Errorf("Rect64ToRect32(%v) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("Rect64ToRect32(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRect64ToRect32_Overflow(t *testing.T) {
	tests := []struct {
		name  string
		input Rect64
	}{
		{"Left overflow", Rect64{MinInt32 - 1, 0, 0, 0}},
		{"Top overflow", Rect64{0, MinInt32 - 1, 0, 0}},
		{"Right overflow", Rect64{0, 0, MaxInt32 + 1, 0}},
		{"Bottom overflow", Rect64{0, 0, 0, MaxInt32 + 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Rect64ToRect32(tt.input)
			if err != ErrInt32Overflow {
				t.Errorf("Rect64ToRect32(%v) expected ErrInt32Overflow, got %v", tt.input, err)
			}
		})
	}
}

func TestRect32ToRect64(t *testing.T) {
	tests := []struct {
		name     string
		input    Rect32
		expected Rect64
	}{
		{"Zero rect", Rect32{0, 0, 0, 0}, Rect64{0, 0, 0, 0}},
		{"Positive rect", Rect32{10, 20, 100, 200}, Rect64{10, 20, 100, 200}},
		{"At boundaries", Rect32{math.MinInt32, math.MinInt32, math.MaxInt32, math.MaxInt32}, Rect64{MinInt32, MinInt32, MaxInt32, MaxInt32}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rect32ToRect64(tt.input)
			if result != tt.expected {
				t.Errorf("Rect32ToRect64(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ==============================================================================
// Validation Tests
// ==============================================================================

func TestValidateInt32Range(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		wantErr bool
	}{
		{"Zero", 0, false},
		{"Positive in range", 1000, false},
		{"Negative in range", -1000, false},
		{"MaxInt32", MaxInt32, false},
		{"MinInt32", MinInt32, false},
		{"MaxInt32 + 1", MaxInt32 + 1, true},
		{"MinInt32 - 1", MinInt32 - 1, true},
		{"Large positive", math.MaxInt64, true},
		{"Large negative", math.MinInt64, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInt32Range(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInt32Range(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr && err != ErrInt32Overflow {
				t.Errorf("ValidateInt32Range(%d) expected ErrInt32Overflow, got %v", tt.input, err)
			}
		})
	}
}

// ==============================================================================
// Roundtrip Tests
// ==============================================================================

func TestPointRoundtrip32To64To32(t *testing.T) {
	tests := []Point32{
		{0, 0},
		{100, 200},
		{-100, -200},
		{math.MaxInt32, math.MaxInt32},
		{math.MinInt32, math.MinInt32},
	}

	for _, original := range tests {
		t.Run("Roundtrip", func(t *testing.T) {
			pt64 := Point32ToPoint64(original)
			result, err := Point64ToPoint32(pt64)
			if err != nil {
				t.Errorf("Roundtrip failed with error: %v", err)
			}
			if result != original {
				t.Errorf("Roundtrip: got %v, want %v", result, original)
			}
		})
	}
}

func TestPathRoundtrip32To64To32(t *testing.T) {
	original := Path32{{0, 0}, {100, 0}, {100, 100}, {0, 100}}

	path64 := Path32ToPath64(original)
	result, err := Path64ToPath32(path64)
	if err != nil {
		t.Fatalf("Roundtrip failed with error: %v", err)
	}

	if len(result) != len(original) {
		t.Fatalf("Roundtrip length mismatch: got %d, want %d", len(result), len(original))
	}

	for i := range result {
		if result[i] != original[i] {
			t.Errorf("Roundtrip[%d]: got %v, want %v", i, result[i], original[i])
		}
	}
}

func TestPathsRoundtrip32To64To32(t *testing.T) {
	original := Paths32{
		{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		{{25, 25}, {75, 25}, {75, 75}, {25, 75}},
	}

	paths64 := Paths32ToPaths64(original)
	result, err := Paths64ToPaths32(paths64)
	if err != nil {
		t.Fatalf("Roundtrip failed with error: %v", err)
	}

	if len(result) != len(original) {
		t.Fatalf("Roundtrip length mismatch: got %d, want %d", len(result), len(original))
	}

	for i := range result {
		if len(result[i]) != len(original[i]) {
			t.Fatalf("Roundtrip[%d] length mismatch: got %d, want %d", i, len(result[i]), len(original[i]))
		}
		for j := range result[i] {
			if result[i][j] != original[i][j] {
				t.Errorf("Roundtrip[%d][%d]: got %v, want %v", i, j, result[i][j], original[i][j])
			}
		}
	}
}

func TestRectRoundtrip32To64To32(t *testing.T) {
	original := Rect32{10, 20, 100, 200}

	rect64 := Rect32ToRect64(original)
	result, err := Rect64ToRect32(rect64)
	if err != nil {
		t.Fatalf("Roundtrip failed with error: %v", err)
	}

	if result != original {
		t.Errorf("Roundtrip: got %v, want %v", result, original)
	}
}
