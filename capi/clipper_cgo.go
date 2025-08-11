//go:build clipper_cgo

package capi

/*
#cgo CXXFLAGS: -std=c++17
#cgo pkg-config: Clipper2

#include <stdint.h>
#include <stdlib.h>
#include "clipper2/clipper.export.h"

// helper wrapper for RectClip64 to avoid dealing with CRect64 from Go
CPaths64 RectClip64Wrapper(int64_t left, int64_t top, int64_t right, int64_t bottom, const CPaths64 paths) {
    CRect64 r;
    r.left = left;
    r.top = top;
    r.right = right;
    r.bottom = bottom;
    return RectClip64(r, paths);
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Point64 struct{ X, Y int64 }
type Path64 []Point64
type Paths64 []Path64

// packPaths64 allocates the CPaths64 array layout expected by Clipper.Export.
// Layout: [A, C,  (N,0, x1,y1, ...), (N,0, ...), ...]
func packPaths64(ps Paths64) (*C.int64_t, int, func()) {
	// compute total length A and count C
	A := 2 // header
	for _, p := range ps {
		if len(p) == 0 {
			continue
		}
		A += 2 + 2*len(p) // (N,0) + x,y pairs
	}
	Cnt := 0
	for _, p := range ps {
		if len(p) > 0 {
			Cnt++
		}
	}

	if Cnt == 0 {
		// still return a minimal header so C++ side can handle empty input
		mem := C.malloc(C.size_t(2 * C.size_t(unsafe.Sizeof(C.int64_t(0)))))
		arr := (*C.int64_t)(mem)
		hdr := (*[1 << 30]C.int64_t)(unsafe.Pointer(arr))[:2:2]
		hdr[0] = 2
		hdr[1] = 0
		return arr, 2, func() { C.free(mem) }
	}

	mem := C.malloc(C.size_t(A) * C.size_t(unsafe.Sizeof(C.int64_t(0))))
	arr := (*C.int64_t)(mem)
	s := (*[1 << 30]C.int64_t)(unsafe.Pointer(arr))[:A:A]
	i := 0
	s[i] = C.int64_t(A)
	i++
	s[i] = C.int64_t(Cnt)
	i++

	for _, p := range ps {
		if len(p) == 0 {
			continue
		}
		s[i] = C.int64_t(len(p))
		i++
		s[i] = 0
		i++
		for _, q := range p {
			s[i] = C.int64_t(q.X)
			i++
			s[i] = C.int64_t(q.Y)
			i++
		}
	}
	return arr, A, func() { C.free(mem) }
}

func unpackPaths64(ptr *C.int64_t) Paths64 {
	if ptr == nil {
		return nil
	}
	// read A and C
	span := (*[1 << 30]C.int64_t)(unsafe.Pointer(ptr))
	A := int(span[0])
	Cnt := int(span[1])
	arr := span[:A:A]
	i := 2
	out := make(Paths64, 0, Cnt)
	for k := 0; k < Cnt; k++ {
		N := int(arr[i])
		i++
		_ = arr[i] // zero
		i++
		p := make(Path64, N)
		for j := 0; j < N; j++ {
			x := int64(arr[i])
			y := int64(arr[i+1])
			i += 2
			p[j] = Point64{X: x, Y: y}
		}
		out = append(out, p)
	}
	return out
}

// BooleanOp64: thin wrapper around the DLL export.
func BooleanOp64(clipType, fillRule uint8, subjects, subjectsOpen, clips Paths64) (solution Paths64, solutionOpen Paths64, _ error) {
	inSubj, _, freeSubj := packPaths64(subjects)
	defer freeSubj()
	inOpen, _, freeOpen := packPaths64(subjectsOpen)
	defer freeOpen()
	inClps, _, freeClps := packPaths64(clips)
	defer freeClps()

	var outSol *C.int64_t
	var outSolOpen *C.int64_t

	ok := C.BooleanOp64(
		C.uchar(clipType), C.uchar(fillRule),
		inSubj, inOpen, inClps,
		(*C.int64_t)(&outSol), (*C.int64_t)(&outSolOpen),
		true, false,
	)
	if ok == 0 {
		return nil, nil, errors.New("BooleanOp64 failed")
	}

	// Convert outSol and outSolOpen back to Go, then free via DisposeArray64.
	solution = unpackPaths64(outSol)
	if outSol != nil {
		C.DisposeArray64(outSol)
	}

	solutionOpen = unpackPaths64(outSolOpen)
	if outSolOpen != nil {
		C.DisposeArray64(outSolOpen)
	}

	return solution, solutionOpen, nil
}

// InflatePaths64 offsets the given paths by delta using the native library.
func InflatePaths64(paths Paths64, delta float64, joinType, endType uint8, miterLimit, arcTolerance float64) (Paths64, error) {
	in, _, freeIn := packPaths64(paths)
	defer freeIn()

	out := C.InflatePaths64(
		in,
		C.double(delta),
		C.uchar(joinType),
		C.uchar(endType),
		C.double(miterLimit),
		C.double(arcTolerance),
	)

	result := unpackPaths64(out)
	if out != nil {
		C.DisposeArray64(out)
	}
	return result, nil
}

// RectClip64 clips paths against a rectangular region defined by rect.
func RectClip64(rect Path64, paths Paths64) (Paths64, error) {
	if len(rect) < 4 {
		return nil, errors.New("rectangle must have at least 4 points")
	}
	left, right := rect[0].X, rect[0].X
	top, bottom := rect[0].Y, rect[0].Y
	for _, pt := range rect[1:] {
		if pt.X < left {
			left = pt.X
		}
		if pt.X > right {
			right = pt.X
		}
		if pt.Y < top {
			top = pt.Y
		}
		if pt.Y > bottom {
			bottom = pt.Y
		}
	}
	in, _, freeIn := packPaths64(paths)
	defer freeIn()
	out := C.RectClip64Wrapper(C.int64_t(left), C.int64_t(top), C.int64_t(right), C.int64_t(bottom), in)
	result := unpackPaths64(out)
	if out != nil {
		C.DisposeArray64(out)
	}
	return result, nil
}
