package derrors

import "errors"

var (
	// ErrTopicNotFound topic not found error
	ErrTopicNotFound = errors.New("topic not found")
)
