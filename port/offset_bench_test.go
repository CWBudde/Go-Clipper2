package clipper

import (
	"math"
	"testing"
)

// ==============================================================================
// Offset Performance Benchmarks
// ==============================================================================

// BenchmarkOffsetSimpleSquare benchmarks offset on a simple 4-vertex polygon
func BenchmarkOffsetSimpleSquare(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 10.0, JoinRound, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
	}
}

// BenchmarkOffsetComplexPolygon benchmarks offset on a 100-vertex polygon
func BenchmarkOffsetComplexPolygon(b *testing.B) {
	// Create star polygon with 100 vertices
	star := make(Path64, 100)
	for i := 0; i < 100; i++ {
		angle := float64(i) * 2 * math.Pi / 100
		radius := 500.0
		if i%2 == 0 {
			radius = 1000.0
		}
		star[i] = Point64{
			X: int64(2000 + radius*math.Cos(angle)),
			Y: int64(2000 + radius*math.Sin(angle)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(Paths64{star}, 10.0, JoinRound, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
	}
}

// BenchmarkOffsetMultiplePaths benchmarks offset on 10 separate polygons
func BenchmarkOffsetMultiplePaths(b *testing.B) {
	paths := make(Paths64, 10)
	for i := 0; i < 10; i++ {
		offset := int64(i * 200)
		paths[i] = Path64{
			{offset, offset},
			{offset + 100, offset},
			{offset + 100, offset + 100},
			{offset, offset + 100},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(paths, 5.0, JoinRound, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
	}
}

// BenchmarkOffsetJoinTypes benchmarks different join types
func BenchmarkOffsetJoinBevel(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 10.0, JoinBevel, EndPolygon, OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25})
	}
}

func BenchmarkOffsetJoinMiter(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 10.0, JoinMiter, EndPolygon, OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25})
	}
}

func BenchmarkOffsetJoinSquare(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 10.0, JoinSquare, EndPolygon, OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25})
	}
}

func BenchmarkOffsetJoinRound(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 10.0, JoinRound, EndPolygon, OffsetOptions{MiterLimit: 2.0, ArcTolerance: 0.25})
	}
}

// BenchmarkOffsetOpenPaths benchmarks open path offsetting
func BenchmarkOffsetOpenPath(b *testing.B) {
	line := Path64{{0, 0}, {500, 0}, {500, 500}, {1000, 500}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(Paths64{line}, 10.0, JoinRound, EndRound, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
	}
}

// BenchmarkOffsetLargeDelta benchmarks with large offset distance
func BenchmarkOffsetLargeDelta(b *testing.B) {
	square := Paths64{{{0, 0}, {1000, 0}, {1000, 1000}, {0, 1000}}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InflatePaths64(square, 1000.0, JoinRound, EndPolygon, OffsetOptions{
			MiterLimit:   2.0,
			ArcTolerance: 0.25,
		})
	}
}
