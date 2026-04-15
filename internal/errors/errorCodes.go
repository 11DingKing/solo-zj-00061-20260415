package errors

import (
	"fmt"
	"net/http"
)

const prefix = "orders_"

const UnexpectedErrorMessage = "unexpected Error occurred, please try again later"

const (
	OrderGetInvalidParams = prefix + "get_invalid_params"
	OrderGetNotFound      = prefix + "get_not_found"
	OrdersGetServerError  = prefix + "get_server_error"

	OrderCreateInvalidInput = prefix + "create_invalid_input"
	OrderCreateServerError  = prefix + "create_server_error"

	OrderDeleteInvalidID   = prefix + "delete_invalid_order_id"
	OrderDeleteNotFound    = prefix + "delete_not_found"
	OrderDeleteServerError = prefix + "delete_server_error"
)

type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeInternal   ErrorType = "internal"
)

type AppError struct {
	ErrType    ErrorType
	StatusCode int
	ErrorCode  string
	Message    string
	DebugID    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(errorCode, message, debugID string, err error) *AppError {
	return &AppError{
		ErrType:    ErrorTypeValidation,
		StatusCode: http.StatusBadRequest,
		ErrorCode:  errorCode,
		Message:    message,
		DebugID:    debugID,
		Err:        err,
	}
}

func NewNotFoundError(errorCode, message, debugID string, err error) *AppError {
	return &AppError{
		ErrType:    ErrorTypeNotFound,
		StatusCode: http.StatusNotFound,
		ErrorCode:  errorCode,
		Message:    message,
		DebugID:    debugID,
		Err:        err,
	}
}

func NewInternalError(errorCode, message, debugID string, err error) *AppError {
	return &AppError{
		ErrType:    ErrorTypeInternal,
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  errorCode,
		Message:    message,
		DebugID:    debugID,
		Err:        err,
	}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if err == nil {
		return nil, false
	}
	if e, ok := err.(*AppError); ok {
		return e, true
	}
	return nil, false
}
