package validations

import "errors"

var (
	ErrInvalidRequestStatusTransition = errors.New(
		"invalid request status transition",
	)

	ErrConcurrentRequestStatusChange = errors.New(
		"request status changed concurrently",
	)
)
