package error_handling

import "net/http"

type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFoundError(message string) *AppError {
	return &AppError{Code: CodeNotFound, Message: message, HTTPStatus: http.StatusNotFound}
}

func BadRequestError(message string) *AppError {
	return &AppError{Code: CodeBadRequest, Message: message, HTTPStatus: http.StatusBadRequest}
}

func ConflictError(message string) *AppError {
	return &AppError{Code: CodeConflict, Message: message, HTTPStatus: http.StatusConflict}
}

func InternalError(message string) *AppError {
	return &AppError{Code: CodeInternal, Message: message, HTTPStatus: http.StatusInternalServerError}
}

func InternalErrorWrap(message string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, HTTPStatus: http.StatusInternalServerError, Err: err}
}

func MethodNotAllowedError() *AppError {
	return &AppError{Code: CodeMethodNotAllowed, Message: "Method not allowed", HTTPStatus: http.StatusMethodNotAllowed}
}
