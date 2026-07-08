package manager

import (
	"context"
	"iter"

	"github.com/google/uuid"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// CalendarsResource is the calendars namespace (/api/v2/calendars), with calendar dates.
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type CalendarsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over calendars, auto-paginating across pages.
func (r *CalendarsResource) List(ctx context.Context, params managerapi.ListCalendarsParams) iter.Seq2[managerapi.Calendar, error] {
	return func(yield func(managerapi.Calendar, error) bool) {
		var zero managerapi.Calendar
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListCalendarsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, c := range data.Items {
				if !yield(c, nil) {
					return
				}
			}
			if len(data.Items) == 0 || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
		}
	}
}

// ListAll collects every calendar into a slice (convenience over List).
func (r *CalendarsResource) ListAll(ctx context.Context, params managerapi.ListCalendarsParams) ([]managerapi.Calendar, error) {
	var out []managerapi.Calendar
	for c, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Create creates a calendar.
func (r *CalendarsResource) Create(ctx context.Context, body managerapi.RestCreateCalendar) (*managerapi.CalendarItemResponse, error) {
	resp, err := r.gc.CreateCalendarWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a calendar by id.
func (r *CalendarsResource) Get(ctx context.Context, id string) (*managerapi.CalendarItemResponse, error) {
	resp, err := r.gc.GetCalendarWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a calendar.
func (r *CalendarsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateCalendar) (*managerapi.CalendarItemResponse, error) {
	resp, err := r.gc.UpdateCalendarWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a calendar.
func (r *CalendarsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteCalendarWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// GetDates returns a calendar's dates.
func (r *CalendarsResource) GetDates(ctx context.Context, id string) ([]managerapi.CalendarDate, error) {
	resp, err := r.gc.GetCalenderDatesWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// AddDate adds a date to a calendar.
func (r *CalendarsResource) AddDate(ctx context.Context, id string, body managerapi.AddCalendarDateJSONRequestBody) (*managerapi.CalendarDateItemResponse, error) {
	resp, err := r.gc.AddCalendarDateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetDate returns a single calendar date by id.
func (r *CalendarsResource) GetDate(ctx context.Context, id, dateID string) (*managerapi.CalendarDateItemResponse, error) {
	resp, err := r.gc.GetCalendarDateWithResponse(ctx, id, dateID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateDate updates a calendar date by id.
func (r *CalendarsResource) UpdateDate(ctx context.Context, id, dateID string, body managerapi.CalendarDateBody) (*managerapi.CalendarDateItemResponse, error) {
	resp, err := r.gc.UpdateCalendarDateWithResponse(ctx, id, dateID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// RemoveDate removes a date from a calendar.
func (r *CalendarsResource) RemoveDate(ctx context.Context, id, dateID string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.RemoveCalendarDateWithResponse(ctx, id, dateID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TestDate evaluates the calendars against a date (defaults to now when date is nil).
func (r *CalendarsResource) TestDate(ctx context.Context, date *string) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.TestCalendarDateWithResponse(ctx, &managerapi.TestCalendarDateParams{Date: date})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkDelete deletes the given calendars by id.
func (r *CalendarsResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	uuids := make([]managerapi.ObjectUuid, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsed)
	}
	resp, err := r.gc.BulkDeleteCalendarsWithResponse(ctx, managerapi.BulkDeleteCalendarsJSONRequestBody{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkUpdate applies the given updates to multiple calendars at once.
func (r *CalendarsResource) BulkUpdate(ctx context.Context, body managerapi.CalendarBulkUpdateRequest) (*managerapi.CalendarListResponse, error) {
	resp, err := r.gc.BulkUpdateCalendarsWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
