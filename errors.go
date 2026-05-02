package uptime

import (
	"encoding/json"
	"fmt"
)

// APIError is returned when the Tickstem API responds with a 4xx or 5xx status.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tickstem/uptime: API error %d: %s", e.StatusCode, e.Message)
}

// IsUnauthorized reports whether the error is a 401 — invalid or revoked API key.
func IsUnauthorized(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == 401
}

// IsQuotaExceeded reports whether the monitor limit or plan interval minimum was hit.
func IsQuotaExceeded(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == 402
}

// IsNotFound reports whether the monitor does not exist or belongs to another account.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == 404
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

func parseAPIError(statusCode int, body []byte) *APIError {
	var errResp apiErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return &APIError{StatusCode: statusCode, Message: errResp.Error}
	}
	return &APIError{StatusCode: statusCode, Message: string(body)}
}
