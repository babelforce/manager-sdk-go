package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// SmsResource is the SMS-records namespace (/api/v2/sms).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type SmsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over SMS records, auto-paginating across pages.
func (r *SmsResource) List(ctx context.Context, params managerapi.ListSmssParams) iter.Seq2[managerapi.Sms, error] {
	return func(yield func(managerapi.Sms, error) bool) {
		var zero managerapi.Sms
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListSmssWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, s := range data.Items {
				if !yield(s, nil) {
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

// ListAll collects every SMS record into a slice (convenience over List).
func (r *SmsResource) ListAll(ctx context.Context, params managerapi.ListSmssParams) ([]managerapi.Sms, error) {
	var out []managerapi.Sms
	for s, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// Get returns a single SMS record by id.
func (r *SmsResource) Get(ctx context.Context, id string) (*managerapi.SmsItemResponse, error) {
	resp, err := r.gc.GetSmsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
