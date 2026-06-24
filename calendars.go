package manager

import (
	"context"
	"iter"

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

// GetDates returns a calendar's dates as the raw response body.
func (r *CalendarsResource) GetDates(ctx context.Context, id string) ([]byte, error) {
	resp, err := r.gc.GetCalenderDatesWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// AddDate adds a date to a calendar.
func (r *CalendarsResource) AddDate(ctx context.Context, id string) error {
	resp, err := r.gc.AddCalendarDateWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
