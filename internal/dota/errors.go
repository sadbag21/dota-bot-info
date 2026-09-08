package dota

import "errors"

var (
	ErrNotFound    = errors.New("resource not found")
	ErrRateLimited = errors.New("rate limit exceeded")
	ErrTimeout     = errors.New("request timeout")
	ErrUnavailable = errors.New("OpenDota unavailable")
	ErrBadResponse = errors.New("unexpected OpenDota response")
)
