package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/syncride/backend/shared/constants"
)

// AppError represents a domain error with code, HTTP status, and detailed context
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"status"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Err        error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// RFC 7807 Problem Details Standard Output
type ProblemDetails struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Status    int                    `json:"status"`
	Detail    string                 `json:"detail"`
	Instance  string                 `json:"instance,omitempty"`
	Code      string                 `json:"code"`
	Timestamp string                 `json:"timestamp"`
	Errors    map[string]interface{} `json:"errors,omitempty"`
}

func (e *AppError) ToProblemDetails(instance string) ProblemDetails {
	return ProblemDetails{
		Type:      fmt.Sprintf("https://api.syncride.com/errors/%s", e.Code),
		Title:     http.StatusText(e.HTTPStatus),
		Status:    e.HTTPStatus,
		Detail:    e.Message,
		Instance:  instance,
		Code:      e.Code,
		Timestamp: e.Timestamp.Format(time.RFC3339),
		Errors:    e.Details,
	}
}

// Error Constructors
func NewInternalError(msg string, cause error) *AppError {
	return &AppError{
		Code:       constants.ErrCodeInternal,
		Message:    msg,
		HTTPStatus: http.StatusInternalServerError,
		Timestamp:  time.Now().UTC(),
		Err:        cause,
	}
}

func NewNotFoundError(resource string, identifier interface{}) *AppError {
	return &AppError{
		Code:       constants.ErrCodeNotFound,
		Message:    fmt.Sprintf("%s with identifier '%v' was not found", resource, identifier),
		HTTPStatus: http.StatusNotFound,
		Timestamp:  time.Now().UTC(),
	}
}

func NewBadRequestError(msg string) *AppError {
	return &AppError{
		Code:       constants.ErrCodeBadRequest,
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
		Timestamp:  time.Now().UTC(),
	}
}

func NewValidationError(msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       constants.ErrCodeValidation,
		Message:    msg,
		HTTPStatus: http.StatusUnprocessableEntity,
		Details:    details,
		Timestamp:  time.Now().UTC(),
	}
}

func NewUnauthorizedError(msg string) *AppError {
	return &AppError{
		Code:       constants.ErrCodeUnauthorized,
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
		Timestamp:  time.Now().UTC(),
	}
}

func NewForbiddenError(msg string) *AppError {
	return &AppError{
		Code:       constants.ErrCodeForbidden,
		Message:    msg,
		HTTPStatus: http.StatusForbidden,
		Timestamp:  time.Now().UTC(),
	}
}

func NewConflictError(msg string) *AppError {
	return &AppError{
		Code:       constants.ErrCodeConflict,
		Message:    msg,
		HTTPStatus: http.StatusConflict,
		Timestamp:  time.Now().UTC(),
	}
}

// WriteHTTPError formats and writes AppError as RFC 7807 JSON response
func WriteHTTPError(w http.ResponseWriter, reqPath string, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		appErr = NewInternalError("An unexpected internal error occurred", err)
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(appErr.HTTPStatus)

	problem := appErr.ToProblemDetails(reqPath)
	_ = json.NewEncoder(w).Encode(problem)
}
