package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func TestBusinessHoursAndCalendars(t *testing.T) {
	item := `{"item":{"id":"` + uuidA + `"},"success":true}`
	list := `{"items":[{"id":"` + uuidA + `"}],"pagination":{"pages":1,"current":1,"total":1,"max":50}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p, m := r.URL.Path, r.Method
		switch {
		case (p == "/api/v2/business-hours" || p == "/api/v2/calendars") && m == http.MethodGet:
			_, _ = w.Write([]byte(list))
		case (p == "/api/v2/business-hours" || p == "/api/v2/calendars") && m == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(item))
		case strings.HasSuffix(p, "/dates"):
			_, _ = w.Write([]byte(`{"message":"ok","success":true}`))
		case m == http.MethodDelete:
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			_, _ = w.Write([]byte(item))
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	if bs, err := mgr.BusinessHours.ListAll(ctx, managerapi.ListBusinessHoursParams{}); err != nil || len(bs) != 1 {
		t.Fatalf("business hours list: %v len=%d", err, len(bs))
	}
	if _, err := mgr.BusinessHours.Create(ctx, managerapi.RestCreateBusinessHour{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.BusinessHours.Get(ctx, "b1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.BusinessHours.Update(ctx, "b1", managerapi.RestUpdateBusinessHour{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.BusinessHours.Delete(ctx, "b1"); err != nil {
		t.Fatal(err)
	}

	if cs, err := mgr.Calendars.ListAll(ctx, managerapi.ListCalendarsParams{}); err != nil || len(cs) != 1 {
		t.Fatalf("calendars list: %v len=%d", err, len(cs))
	}
	if _, err := mgr.Calendars.Create(ctx, managerapi.RestCreateCalendar{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calendars.Get(ctx, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calendars.GetDates(ctx, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Calendars.AddDate(ctx, "c1", managerapi.AddCalendarDateJSONRequestBody{}); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Calendars.Delete(ctx, "c1"); err != nil {
		t.Fatal(err)
	}
}
