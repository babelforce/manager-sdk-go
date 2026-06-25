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
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
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

	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Settings.Audit.Default.Get(context.Background())

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}

// TestSettingsGenericMethods covers the collection-wide accessors (ListAll, ListInScope, Clear,
// ClearInScope, ClearAll): each is pinned to the method+path its generated op hits and its decoded
// result.
func TestSettingsGenericMethods(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v2/settings" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"scope":"app","key":"customer.logging","data":{"enabled":true}}],"success":true}`))
		case r.URL.Path == "/api/v2/settings" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"items":[],"success":true}`))
		case r.URL.Path == "/api/v2/settings/app" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"scope":"app","key":"customer.logging","data":{"enabled":true}}],"success":true}`))
		case r.URL.Path == "/api/v2/settings/app" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"items":[],"success":true}`))
		case r.URL.Path == "/api/v2/settings/app/customer.logging" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"item":{"scope":"app","key":"customer.logging","data":{"enabled":false}},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	all, err := mgr.Settings.ListAll(ctx)
	if err != nil || !all.Success || len(all.Items) != 1 {
		t.Fatalf("ListAll: %v (%+v)", err, all)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/settings" {
		t.Fatalf("ListAll hit %s %q, want GET /api/v2/settings", gotMethod, gotPath)
	}

	inScope, err := mgr.Settings.ListInScope(ctx, "app")
	if err != nil || len(inScope.Items) != 1 {
		t.Fatalf("ListInScope: %v (%+v)", err, inScope)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v2/settings/app" {
		t.Fatalf("ListInScope hit %s %q, want GET /api/v2/settings/app", gotMethod, gotPath)
	}

	cleared, err := mgr.Settings.Clear(ctx, "app", "customer.logging")
	if err != nil || !cleared.Success || cleared.Item.Scope != "app" || cleared.Item.Key != "customer.logging" {
		t.Fatalf("Clear: %v (%+v)", err, cleared)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v2/settings/app/customer.logging" {
		t.Fatalf("Clear hit %s %q, want DELETE /api/v2/settings/app/customer.logging", gotMethod, gotPath)
	}

	clearedScope, err := mgr.Settings.ClearInScope(ctx, "app")
	if err != nil || !clearedScope.Success || len(clearedScope.Items) != 0 {
		t.Fatalf("ClearInScope: %v (%+v)", err, clearedScope)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v2/settings/app" {
		t.Fatalf("ClearInScope hit %s %q, want DELETE /api/v2/settings/app", gotMethod, gotPath)
	}

	clearedAll, err := mgr.Settings.ClearAll(ctx)
	if err != nil || !clearedAll.Success || len(clearedAll.Items) != 0 {
		t.Fatalf("ClearAll: %v (%+v)", err, clearedAll)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v2/settings" {
		t.Fatalf("ClearAll hit %s %q, want DELETE /api/v2/settings", gotMethod, gotPath)
	}
}

// TestSettingsAllGroupsWiring pins every settings accessor to the path its generated op hits and the
// scope/key literals its update sends — the safety net for the hand-written wiring (scope/key are
// untyped string literals the compiler cannot check, and only a couple of groups are covered above).
func TestSettingsAllGroupsWiring(t *testing.T) {
	var gotPath, gotScope, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			var env struct {
				Scope string          `json:"scope"`
				Key   string          `json:"key"`
				Data  json.RawMessage `json:"data"`
			}
			_ = json.Unmarshal(body, &env)
			gotScope, gotKey = env.Scope, env.Key
			_, _ = w.Write([]byte(`{"item":{"scope":"` + env.Scope + `","key":"` + env.Key + `","data":` + string(env.Data) + `}}`))
			return
		}
		_, _ = w.Write([]byte(`{"item":{"scope":"x","key":"y","data":{}}}`))
	}))
	defer srv.Close()

	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	cases := []struct {
		name       string
		path       string
		scope, key string
		get        func(context.Context) error
		update     func(context.Context) error
	}{
		{"app.CustomerLogging", "/api/v2/settings/app/customer.logging", "app", "customer.logging",
			func(ctx context.Context) error { _, e := mgr.Settings.App.CustomerLogging.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.App.CustomerLogging.Update(ctx, managerapi.SettingsAppCustomerLoggingRequestData{})
				return e
			}},
		{"app.Conversations", "/api/v2/settings/app/conversations", "app", "conversations",
			func(ctx context.Context) error { _, e := mgr.Settings.App.Conversations.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.App.Conversations.Update(ctx, managerapi.SettingsAppConversationsRequestData{})
				return e
			}},
		{"app.Integrations", "/api/v2/settings/app/integrations", "app", "integrations",
			func(ctx context.Context) error { _, e := mgr.Settings.App.Integrations.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.App.Integrations.Update(ctx, managerapi.SettingsAppIntegrationsRequestData{})
				return e
			}},
		{"app.AgentStatus", "/api/v2/settings/app/agent.status", "app", "agent.status",
			func(ctx context.Context) error { _, e := mgr.Settings.App.AgentStatus.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.App.AgentStatus.Update(ctx, managerapi.SettingsAppAgentStatusRequestData{})
				return e
			}},
		{"telephony.AgentInbound", "/api/v2/settings/telephony/agent.inbound", "telephony", "agent.inbound",
			func(ctx context.Context) error { _, e := mgr.Settings.Telephony.AgentInbound.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Telephony.AgentInbound.Update(ctx, managerapi.SettingsTelephonyAgentInboundRequestData{})
				return e
			}},
		{"telephony.AgentOutbound", "/api/v2/settings/telephony/agent.outbound", "telephony", "agent.outbound",
			func(ctx context.Context) error { _, e := mgr.Settings.Telephony.AgentOutbound.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Telephony.AgentOutbound.Update(ctx, managerapi.SettingsTelephonyAgentOutboundRequestData{})
				return e
			}},
		{"telephony.AgentRecording", "/api/v2/settings/telephony/agent.recording", "telephony", "agent.recording",
			func(ctx context.Context) error { _, e := mgr.Settings.Telephony.AgentRecording.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Telephony.AgentRecording.Update(ctx, managerapi.SettingsTelephonyAgentRecordingRequestData{})
				return e
			}},
		{"telephony.AgentWrapup", "/api/v2/settings/telephony/agent.wrapup", "telephony", "agent.wrapup",
			func(ctx context.Context) error { _, e := mgr.Settings.Telephony.AgentWrapup.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Telephony.AgentWrapup.Update(ctx, managerapi.SettingsTelephonyAgentWrapupRequestData{})
				return e
			}},
		{"telephony.PostCall", "/api/v2/settings/telephony/post-call", "telephony", "post-call",
			func(ctx context.Context) error { _, e := mgr.Settings.Telephony.PostCall.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Telephony.PostCall.Update(ctx, managerapi.SettingsTelephonyPostCallRequestData{})
				return e
			}},
		{"audit.Default", "/api/v2/settings/audit/default", "audit", "default",
			func(ctx context.Context) error { _, e := mgr.Settings.Audit.Default.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Audit.Default.Update(ctx, managerapi.SettingsAuditDefaultRequestData{})
				return e
			}},
		{"ui.I18n", "/api/v2/settings/ui/i18n", "ui", "i18n",
			func(ctx context.Context) error { _, e := mgr.Settings.Ui.I18n.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Ui.I18n.Update(ctx, managerapi.SettingsUiI18nRequestData{})
				return e
			}},
		{"retention.Periods", "/api/v2/settings/retention/periods", "retention", "periods",
			func(ctx context.Context) error { _, e := mgr.Settings.Retention.Periods.Get(ctx); return e },
			func(ctx context.Context) error {
				_, e := mgr.Settings.Retention.Periods.Update(ctx, managerapi.SettingsRetentionPeriodsRequestData{})
				return e
			}},
	}

	if len(cases) != 12 {
		t.Fatalf("expected 12 settings groups, table has %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotScope, gotKey = "", "", ""
			if err := tc.get(ctx); err != nil {
				t.Fatalf("get: %v", err)
			}
			if gotPath != tc.path {
				t.Fatalf("get hit %q, want %q", gotPath, tc.path)
			}
			gotPath, gotScope, gotKey = "", "", ""
			if err := tc.update(ctx); err != nil {
				t.Fatalf("update: %v", err)
			}
			if gotPath != tc.path {
				t.Fatalf("update hit %q, want %q", gotPath, tc.path)
			}
			if gotScope != tc.scope || gotKey != tc.key {
				t.Fatalf("update body scope/key = %q/%q, want %q/%q", gotScope, gotKey, tc.scope, tc.key)
			}
		})
	}
}
