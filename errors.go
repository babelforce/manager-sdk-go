package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError is returned when the manager API responds with a non-2xx status.
type APIError struct {
	// Status is the HTTP status code.
	Status int
	// Code is the API error code, when the body carries one.
	Code string
	// Message is a human-readable message (from the body when available, else the status text).
	Message string
	// Body is the raw response body.
	Body []byte
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("manager: %d %s (%s)", e.Status, e.Message, e.Code)
	}
	return fmt.Sprintf("manager: %d %s", e.Status, e.Message)
}

func newAPIError(resp *http.Response, body []byte) *APIError {
	e := &APIError{Body: body}
	if resp != nil {
		e.Status = resp.StatusCode
		e.Message = http.StatusText(resp.StatusCode)
	}
	var parsed map[string]any
	if json.Unmarshal(body, &parsed) == nil {
		if c, ok := parsed["code"].(string); ok {
			e.Code = c
		}
		for _, k := range []string{"message", "error", "detail"} {
			if m, ok := parsed[k].(string); ok && m != "" {
				e.Message = m
				break
			}
		}
	}
	return e
}

func isOK(resp *http.Response) bool {
	return resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300
}

// result returns the typed success payload, or an APIError if the response was not 2xx.
func result[T any](payload *T, resp *http.Response, body []byte) (*T, error) {
	if isOK(resp) && payload != nil {
		return payload, nil
	}
	return nil, newAPIError(resp, body)
}

// resultVoid checks for a 2xx status on responses without a typed body.
func resultVoid(resp *http.Response, body []byte) error {
	if isOK(resp) {
		return nil
	}
	return newAPIError(resp, body)
}
