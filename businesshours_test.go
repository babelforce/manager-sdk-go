package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func businessHoursServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case p == "/api/v2/business-hours/bulk" && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"business hours deleted","success":true}`))
		case p == "/api/v2/business-hours/bulk" && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `","name":"Weekdays"}],"success":true}`))
		case strings.HasSuffix(p, "/ranges") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","name":"Weekdays"},"success":true}`))
		case strings.HasSuffix(p, "/ranges") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidB + `"}],"success":true}`))
		case strings.Contains(p, "/ranges/") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidB + `"},"success":true}`))
		case strings.Contains(p, "/ranges/") && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"range removed","success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
}

func TestBusinessHoursRanges(t *testing.T) {
	srv := businessHoursServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	added, err := mgr.BusinessHours.AddRanges(ctx, uuidA, managerapi.BusinessHourRangesBody{})
	if err != nil {
		t.Fatalf("add ranges: %v", err)
	}
	if added.Item.Name != "Weekdays" {
		t.Fatalf("expected Weekdays, got %q", added.Item.Name)
	}

	list, err := mgr.BusinessHours.ListRanges(ctx, uuidA)
	if err != nil {
		t.Fatalf("list ranges: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Id.String() != uuidB {
		t.Fatalf("expected 1 range %s, got %+v", uuidB, list.Items)
	}

	got, err := mgr.BusinessHours.GetRange(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("get range: %v", err)
	}
	if got.Item.Id.String() != uuidB {
		t.Fatalf("expected range %s, got %s", uuidB, got.Item.Id)
	}

	removed, err := mgr.BusinessHours.RemoveRange(ctx, uuidA, uuidB)
	if err != nil {
		t.Fatalf("remove range: %v", err)
	}
	if removed.Message == nil || *removed.Message != "range removed" {
		t.Fatalf("expected range removed message, got %+v", removed)
	}
}

func TestBusinessHoursBulk(t *testing.T) {
	srv := businessHoursServer()
	defer srv.Close()
	mgr, _ := Connect(context.Background(), Options{BaseURL: srv.URL, Auth: Bearer("TEST")})
	ctx := context.Background()

	deleted, err := mgr.BusinessHours.BulkDelete(ctx, []string{uuidA, uuidB})
	if err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if deleted.Message == nil || *deleted.Message != "business hours deleted" {
		t.Fatalf("expected business hours deleted message, got %+v", deleted)
	}

	updated, err := mgr.BusinessHours.BulkUpdate(ctx, managerapi.BusinessHourBulkUpdateRequest{})
	if err != nil {
		t.Fatalf("bulk update: %v", err)
	}
	if len(updated.Items) != 1 || updated.Items[0].Name != "Weekdays" {
		t.Fatalf("expected 1 updated business hour Weekdays, got %+v", updated.Items)
	}
}
