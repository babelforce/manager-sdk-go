package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func systemServer(t *testing.T, gotQuery *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method %s on %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v2/echo":
			_, _ = w.Write([]byte(`{"item":{"hello":"world"},"success":true}`))
		case "/api/v2/ping":
			_, _ = w.Write([]byte(`{"item":{"pong":true},"success":true}`))
		case "/api/v2/status":
			_, _ = w.Write([]byte(`{"item":{"status":"ok"},"success":true}`))
		case "/api/v2/data/time":
			_, _ = w.Write([]byte(`{"item":{"time":"2026-06-25T00:00:00Z"},"success":true}`))
		case "/api/v2/data/timezones":
			if gotQuery != nil {
				*gotQuery = r.URL.RawQuery
			}
			_, _ = w.Write([]byte(`{"items":[{"id":"Europe/Berlin"}],"success":true}`))
		case "/api/v2/push-token":
			_, _ = w.Write([]byte(`{"item":{"token":"tok"},"success":true}`))
		case "/api/v2/tags":
			_, _ = w.Write([]byte(`{"items":[{"id":"t1"},{"id":"t2"}],"success":true}`))
		case "/api/v2/tags/colors":
			_, _ = w.Write([]byte(`{"items":[{"id":"red"}],"success":true}`))
		case "/api/v2/templates/export/all":
			_, _ = w.Write([]byte(`{"applications":[{"id":"a1"}],"version":2}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestSystemResource(t *testing.T) {
	var gotQuery string
	srv := systemServer(t, &gotQuery)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	echo, err := mgr.System.Echo(ctx)
	if err != nil || echo == nil || (*echo.Item)["hello"] != "world" {
		t.Fatalf("echo: %v %+v", err, echo)
	}

	ping, err := mgr.System.Ping(ctx)
	if err != nil || ping == nil || (*ping.Item)["pong"] != true {
		t.Fatalf("ping: %v %+v", err, ping)
	}

	status, err := mgr.System.ApiStatus(ctx)
	if err != nil || status == nil || (*status.Item)["status"] != "ok" {
		t.Fatalf("apiStatus: %v %+v", err, status)
	}

	st, err := mgr.System.ServerTime(ctx)
	if err != nil || st == nil || (*st.Item)["time"] != "2026-06-25T00:00:00Z" {
		t.Fatalf("serverTime: %v %+v", err, st)
	}

	pt, err := mgr.System.PushToken(ctx)
	if err != nil || pt == nil || (*pt.Item)["token"] != "tok" {
		t.Fatalf("pushToken: %v %+v", err, pt)
	}

	tags, err := mgr.System.Tags(ctx)
	if err != nil || tags == nil || len(tags.Items) != 2 {
		t.Fatalf("tags: %v %+v", err, tags)
	}

	byCat, err := mgr.System.TagsByCategory(ctx, "colors")
	if err != nil || byCat == nil || len(byCat.Items) != 1 || byCat.Items[0]["id"] != "red" {
		t.Fatalf("tagsByCategory: %v %+v", err, byCat)
	}

	tmpl, err := mgr.System.ExportTemplates(ctx, "all")
	if err != nil || tmpl == nil {
		t.Fatalf("exportTemplates: %v %+v", err, tmpl)
	}
	if _, ok := tmpl["applications"]; !ok {
		t.Fatalf("exportTemplates missing applications key: %+v", tmpl)
	}
}

func TestSystemTimezonesQuery(t *testing.T) {
	var gotQuery string
	srv := systemServer(t, &gotQuery)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}

	q := "berlin"
	max := 5
	tzs, err := mgr.System.Timezones(context.Background(), managerapi.ListTimezonesParams{Q: &q, Max: &max})
	if err != nil || tzs == nil || len(tzs.Items) != 1 {
		t.Fatalf("timezones: %v %+v", err, tzs)
	}

	vals, perr := url.ParseQuery(gotQuery)
	if perr != nil {
		t.Fatalf("parse query %q: %v", gotQuery, perr)
	}
	if vals.Get("q") != "berlin" {
		t.Errorf("expected q=berlin, got query %q", gotQuery)
	}
	if vals.Get("max") != "5" {
		t.Errorf("expected max=5, got query %q", gotQuery)
	}
}
