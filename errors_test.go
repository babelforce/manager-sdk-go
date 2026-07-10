package manager

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestResultDecodesUnexpected2xxPayload(t *testing.T) {
	// The generated clients only fill the typed field on an exact spec-status match (e.g.
	// JSON201). A live 200 where the spec says 201 leaves it nil; result must decode the body
	// itself instead of misreporting the successful call as an APIError.
	type payload struct {
		Id string `json:"id"`
	}
	resp := &http.Response{StatusCode: http.StatusOK}
	out, err := result[payload](nil, resp, []byte(`{"id":"abc"}`))
	if err != nil {
		t.Fatalf("expected decoded payload, got error: %v", err)
	}
	if out == nil || out.Id != "abc" {
		t.Fatalf("decoded payload = %+v", out)
	}
}

func TestResultUndecodable2xxIsNotAPIError(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusOK}
	_, err := result[struct{}](nil, resp, []byte("not json"))
	if err == nil {
		t.Fatal("expected an error for a 2xx response with an undecodable body")
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("expected a decode error, got *APIError implying the API rejected the call: %v", err)
	}
	if !strings.Contains(err.Error(), "200") {
		t.Fatalf("error should name the 2xx status, got: %v", err)
	}
}

func TestResultNon2xxStaysAPIError(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusNotFound}
	_, err := result[struct{}](nil, resp, []byte(`{"code":"NOT_FOUND","message":"missing"}`))
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 || apiErr.Code != "NOT_FOUND" {
		t.Fatalf("unexpected APIError: %+v", apiErr)
	}
}
