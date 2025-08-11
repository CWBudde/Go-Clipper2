//go:build clipper_cgo

package clipper

import "github.com/go-clipper/clipper2/capi"

// rectClipOracleImpl provides CGO oracle implementation when clipper_cgo build tag is active
func rectClipOracleImpl(rect Path64, paths Paths64) (Paths64, error) {
	// Convert to CAPI types
	capiRect := make(capi.Path64, len(rect))
	for i, pt := range rect {
		capiRect[i] = capi.Point64{X: pt.X, Y: pt.Y}
	}
	
	capiPaths := pathsToCAPI(paths)
	
	// Call CAPI RectClip64
	result, err := capi.RectClip64(capiRect, capiPaths)
	if err != nil {
		return nil, err
	}
	
	// Convert back to port types
	return pathsFromCAPI(result), nil
}

// isRealOracleModeImpl returns true in CGO mode
func isRealOracleModeImpl() bool {
	return true
}