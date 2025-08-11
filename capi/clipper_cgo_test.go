//go:build clipper_cgo

package capi

import "testing"

func TestUnionTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64(/*Union*/ 1, /*NonZero*/ 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected merged polygon, got empty")
	}
	t.Logf("Union result: %v", got)
}

func TestIntersectionTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64(/*Intersection*/ 0, /*NonZero*/ 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected intersection polygon, got empty")
	}
	t.Logf("Intersection result: %v", got)
}

func TestDifferenceTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64(/*Difference*/ 2, /*NonZero*/ 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected difference polygon, got empty")
	}
	t.Logf("Difference result: %v", got)
}

func TestXorTiny(t *testing.T) {
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	b := Paths64{{{5, 5}, {15, 5}, {15, 15}, {5, 15}}}
	got, open, err := BooleanOp64(/*Xor*/ 3, /*NonZero*/ 1, a, nil, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) == 0 {
		t.Fatalf("expected xor polygon, got empty")
	}
	t.Logf("Xor result: %v", got)
}

func TestEmptyInputs(t *testing.T) {
	empty := Paths64{}
	a := Paths64{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}}
	
	got, open, err := BooleanOp64(/*Union*/ 1, /*NonZero*/ 1, a, nil, empty)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 {
		t.Fatalf("unexpected open paths")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 path for union with empty, got %d", len(got))
	}
	t.Logf("Union with empty result: %v", got)
}