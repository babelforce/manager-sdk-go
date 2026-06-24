package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	taskautomationapi "github.com/babelforce/manager-sdk-go/gen/taskautomation"
)

func TestTaskConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case p == "/api/v3/tasks/scripts/javascript" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/scripts/javascript" && m == http.MethodGet:
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/scripts/javascript/c1" && m == http.MethodPut:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/scripts/javascript/c1":
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/configurations/secrets":
			_, _ = w.Write([]byte(`["pfx"]`))
		case p == "/api/v3/tasks/configurations/secrets/pfx" && m == http.MethodGet:
			_, _ = w.Write([]byte(`["a"]`))
		case p == "/api/v3/tasks/configurations/secrets/pfx":
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/configurations/selection" && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case p == "/api/v3/tasks/configurations/selection":
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: APIKey("x", "y")})
	js := taskautomationapi.ScriptType("javascript")

	if _, err := mgr.Tasks.Scripts.List(ctx, js); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Scripts.Submit(ctx, js, taskautomationapi.Script{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Scripts.Get(ctx, js, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Scripts.Update(ctx, js, "c1", taskautomationapi.Script{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.Scripts.Delete(ctx, js, "c1"); err != nil {
		t.Fatal(err)
	}

	if _, err := mgr.Tasks.Secrets.ListPrefixes(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.Secrets.ListKeys(ctx, "pfx"); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.Secrets.Create(ctx, "pfx", taskautomationapi.CreateSecretsJSONRequestBody{"token": "x"}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.Secrets.Patch(ctx, "pfx", taskautomationapi.PatchSecretsJSONRequestBody{"token": "y"}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.Secrets.DeleteKeys(ctx, "pfx", taskautomationapi.SecretKeys{"token"}); err != nil {
		t.Fatal(err)
	}

	if _, err := mgr.Tasks.SelectionConfig.Read(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.SelectionConfig.Create(ctx, taskautomationapi.CreateSelectionConfigurationJSONRequestBody{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Tasks.SelectionConfig.Update(ctx, taskautomationapi.UpdateSelectionConfigurationJSONRequestBody{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tasks.SelectionConfig.Delete(ctx); err != nil {
		t.Fatal(err)
	}
}
