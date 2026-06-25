package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestCallControl(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/calls/test":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `"},"success":true}`))
		case strings.HasSuffix(r.URL.Path, "/hangup"):
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		case strings.HasSuffix(r.URL.Path, "/session/set"):
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if _, err := mgr.Calls.Get(ctx, uuidA); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calls.Hangup(ctx, uuidA); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calls.CreateTestCall(ctx, managerapi.CreateTestCall{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calls.SetSessionVariables(ctx, uuidA, managerapi.SetCallSessionVariablesRequest{Variables: &map[string]interface{}{"app.foo": "bar"}}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /api/v2/calls/" + uuidA,
		"POST /api/v2/calls/" + uuidA + "/hangup",
		"POST /api/v2/calls/test",
		"PUT /api/v2/calls/" + uuidA + "/session/set",
	}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %v, want %v", seen, want)
	}
}

func TestSmsNumbersConferences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/sms":
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"},{"id":"` + uuidB + `"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
		case strings.HasPrefix(p, "/api/v2/sms/"):
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		case p == "/api/v2/numbers":
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`))
		case strings.HasSuffix(p, "/tags"):
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		case strings.HasPrefix(p, "/api/v2/numbers/"):
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		case p == "/api/v2/conferences":
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"},{"id":"` + uuidB + `"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
		case strings.HasPrefix(p, "/api/v2/conferences/"):
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	sms, err := mgr.Sms.ListAll(ctx, managerapi.ListSmssParams{})
	if err != nil || len(sms) != 2 {
		t.Fatalf("sms.ListAll: err=%v len=%d", err, len(sms))
	}
	if _, err := mgr.Sms.Get(ctx, uuidA); err != nil {
		t.Fatal(err)
	}

	nums, err := mgr.Numbers.ListAll(ctx, managerapi.ListServiceNumbersParams{})
	if err != nil || len(nums) != 1 {
		t.Fatalf("numbers.ListAll: err=%v len=%d", err, len(nums))
	}
	if _, err := mgr.Numbers.Get(ctx, uuidA); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Numbers.AddTags(ctx, uuidA, []managerapi.Tag{"vip", "sales"}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Numbers.Update(ctx, uuidA, map[string]any{"name": "main line"}); err != nil {
		t.Fatal(err)
	}

	confs, err := mgr.Conferences.ListAll(ctx, managerapi.ListConferencesParams{})
	if err != nil || len(confs) != 2 {
		t.Fatalf("conferences.ListAll: err=%v len=%d", err, len(confs))
	}
	if _, err := mgr.Conferences.Get(ctx, uuidA); err != nil {
		t.Fatal(err)
	}
}
