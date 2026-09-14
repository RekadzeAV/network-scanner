// Package apperror предоставляет единый слой обработки ошибок приложения:
// стабильные коды, обёртки с контекстом, извлечение кода из цепочки,
// конвертацию паник в ошибки и контекстное логирование.
package apperror

import "net/http"

// Code — стабильный машинно-читаемый код ошибки.
//
// Коды не меняются: клиент может ветвиться по ним, не разбирая текст.
type Code string

const (
	CodeUnknown           Code = "unknown"
	CodeInvalidInput      Code = "invalid_input"
	CodeNotFound          Code = "not_found"
	CodeAlreadyExists     Code = "already_exists"
	CodePermissionDenied  Code = "permission_denied"
	CodeConflict          Code = "conflict"
	CodeTimeout           Code = "timeout"
	CodeCanceled          Code = "canceled"
	CodeUnavailable       Code = "unavailable"
	CodeResourceExhausted Code = "resource_exhausted"
	CodeUnimplemented     Code = "unimplemented"
	CodeInternal          Code = "internal"
	CodeDataLoss          Code = "data_loss"
)

// Valid — известен ли код.
func (c Code) Valid() bool {
	switch c {
	case CodeUnknown, CodeInvalidInput, CodeNotFound, CodeAlreadyExists,
		CodePermissionDenied, CodeConflict, CodeTimeout, CodeCanceled,
		CodeUnavailable, CodeResourceExhausted, CodeUnimplemented,
		CodeInternal, CodeDataLoss:
		return true
	default:
		return false
	}
}

// Retryable — имеет ли смысл повторить операцию.
func (c Code) Retryable() bool {
	switch c {
	case CodeTimeout, CodeUnavailable, CodeResourceExhausted:
		return true
	default:
		return false
	}
}

// HTTPStatus — отображение кода на HTTP-статус для API слоя.
func (c Code) HTTPStatus() int {
	switch c {
	case CodeInvalidInput:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists, CodeConflict:
		return http.StatusConflict
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeCanceled:
		return 499 // Client Closed Request (nginx), нет в net/http
	case CodeUnimplemented:
		return http.StatusNotImplemented
	case CodeResourceExhausted:
		return http.StatusTooManyRequests
	case CodeDataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
