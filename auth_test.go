package manager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestTokenManager returns a tokenManager wired to the given server's /oauth/token endpoint,
// mirroring the production client_credentials wiring.
func newTestTokenManager(srv *httptest.Server) *tokenManager {
	return &tokenManager{grant: func(ctx context.Context) (*TokenResponse, error) {
		return ClientCredentialsGrant(ctx, srv.Client(), srv.URL, "id", "secret")
	}}
}

func TestTokenManagerWaiterHonorsContext(t *testing.T) {
	// While one caller's grant hangs on the network, a second caller with a short context
	// deadline must return ctx.Err() within its deadline instead of blocking behind the grant.
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	}))
	defer srv.Close()
	defer unblock()

	tm := newTestTokenManager(srv)

	leaderDone := make(chan error, 1)
	go func() {
		_, err := tm.get(context.Background())
		leaderDone <- err
	}()
	<-entered // the leader's grant is now blocked inside the token endpoint

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	waiterDone := make(chan error, 1)
	go func() {
		_, err := tm.get(ctx)
		waiterDone <- err
	}()

	select {
	case err := <-waiterDone:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("waiter error = %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("waiter did not honor its context deadline while a grant was in flight")
	}

	unblock()
	if err := <-leaderDone; err != nil {
		t.Fatalf("leader grant failed: %v", err)
	}
	tok, err := tm.get(context.Background())
	if err != nil || tok != "tok" {
		t.Fatalf("get after grant = %q, %v; want \"tok\", nil", tok, err)
	}
}

func TestTokenManagerSingleFlight(t *testing.T) {
	// N concurrent callers with no cached token must trigger exactly one grant request.
	var grants atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grants.Add(1)
		time.Sleep(20 * time.Millisecond) // widen the window so racing callers overlap
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	}))
	defer srv.Close()

	tm := newTestTokenManager(srv)

	const n = 10
	toks := make([]string, n)
	errs := make([]error, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			toks[i], errs[i] = tm.get(context.Background())
		}()
	}
	close(start)
	wg.Wait()

	for i := range n {
		if errs[i] != nil || toks[i] != "tok" {
			t.Fatalf("caller %d: got %q, %v; want \"tok\", nil", i, toks[i], errs[i])
		}
	}
	if got := grants.Load(); got != 1 {
		t.Fatalf("expected exactly 1 grant request, got %d", got)
	}
}

func TestTokenManagerLeaderFailureDoesNotStrandWaiters(t *testing.T) {
	// A failed grant must not strand concurrent waiters: one of them becomes the next leader
	// and retries the grant.
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if hits.Add(1) == 1 {
			entered <- struct{}{}
			<-release
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"tok2","expires_in":3600}`))
	}))
	defer srv.Close()
	defer unblock()

	tm := newTestTokenManager(srv)

	leaderDone := make(chan error, 1)
	go func() {
		_, err := tm.get(context.Background())
		leaderDone <- err
	}()
	<-entered // the leader's grant is now blocked inside the token endpoint

	type res struct {
		tok string
		err error
	}
	waiterDone := make(chan res, 1)
	go func() {
		tok, err := tm.get(context.Background())
		waiterDone <- res{tok, err}
	}()

	unblock() // the leader's grant now fails
	if err := <-leaderDone; err == nil {
		t.Fatal("expected the leader's grant to fail")
	}
	select {
	case r := <-waiterDone:
		if r.err != nil || r.tok != "tok2" {
			t.Fatalf("waiter after leader failure = %q, %v; want \"tok2\", nil", r.tok, r.err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("waiter was stranded after the leader's grant failed")
	}
	if got := hits.Load(); got != 2 {
		t.Fatalf("expected 2 grant requests (failed + retried), got %d", got)
	}
}

// newFormCapturingTokenServer returns a token endpoint that captures each grant's form body on the
// channel and answers with a valid token response.
func newFormCapturingTokenServer(t *testing.T) (*httptest.Server, chan url.Values) {
	t.Helper()
	forms := make(chan url.Values, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		forms <- r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	}))
	t.Cleanup(srv.Close)
	return srv, forms
}

func TestPasswordWithClientIDSendsCustomClientID(t *testing.T) {
	srv, forms := newFormCapturingTokenServer(t)

	ed := PasswordWithClientID("user", "pass", "my-app").editor(srv.URL, srv.Client())
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/v2/ping", nil)
	if err := ed(context.Background(), req); err != nil {
		t.Fatalf("editor: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok" {
		t.Fatalf("Authorization = %q, want \"Bearer tok\"", got)
	}
	form := <-forms
	if got := form.Get("grant_type"); got != "password" {
		t.Fatalf("grant_type = %q, want \"password\"", got)
	}
	if got := form.Get("client_id"); got != "my-app" {
		t.Fatalf("client_id = %q, want \"my-app\"", got)
	}
}

func TestRefreshTokenWithSecretSendsClientSecret(t *testing.T) {
	srv, forms := newFormCapturingTokenServer(t)

	ed := RefreshTokenWithSecret("seed-refresh", "my-app", "s3cret").editor(srv.URL, srv.Client())
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/v2/ping", nil)
	if err := ed(context.Background(), req); err != nil {
		t.Fatalf("editor: %v", err)
	}
	form := <-forms
	if got := form.Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q, want \"refresh_token\"", got)
	}
	if got := form.Get("refresh_token"); got != "seed-refresh" {
		t.Fatalf("refresh_token = %q, want \"seed-refresh\"", got)
	}
	if got := form.Get("client_id"); got != "my-app" {
		t.Fatalf("client_id = %q, want \"my-app\"", got)
	}
	if got := form.Get("client_secret"); got != "s3cret" {
		t.Fatalf("client_secret = %q, want \"s3cret\"", got)
	}
}
