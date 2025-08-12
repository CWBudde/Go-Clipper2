//go:build clipper_cgo

package clipper

// convertToCAPI converts port types to capi types
func pathsToCAPI(paths Paths64) capi.Paths64 {
	result := make(capi.Paths64, len(paths))
	for i, path := range paths {
		capiPath := make(capi.Path64, len(path))
		for j, pt := range path {
			capiPath[j] = capi.Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = capiPath
	}
	return result
}

// convertFromCAPI converts capi types to port types
func pathsFromCAPI(paths capi.Paths64) Paths64 {
	result := make(Paths64, len(paths))
	for i, path := range paths {
		portPath := make(Path64, len(path))
		for j, pt := range path {
			portPath[j] = Point64{X: pt.X, Y: pt.Y}
		}
		result[i] = portPath
	}
	return result
}

// booleanOp64Impl delegates to the CGO oracle implementation
func booleanOp64Impl(clipType ClipType, fillRule FillRule, subjects, subjectsOpen, clips Paths64) (solution Paths64, solutionOpen Paths64, err error) {
	capiSubjects := pathsToCAPI(subjects)
	capiSubjectsOpen := pathsToCAPI(subjectsOpen)
	capiClips := pathsToCAPI(clips)

	capiSolution, capiSolutionOpen, err := capi.BooleanOp64(
		uint8(clipType),
		uint8(fillRule),
		capiSubjects,
		capiSubjectsOpen,
		capiClips,
	)
	if err != nil {
		return nil, nil, err
	}

	solution = pathsFromCAPI(capiSolution)
	solutionOpen = pathsFromCAPI(capiSolutionOpen)
	return solution, solutionOpen, nil
}

// inflatePathsImpl delegates to the CGO oracle implementation
func inflatePathsImpl(paths Paths64, delta float64, joinType JoinType, endType EndType, opts OffsetOptions) (Paths64, error) {
	capiPaths := pathsToCAPI(paths)
	capiResult, err := capi.InflatePaths64(capiPaths, delta, uint8(joinType), uint8(endType), opts.MiterLimit, opts.ArcTolerance)
	if err != nil {
		return nil, err
	}
	return pathsFromCAPI(capiResult), nil
}

// rectClipImpl delegates to the CGO oracle implementation
func rectClipImpl(rect Path64, paths Paths64) (Paths64, error) {
	capiRect := make(capi.Path64, len(rect))
	for i, pt := range rect {
		capiRect[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	capiPaths := pathsToCAPI(paths)
	capiResult, err := capi.RectClip64(capiRect, capiPaths)
	if err != nil {
		return nil, err
	}
	return pathsFromCAPI(capiResult), nil
}

// areaImpl calculates area using basic polygon area formula
func areaImpl(path Path64) float64 {
	if len(path) < 3 {
		return 0.0
	}

	area := 0.0
	for i := 0; i < len(path); i++ {
		j := (i + 1) % len(path)
		area += float64(path[i].X * path[j].Y)
		area -= float64(path[j].X * path[i].Y)
	}
	return area / 2.0
}
