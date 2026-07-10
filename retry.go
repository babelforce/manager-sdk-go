package manager

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy tunes automatic request retries. Transient failures — network errors and a small set
// of "try again" status codes (429/502/503/504 by default) — are retried with exponential backoff
// and jitter, honouring a Retry-After header when present.
//
// Retries are on by default with conservative settings (see [Connect]). To customise, set
// [Options.Retry]; zero-valued delay/status fields fall back to the defaults, but MaxRetries is
// taken literally — set it to 0 to disable retries.
type RetryPolicy struct {
	// MaxRetries is the number of retry attempts after the initial request. Default 2; 0 disables.
	MaxRetries int
	// BaseDelay is the base backoff; it grows exponentially per attempt. Default 250ms.
	BaseDelay time.Duration
	// MaxDelay caps any single backoff and also caps Retry-After. Default 10s.
	MaxDelay time.Duration
	// RetryStatus is the set of response status codes that trigger a retry.
	// Default: 429, 502, 503, 504.
	RetryStatus []int
}

var defaultRetryPolicy = RetryPolicy{
	MaxRetries:  2,
	BaseDelay:   250 * time.Millisecond,
	MaxDelay:    10 * time.Second,
	RetryStatus: []int{429, 502, 503, 504},
}

// effective fills any zero-valued tuning fields from the defaults. MaxRetries is left as-is so 0
// disables retries.
func (p RetryPolicy) effective() RetryPolicy {
	if p.BaseDelay <= 0 {
		p.BaseDelay = defaultRetryPolicy.BaseDelay
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = defaultRetryPolicy.MaxDelay
	}
	if len(p.RetryStatus) == 0 {
		p.RetryStatus = defaultRetryPolicy.RetryStatus
	}
	return p
}

// idempotentMethod reports whether a request method is safe to replay after a network error or 5xx.
func idempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

// retryTransport is an http.RoundTripper that retries transient failures per a RetryPolicy.
type retryTransport struct {
	base   http.RoundTripper
	policy RetryPolicy
	codes  map[int]bool
}

func newRetryTransport(base http.RoundTripper, p RetryPolicy) *retryTransport {
	codes := make(map[int]bool, len(p.RetryStatus))
	for _, c := range p.RetryStatus {
		codes[c] = true
	}
	return &retryTransport{base: base, policy: p, codes: codes}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Buffer the body so it can be replayed on each attempt.
	var body []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
		body = b
	}
	idempotent := idempotentMethod(req.Method)

	for attempt := 0; ; attempt++ {
		attemptReq := req.Clone(req.Context())
		if body != nil {
			attemptReq.Body = io.NopCloser(bytes.NewReader(body))
			attemptReq.ContentLength = int64(len(body))
		}

		resp, err := t.base.RoundTrip(attemptReq)
		wait, retry := t.shouldRetry(attempt, idempotent, resp, err)
		if !retry {
			return resp, err
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		select {
		case <-time.After(wait):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
}

// shouldRetry decides whether to retry and, if so, how long to wait first.
func (t *retryTransport) shouldRetry(attempt int, idempotent bool, resp *http.Response, err error) (time.Duration, bool) {
	if attempt >= t.policy.MaxRetries {
		return 0, false
	}
	if err != nil {
		// Network-level failure: only safe to replay idempotent requests.
		if !idempotent {
			return 0, false
		}
		return t.backoff(attempt), true
	}
	if !t.codes[resp.StatusCode] {
		return 0, false
	}
	// For non-idempotent methods, only retry on 429 (server rejected without acting on it).
	if !idempotent && resp.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}
	if d, ok := retryAfter(resp); ok {
		if d > t.policy.MaxDelay {
			d = t.policy.MaxDelay
		}
		return d, true
	}
	return t.backoff(attempt), true
}

// backoff returns an exponential delay with full jitter, capped at MaxDelay.
func (t *retryTransport) backoff(attempt int) time.Duration {
	ceiling := t.policy.BaseDelay << uint(attempt)
	if ceiling <= 0 || ceiling > t.policy.MaxDelay {
		ceiling = t.policy.MaxDelay
	}
	return time.Duration(rand.Int63n(int64(ceiling) + 1))
}

// retryAfter parses a Retry-After header (delta-seconds or HTTP-date) into a duration.
func retryAfter(resp *http.Response) (time.Duration, bool) {
	h := resp.Header.Get("Retry-After")
	if h == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(h); err == nil {
		if secs < 0 {
			secs = 0
		}
		return time.Duration(secs) * time.Second, true
	}
	if when, err := http.ParseTime(h); err == nil {
		d := time.Until(when)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}
