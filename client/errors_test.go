package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSpaceshipApiError_Error_WithMessage(t *testing.T) {
	err := &SpaceshipApiError{Status: 400, Message: "bad request"}
	expected := "spaceship api error (status 400): bad request"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestSpaceshipApiError_Error_WithoutMessage(t *testing.T) {
	err := &SpaceshipApiError{Status: 500}
	expected := "spaceship api error (status 500)"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestSpaceshipApiError_Error_Nil(t *testing.T) {
	var err *SpaceshipApiError
	if err.Error() != "<nil>" {
		t.Errorf("expected %q, got %q", "<nil>", err.Error())
	}
}

func TestIsNotFoundError_True(t *testing.T) {
	err := &SpaceshipApiError{Status: http.StatusNotFound, Message: "not found"}
	if !IsNotFoundError(err) {
		t.Error("expected IsNotFoundError to return true for 404")
	}
}

func TestIsNotFoundError_False_OtherStatus(t *testing.T) {
	err := &SpaceshipApiError{Status: http.StatusBadRequest}
	if IsNotFoundError(err) {
		t.Error("expected IsNotFoundError to return false for 400")
	}
}

func TestIsNotFoundError_False_NonApiError(t *testing.T) {
	err := fmt.Errorf("some other error")
	if IsNotFoundError(err) {
		t.Error("expected IsNotFoundError to return false for non-API error")
	}
}

func TestIsNotFoundError_WrappedError(t *testing.T) {
	apiErr := &SpaceshipApiError{Status: http.StatusNotFound}
	wrapped := fmt.Errorf("wrapped: %w", apiErr)
	if !IsNotFoundError(wrapped) {
		t.Error("expected IsNotFoundError to return true for wrapped 404")
	}
}

func TestIsNotFoundError_Nil(t *testing.T) {
	if IsNotFoundError(nil) {
		t.Error("expected IsNotFoundError to return false for nil")
	}
}

func TestSpaceshipApiError_ImplementsError(t *testing.T) {
	var err error = &SpaceshipApiError{Status: 500}
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		t.Error("expected SpaceshipApiError to implement error interface")
	}
}

func testResponse(status int, headers map[string]string, body string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestErrorFromResponse_RetryAfterParsed(t *testing.T) {
	c := &Client{}
	err := c.errorFromResponse(testResponse(429, map[string]string{"Retry-After": "280"}, "slow down"))
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SpaceshipApiError, got %T", err)
	}
	if apiErr.RetryAfter != 280*time.Second {
		t.Errorf("expected RetryAfter 280s, got %s", apiErr.RetryAfter)
	}
	if apiErr.Status != 429 || apiErr.Message != "slow down" {
		t.Errorf("existing fields regressed: %+v", apiErr)
	}
}

func TestErrorFromResponse_RetryAfterMissing(t *testing.T) {
	c := &Client{}
	err := c.errorFromResponse(testResponse(429, nil, ""))
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SpaceshipApiError, got %T", err)
	}
	if apiErr.RetryAfter != 0 {
		t.Errorf("expected zero RetryAfter, got %s", apiErr.RetryAfter)
	}
}

func TestErrorFromResponse_RetryAfterNotANumber(t *testing.T) {
	c := &Client{}
	err := c.errorFromResponse(testResponse(429, map[string]string{"Retry-After": "soon"}, ""))
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SpaceshipApiError, got %T", err)
	}
	if apiErr.RetryAfter != 0 {
		t.Errorf("expected zero RetryAfter for unparsable header, got %s", apiErr.RetryAfter)
	}
}

func TestErrorFromResponse_RetryAfterNegative(t *testing.T) {
	c := &Client{}
	err := c.errorFromResponse(testResponse(429, map[string]string{"Retry-After": "-5"}, ""))
	var apiErr *SpaceshipApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SpaceshipApiError, got %T", err)
	}
	if apiErr.RetryAfter != 0 {
		t.Errorf("expected zero RetryAfter for negative header, got %s", apiErr.RetryAfter)
	}
}

func TestIsRateLimitError_True(t *testing.T) {
	err := &SpaceshipApiError{Status: http.StatusTooManyRequests}
	if !IsRateLimitError(err) {
		t.Error("expected IsRateLimitError to return true for 429")
	}
}

func TestIsRateLimitError_False_OtherStatus(t *testing.T) {
	err := &SpaceshipApiError{Status: http.StatusBadRequest}
	if IsRateLimitError(err) {
		t.Error("expected IsRateLimitError to return false for 400")
	}
}

func TestIsRateLimitError_False_NonApiError(t *testing.T) {
	if IsRateLimitError(fmt.Errorf("some other error")) {
		t.Error("expected IsRateLimitError to return false for non-API error")
	}
}

func TestIsRateLimitError_WrappedError(t *testing.T) {
	wrapped := fmt.Errorf("wrapped: %w", &SpaceshipApiError{Status: http.StatusTooManyRequests})
	if !IsRateLimitError(wrapped) {
		t.Error("expected IsRateLimitError to return true for wrapped 429")
	}
}

func TestIsRateLimitError_Nil(t *testing.T) {
	if IsRateLimitError(nil) {
		t.Error("expected IsRateLimitError to return false for nil")
	}
}
