package apperror

import "net/http"

type AppError struct {
	Code    int
	Message string
	ErrCode string
}

func (e *AppError) Error() string { return e.Message }

func BadRequest(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg, ErrCode: "BAD_REQUEST"}
}

func NotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, ErrCode: "NOT_FOUND"}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, ErrCode: "CONFLICT"}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg, ErrCode: "UNAUTHORIZED"}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg, ErrCode: "FORBIDDEN"}
}

func Internal(msg string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, ErrCode: "INTERNAL_ERROR"}
}

func TokenExpired() *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: "token expired", ErrCode: "TOKEN_EXPIRED"}
}
