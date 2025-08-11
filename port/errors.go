package clipper

import "errors"

var (
	// ErrInvalidRectangle indicates an invalid rectangle was provided
	ErrInvalidRectangle = errors.New("invalid rectangle: must have exactly 4 points")

	// ErrNotImplemented indicates a feature is not yet implemented
	ErrNotImplemented = errors.New("not implemented yet")

	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input parameters")
)
