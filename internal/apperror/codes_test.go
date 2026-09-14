package apperror

import (
	"net/http"
	"testing"
)

func TestCodeValid(t *testing.T) {
	known := []Code{
		CodeUnknown, CodeInvalidInput, CodeNotFound, CodeAlreadyExists,
		CodePermissionDenied, CodeConflict, CodeTimeout, CodeCanceled,
		CodeUnavailable, CodeResourceExhausted, CodeUnimplemented,
		CodeInternal, CodeDataLoss,
	}
	for _, code := range known {
		if !code.Valid() {
			t.Errorf("code %q should be valid", code)
		}
	}

	for _, code := range []Code{"", "boom", "Invalid_Input"} {
		if code.Valid() {
			t.Errorf("code %q should be invalid", code)
		}
	}
}

func TestCodeRetryable(t *testing.T) {
	retryable := []Code{CodeTimeout, CodeUnavailable, CodeResourceExhausted}
	for _, code := range retryable {
		if !code.Retryable() {
			t.Errorf("code %q should be retryable", code)
		}
	}

	notRetryable := []Code{CodeNotFound, CodeInvalidInput, CodeInternal, CodeCanceled, CodeConflict}
	for _, code := range notRetryable {
		if code.Retryable() {
			t.Errorf("code %q should not be retryable", code)
		}
	}
}

func TestCodeHTTPStatus(t *testing.T) {
	tests := []struct {
		code Code
		want int
	}{
		{CodeInvalidInput, http.StatusBadRequest},
		{CodeNotFound, http.StatusNotFound},
		{CodeAlreadyExists, http.StatusConflict},
		{CodeConflict, http.StatusConflict},
		{CodePermissionDenied, http.StatusForbidden},
		{CodeUnavailable, http.StatusServiceUnavailable},
		{CodeTimeout, http.StatusGatewayTimeout},
		{CodeCanceled, 499},
		{CodeUnimplemented, http.StatusNotImplemented},
		{CodeResourceExhausted, http.StatusTooManyRequests},
		{CodeDataLoss, http.StatusInternalServerError},
		{CodeInternal, http.StatusInternalServerError},
		{CodeUnknown, http.StatusInternalServerError},
		{Code("custom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := tt.code.HTTPStatus(); got != tt.want {
			t.Errorf("Code(%q).HTTPStatus() = %d, want %d", tt.code, got, tt.want)
		}
	}
}
