package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func calendarsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/calendars/bulk" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"calendars deleted","success":true}`))
		case p == "/api/v2/calendars/bulk" && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `","name":"Holidays"}],"success":true}`))
		case p == "/api/v2/calendars/test" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"open":true},"success":true}`))
		case strings.Contains(p, "/dates/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `","label":"New Year"},"success":true}`))
		case strings.Contains(p, "/dates/") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `","label":"New Year (updated)"},"success":true}`))
		case strings.Contains(p, "/dates/") && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"date removed","success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestCalendarDates(t *testing.T) {
	srv := calendarsServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	got, err := mgr.Calendars.GetDate(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("get date: %v", err)
	}
	if got.Item.Label != "New Year" {
		t.Fatalf("expected New Year, got %q", got.Item.Label)
	}

	updated, err := mgr.Calendars.UpdateDate(ctx, uuidA, uuidB, managerapi.CalendarDateBody{Label: "New Year (updated)"})
	if err != nil {
		t.Fatalf("update date: %v", err)
	}
	if updated.Item.Label != "New Year (updated)" {
		t.Fatalf("expected updated label, got %q", updated.Item.Label)
	}

	removed, err := mgr.Calendars.RemoveDate(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("remove date: %v", err)
	}
	if removed.Message == nil || *removed.Message != "date removed" {
		t.Fatalf("expected date removed message, got %+v", removed)
	}
}

func TestCalendarTestDateAndBulk(t *testing.T) {
	srv := calendarsServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	date := "2026-01-01"
	tested, err := mgr.Calendars.TestDate(ctx, &date)
	if err != nil {
		t.Fatalf("test date: %v", err)
	}
	if open, ok := (*tested.Item)["open"].(bool); !ok || !open {
		t.Fatalf("expected open=true, got %+v", tested.Item)
	}

	deleted, err := mgr.Calendars.BulkDelete(ctx, []string{uuidA, uuidB})
	if err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if deleted.Message == nil || *deleted.Message != "calendars deleted" {
		t.Fatalf("expected calendars deleted message, got %+v", deleted)
	}

	updated, err := mgr.Calendars.BulkUpdate(ctx, managerapi.CalendarBulkUpdateRequest{})
	if err != nil {
		t.Fatalf("bulk update: %v", err)
	}
	if len(updated.Items) != 1 || updated.Items[0].Name != "Holidays" {
		t.Fatalf("expected 1 updated calendar Holidays, got %+v", updated.Items)
	}
}
