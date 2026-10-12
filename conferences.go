package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// ConferencesResource is the conferences namespace (/api/v2/conferences).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type ConferencesResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over conferences, auto-paginating across pages.
func (r *ConferencesResource) List(ctx context.Context, params managerapi.ListConferencesParams) iter.Seq2[managerapi.Conference, error] {
	return func(yield func(managerapi.Conference, error) bool) {
		var zero managerapi.Conference
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListConferencesWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// ListAll collects every conference into a slice (convenience over List).
func (r *ConferencesResource) ListAll(ctx context.Context, params managerapi.ListConferencesParams) ([]managerapi.Conference, error) {
	var out []managerapi.Conference
	for c, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Get returns a single conference by id.
func (r *ConferencesResource) Get(ctx context.Context, id string) (*managerapi.ConferenceItemResponse, error) {
	resp, err := r.gc.GetConferenceWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
