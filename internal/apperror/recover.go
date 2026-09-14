package apperror

import (
	"fmt"
	"runtime/debug"
)

// Recover — конвертирует значение recover() в ошибку с кодом CodeInternal.
//
// Возвращает nil для r == nil, чтобы вызывающий код мог писать
// `defer func() { err = apperror.Recover(recover(), "scan") }()`.
func Recover(r any, op string) *Error {
	if r == nil {
		return nil
	}

	var message string
	switch value := r.(type) {
	case error:
		message = value.Error()
	case string:
		message = value
	default:
		message = fmt.Sprintf("%v", value)
	}

	appErr := &Error{
		Code:    CodeInternal,
		Op:      op,
		Message: "panic: " + message,
		Cause:   asError(r),
	}
	return appErr.With("stack", string(debug.Stack()))
}

// RecoverTo — вариант для defer: записывает панику в указатель на ошибку.
//
//	func () (err error) {
//	    defer apperror.RecoverTo(&err, "scanner.scan")
//	    ...
//	}
//
// Существующее значение err не перезаписывается, если паники не было.
func RecoverTo(err *error, op string) {
	if r := recover(); r != nil {
		recovered := Recover(r, op)
		if err != nil {
			*err = recovered
		}
	}
}

// Safe — выполняет fn, превращая панику в ошибку вместо падения процесса.
func Safe(op string, fn func() error) (err error) {
	if fn == nil {
		return nil
	}
	defer RecoverTo(&err, op)
	return fn()
}

// SafeVoid — Safe для функций без результата.
func SafeVoid(op string, fn func()) (err error) {
	if fn == nil {
		return nil
	}
	defer RecoverTo(&err, op)
	fn()
	return nil
}

func asError(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	if r == nil {
		return nil
	}
	return fmt.Errorf("%v", r)
}
