package linkedin

import "errors"

var (
	ErrInvalidIdentity = errors.New("linkedin member identity is empty")
	ErrSessionExpired  = errors.New("linkedin session expired or forbidden")
	ErrNotFound        = errors.New("linkedin profile not found")
	ErrRateLimited     = errors.New("linkedin rate limited")
	ErrUpstream        = errors.New("linkedin upstream error")
)
