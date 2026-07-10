package manager

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func filePage(id, filename string, current, pages int) string {
	return `{"items":[{"id":"` + id + `","url":"https://files/x","path":"/x","filename":"` + filename + `"}],` +
		`"pagination":{"pages":` + itoa(pages) + `,"current":` + itoa(current) + `,"total":` + itoa(pages) + `,"max":1}}`
}

func filesServer(t *testing.T, gotBulkBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer TEST" {
			t.Errorf("missing auth header on %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/files" && r.Method == http.MethodGet:
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(filePage(uuidB, "b.wav", 2, 2)))
			} else {
				_, _ = w.Write([]byte(filePage(uuidA, "a.wav", 1, 2)))
			}
		case p == "/api/v2/files/prompts/bulk" && r.Method == http.MethodDelete:
			body, _ := io.ReadAll(r.Body)
			if gotBulkBody != nil {
				*gotBulkBody = string(body)
			}
			_, _ = w.Write([]byte(`{"message":"deleted","success":true}`))
		case strings.HasPrefix(p, "/api/v2/files/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","url":"https://files/x","path":"/x","filename":"a.wav"},"success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestFilesListAllAndGet(t *testing.T) {
	srv := filesServer(t, nil)
	defer srv.Close()
	mgr, err := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	files, err := mgr.Files.ListAll(ctx, managerapi.ListFilesParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Filename != "a.wav" || files[1].Filename != "b.wav" {
		t.Fatalf("expected 2 auto-paged files, got %+v", files)
	}

	got, err := mgr.Files.Get(ctx, uuidA)
	if err != nil {
		t.Fatal(err)
	}
	if got.Item.Filename != "a.wav" {
		t.Fatalf("get: %q", got.Item.Filename)
	}
}

func TestFilesBulkDeleteSendsIds(t *testing.T) {
	var bulkBody string
	srv := filesServer(t, &bulkBody)
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if _, err := mgr.Files.BulkDelete(context.Background(), []string{uuidA}); err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if !strings.Contains(bulkBody, `"ids":["`+uuidA+`"]`) {
		t.Fatalf("expected ids body, got %q", bulkBody)
	}
}

func TestFilesBulkDeleteRejectsMalformedID(t *testing.T) {
	mgr, _ := Connect(context.Background(), Options{BaseURL: "https://example.test", Auth: Bearer("TEST")})
	if _, err := mgr.Files.BulkDelete(context.Background(), []string{"not-a-uuid"}); err == nil {
		t.Fatal("expected an error for a malformed id")
	}
}

func TestFilesErrorIsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"FORBIDDEN","message":"nope"}`))
	}))
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	_, err := mgr.Files.Get(context.Background(), uuidA)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403 APIError, got %v", err)
	}
}
