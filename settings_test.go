package manager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func settingsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/settings/telephony/agent.recording" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"scope":"telephony","key":"agent.recording","data":` +
				`{"hideRecordingUi":false,"allowDelete":true,"tag":"a","availableTags":[],"alwaysRecordOutbound":false}}}`))
		case p == "/api/v2/settings/telephony/agent.recording" && r.Method == http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			var req map[string]json.RawMessage
			_ = json.Unmarshal(body, &req)
			_, _ = w.Write([]byte(`{"item":{"scope":"telephony","key":"agent.recording","data":` + string(req["data"]) + `}}`))
		case p == "/api/v2/settings/app/customer.logging" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"scope":"app","key":"customer.logging","data":{"enabled":true}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestSettingsGetUpdate(t *testing.T) {
	srv := settingsServer()
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	rec, err := mgr.Settings.Telephony.AgentRecording.Get(ctx)
	if err != nil || rec.Tag != "a" {
		t.Fatalf("get recording: %v (%+v)", err, rec)
	}

	tag, hide := "b", true
	upd, err := mgr.Settings.Telephony.AgentRecording.Update(ctx, managerapi.SettingsTelephonyAgentRecordingRequestData{
		Tag:             &tag,
		HideRecordingUi: &hide,
	})
	if err != nil || upd.Tag != "b" || !upd.HideRecordingUi {
		t.Fatalf("update recording: %v (%+v)", err, upd)
	}

	cl, err := mgr.Settings.App.CustomerLogging.Get(ctx)
	if err != nil || !cl.Enabled {
		t.Fatalf("get customer.logging: %v (%+v)", err, cl)
	}
}

func TestSettingsErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: APIKey("a", "b")})
	_, err := mgr.Settings.Audit.Default.Get(context.Background())

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
