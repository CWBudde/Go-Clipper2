package clipper

import "errors"

var (
	// ErrInvalidRectangle indicates an invalid rectangle was provided
	ErrInvalidRectangle = errors.New("invalid rectangle: must have exactly 4 points")

	// ErrNotImplemented indicates a feature is not yet implemented
	ErrNotImplemented = errors.New("not implemented yet")

	// ErrInvalidInput indicates invalid input parameters (deprecated: use more specific errors)
	ErrInvalidInput = errors.New("invalid input parameters")

	// ErrClipperExecution indicates the clipper algorithm failed during execution
	ErrClipperExecution = errors.New("clipper execution failed")

	// ErrEmptyPath indicates a nil or empty path was provided where a valid path is required
	ErrEmptyPath = errors.New("empty or nil path")

	// ErrDegeneratePolygon indicates a polygon with fewer than 3 points
	ErrDegeneratePolygon = errors.New("degenerate polygon: must have at least 3 points")

	// ErrInvalidFillRule indicates an out-of-range or invalid fill rule
	ErrInvalidFillRule = errors.New("invalid fill rule: must be EvenOdd, NonZero, Positive, or Negative")

	// ErrInvalidClipType indicates an out-of-range or invalid clip type
	ErrInvalidClipType = errors.New("invalid clip type: must be Intersection, Union, Difference, or Xor")

	// ErrInvalidParameter indicates an invalid numeric parameter (negative where positive required, NaN, Inf, etc.)
	ErrInvalidParameter = errors.New("invalid parameter value")

	// ErrInvalidOptions indicates invalid option struct values (miterLimit <= 0, arcTolerance <= 0, etc.)
	ErrInvalidOptions = errors.New("invalid options")

	// ErrInvalidJoinType indicates an out-of-range or invalid join type
	ErrInvalidJoinType = errors.New("invalid join type: must be JoinSquare, JoinBevel, JoinRound, or JoinMiter")

	// ErrInvalidEndType indicates an out-of-range or invalid end type
	ErrInvalidEndType = errors.New("invalid end type: must be EndPolygon, EndJoined, EndButt, EndSquare, or EndRound")

	// ErrInt32Overflow indicates a coordinate value exceeds the int32 range (-2147483648 to 2147483647)
	ErrInt32Overflow = errors.New("coordinate value exceeds int32 range")

	// ErrResultOverflow indicates an operation result doesn't fit in int32 range
	ErrResultOverflow = errors.New("operation result exceeds int32 range")
)
