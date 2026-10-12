package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestSmsSendReportDelete(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v2/sms" && m == http.MethodPost:
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/sms/test" && m == http.MethodPost:
			_, _ = w.Write([]byte(item))
		case p == "/api/v2/sms/reporting" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `"},{"id":"` + uuidB + `"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
		case strings.HasPrefix(p, "/api/v2/sms/") && m == http.MethodDelete:
			_, _ = w.Write([]byte(item))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if _, err := mgr.Sms.Send(ctx, managerapi.SmsSendRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Sms.TestInbound(ctx, managerapi.SmsSendRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Sms.Delete(ctx, uuidA); err != nil {
		t.Fatal(err)
	}

	var n int
	for _, err := range mgr.Sms.Report(ctx, managerapi.ReportSmsParams{}) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Fatalf("Report iterated %d, want 2", n)
	}

	// Path ids reach the wire normalized to the API's unhyphenated form (see ids.go).
	want := []string{
		"POST /api/v2/sms",
		"POST /api/v2/sms/test",
		"DELETE /api/v2/sms/" + strings.ReplaceAll(uuidA, "-", ""),
		"GET /api/v2/sms/reporting",
	}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %v, want %v", seen, want)
	}
}
