//go:build !clipper_cgo

package clipper

// rectClipOracleImpl provides stub implementation when CGO is not available
// In pure Go mode, this just calls the pure implementation for self-consistency testing
func rectClipOracleImpl(rect Path64, paths Paths64) (Paths64, error) {
	// In pure Go mode, compare against the same implementation
	// This allows the fuzz test to run and validate internal consistency
	return RectClip64(rect, paths)
}

// isRealOracleModeImpl returns false in pure Go mode
func isRealOracleModeImpl() bool {
	return false
}