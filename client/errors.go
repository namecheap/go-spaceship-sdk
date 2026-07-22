package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SpaceshipApiError represents a non-2xx error response from the Spaceship API.
// Status is the HTTP status code; Message is the trimmed response body and may
// be empty. RetryAfter is the wait requested by the Retry-After header
// (delta-seconds); zero when the header is absent or unparsable.
type SpaceshipApiError struct {
	Status     int
	Message    string
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *SpaceshipApiError) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Message != "" {
		return fmt.Sprintf("spaceship api error (status %d): %s", e.Status, e.Message)
	}

	return fmt.Sprintf("spaceship api error (status %d)", e.Status)
}

// errorFromResponse builds a *SpaceshipApiError from a non-2xx response. The
// body is read up to 64 KiB (LimitReader guards against an unexpectedly large
// body) and used, trimmed, as the error Message.
func (c *Client) errorFromResponse(resp *http.Response) error {
	retryAfter := parseRetryAfter(resp.Header)

	data, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return &SpaceshipApiError{
			Status:     resp.StatusCode,
			RetryAfter: retryAfter,
		}
	}

	return &SpaceshipApiError{
		Status:     resp.StatusCode,
		Message:    strings.TrimSpace(string(data)),
		RetryAfter: retryAfter,
	}
}

// parseRetryAfter reads the Retry-After header as delta-seconds (the only
// format the Spaceship API sends). Absent, unparsable, or negative values
// yield zero.
func parseRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	seconds, err := strconv.Atoi(v)
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// IsNotFoundError reports whether err is a *SpaceshipApiError with HTTP 404.
// Callers use it to treat "already gone" as success on delete paths.
func IsNotFoundError(err error) bool {
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Status == http.StatusNotFound
}

// IsRateLimitError reports whether err is a *SpaceshipApiError with HTTP 429.
// A 429 is rejected before execution, so the failed request is safe to retry;
// the error's RetryAfter carries the server-requested wait.
func IsRateLimitError(err error) bool {
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Status == http.StatusTooManyRequests
}
