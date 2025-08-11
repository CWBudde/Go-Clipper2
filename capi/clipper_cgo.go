//go:build clipper_cgo

package capi

/*
#cgo CXXFLAGS: -std=c++17 -fPIC -I${SRCDIR}/../third_party/clipper2/CPP/Clipper2Lib/include
#cgo darwin LDFLAGS: -lstdc++
#cgo linux  LDFLAGS: -lstdc++

#include "clipper_bridge.h"
*/
import "C"
import (
    "errors"
    "unsafe"
)

type Point64 struct{ X, Y int64 }
type Path64 []Point64
type Paths64 []Path64

var ErrClipper = errors.New("clipper2c error")

// --- pack/unpack between Go Paths64 and cpaths64 (AoS) -----------------------

func toCPaths64(ps Paths64) (C.cpaths64, func()) {
    var cp C.cpaths64
    n := len(ps)
    if n == 0 {
        return cp, func() {}
    }
    cp.len = C.int(n)
    cp.items = (*C.cpath64)(C.malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(C.cpath64{}))))

    items := unsafe.Slice(cp.items, n)
    // allocate each path pts
    for i, p := range ps {
        items[i].len = C.int(len(p))
        if len(p) > 0 {
            items[i].pts = (*C.cpt64)(C.malloc(C.size_t(len(p)) * C.size_t(unsafe.Sizeof(C.cpt64{}))))
            pts := unsafe.Slice(items[i].pts, len(p))
            for j, pt := range p {
                pts[j].x = C.int64_t(pt.X)
                pts[j].y = C.int64_t(pt.Y)
            }
        } else {
            items[i].pts = nil
        }
    }
    cleanup := func() {
        for i := 0; i < n; i++ {
            it := &unsafe.Slice(cp.items, n)[i]
            if it.pts != nil {
                C.free(unsafe.Pointer(it.pts))
            }
        }
        if cp.items != nil {
            C.free(unsafe.Pointer(cp.items))
        }
    }
    return cp, cleanup
}

func fromCPaths64(cp *C.cpaths64) Paths64 {
    if cp == nil || cp.items == nil || cp.len == 0 {
        return nil
    }
    n := int(cp.len)
    items := unsafe.Slice(cp.items, n)
    out := make(Paths64, n)
    for i := 0; i < n; i++ {
        m := int(items[i].len)
        if m > 0 && items[i].pts != nil {
            pts := unsafe.Slice(items[i].pts, m)
            p := make(Path64, m)
            for j := 0; j < m; j++ {
                p[j] = Point64{X: int64(pts[j].x), Y: int64(pts[j].y)}
            }
            out[i] = p
        } else {
            out[i] = nil
        }
    }
    return out
}

// --- exported Go APIs ---------------------------------------------------------

func BooleanOp64(clipType, fillRule uint8, subjects, subjectsOpen, clips Paths64) (Paths64, Paths64, error) {
    cs, cleanS := toCPaths64(subjects)
    defer cleanS()
    cso, cleanSO := toCPaths64(subjectsOpen)
    defer cleanSO()
    cc, cleanC := toCPaths64(clips)
    defer cleanC()

    var outClosed C.cpaths64
    var outOpen C.cpaths64

    rc := C.clipper2c_boolean64(
        C.c_cliptype(clipType),
        C.c_fillrule(fillRule),
        &cs, &cso, &cc,
        &outClosed, &outOpen,
    )
    if rc != 0 {
        return nil, nil, ErrClipper
    }
    defer C.clipper2c_free_paths(&outClosed)
    defer C.clipper2c_free_paths(&outOpen)

    return fromCPaths64(&outClosed), fromCPaths64(&outOpen), nil
}

func InflatePaths64(paths Paths64, delta float64, joinType, endType uint8, miterLimit, arcTolerance float64) (Paths64, error) {
    cp, clean := toCPaths64(paths)
    defer clean()

    var out C.cpaths64
    rc := C.clipper2c_offset64(
        &cp, C.double(delta),
        C.c_jointype(joinType), C.c_endtype(endType),
        C.double(miterLimit), C.double(arcTolerance),
        &out,
    )
    if rc != 0 {
        return nil, ErrClipper
    }
    defer C.clipper2c_free_paths(&out)

    return fromCPaths64(&out), nil
}

func RectClip64(rect Path64, paths Paths64) (Paths64, error) {
    if len(rect) < 2 {
        return nil, errors.New("rect needs 2+ points (lt rb) or a bbox poly")
    }
    // compute bbox
    l, t, r, b := rect[0].X, rect[0].Y, rect[0].X, rect[0].Y
    for _, p := range rect {
        if p.X < l { l = p.X }
        if p.X > r { r = p.X }
        if p.Y < t { t = p.Y }
        if p.Y > b { b = p.Y }
    }

    cp, clean := toCPaths64(paths)
    defer clean()

    var out C.cpaths64
    rc := C.clipper2c_rectclip64(
        C.int64_t(l), C.int64_t(t), C.int64_t(r), C.int64_t(b),
        &cp, &out,
    )
    if rc != 0 {
        return nil, ErrClipper
    }
    defer C.clipper2c_free_paths(&out)

    return fromCPaths64(&out), nil
}