package manager

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fastPolicy retries quickly so tests don't sleep.
func fastPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:  2,
		BaseDelay:   time.Millisecond,
		MaxDelay:    2 * time.Millisecond,
		RetryStatus: []int{429, 502, 503, 504},
	}
}

func retryClient(p RetryPolicy) *http.Client {
	return &http.Client{Transport: newRetryTransport(http.DefaultTransport, p)}
}

func countingServer(t *testing.T, h func(n int32, w http.ResponseWriter, r *http.Request)) (*httptest.Server, *int32) {
	t.Helper()
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(atomic.AddInt32(&n, 1), w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

func TestRetryGetOn503ThenSucceeds(t *testing.T) {
	srv, n := countingServer(t, func(c int32, w http.ResponseWriter, _ *http.Request) {
		if c <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	resp, err := retryClient(fastPolicy()).Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(n); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
}

func TestRetryHonoursRetryAfterOn429(t *testing.T) {
	srv, n := countingServer(t, func(c int32, w http.ResponseWriter, _ *http.Request) {
		if c == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	resp, err := retryClient(fastPolicy()).Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(n); got != 2 {
		t.Fatalf("requests = %d, want 2", got)
	}
}

func TestRetrySkipsNonIdempotentOn503(t *testing.T) {
	srv, n := countingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	resp, err := retryClient(fastPolicy()).Post(srv.URL, "text/plain", strings.NewReader("hi"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if got := atomic.LoadInt32(n); got != 1 {
		t.Fatalf("requests = %d, want 1 (POST+503 must not retry)", got)
	}
}

func TestRetryReplaysBodyForPostOn429(t *testing.T) {
	srv, n := countingServer(t, func(c int32, w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "payload" {
			t.Errorf("request %d body = %q, want %q", c, body, "payload")
		}
		if c == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	resp, err := retryClient(fastPolicy()).Post(srv.URL, "text/plain", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(n); got != 2 {
		t.Fatalf("requests = %d, want 2 (429 retried, body replayed)", got)
	}
}

func TestRetryDisabledWhenMaxRetriesZero(t *testing.T) {
	srv, n := countingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	p := fastPolicy()
	p.MaxRetries = 0
	resp, err := retryClient(p).Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(n); got != 1 {
		t.Fatalf("requests = %d, want 1 (retries disabled)", got)
	}
}

func TestRetryExhaustsAndReturnsLastResponse(t *testing.T) {
	srv, n := countingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	resp, err := retryClient(fastPolicy()).Get(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
	if got := atomic.LoadInt32(n); got != 3 {
		t.Fatalf("requests = %d, want 3 (initial + 2 retries)", got)
	}
}

func TestConnectWiresRetries(t *testing.T) {
	srv, n := countingServer(t, func(c int32, w http.ResponseWriter, _ *http.Request) {
		if c == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[],"pagination":{"pages":1,"current":1}}`))
	})
	mgr, err := Connect(context.Background(), Options{
		BaseURL: srv.URL,
		Auth:    Bearer("TEST"),
		Retry:   func() *RetryPolicy { p := fastPolicy(); return &p }(),
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	for _, err := range mgr.Users.List(context.Background(), ListUsersQuery{}) {
		if err != nil {
			t.Fatalf("list: %v", err)
		}
	}
	if got := atomic.LoadInt32(n); got != 2 {
		t.Fatalf("requests = %d, want 2 (503 then 200)", got)
	}
}
