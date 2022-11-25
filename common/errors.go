package common

import "errors"

// ErrInvalidTimeout signals that an invalid timeout period has been provided
var ErrInvalidTimeout = errors.New("invalid timeout value")

// ErrNilArwenChangeLocker signals that a nil arwen change locker has been provided
var ErrNilArwenChangeLocker = errors.New("nil arwen change locker")

// ErrNilExclusionTree signals that a nil exclusion tree was provided
var ErrNilExclusionTree = errors.New("nil exclusion tree")
