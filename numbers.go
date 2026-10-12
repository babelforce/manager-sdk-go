package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// NumbersResource is the service-numbers namespace (/api/v2/numbers).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type NumbersResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over service numbers, auto-paginating across pages.
func (r *NumbersResource) List(ctx context.Context, params managerapi.ListServiceNumbersParams) iter.Seq2[managerapi.ServiceNumber, error] {
	return func(yield func(managerapi.ServiceNumber, error) bool) {
		var zero managerapi.ServiceNumber
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListServiceNumbersWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, n := range data.Items {
				if !yield(n, nil) {
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

// ListAll collects every service number into a slice (convenience over List).
func (r *NumbersResource) ListAll(ctx context.Context, params managerapi.ListServiceNumbersParams) ([]managerapi.ServiceNumber, error) {
	var out []managerapi.ServiceNumber
	for n, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

// Get returns a single service number by id.
func (r *NumbersResource) Get(ctx context.Context, id string) (*managerapi.ServiceNumberItemResponse, error) {
	resp, err := r.gc.GetServiceNumberWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a service number from a free-form body (the API accepts an arbitrary object).
func (r *NumbersResource) Update(ctx context.Context, id string, body map[string]any) (*managerapi.ServiceNumberItemResponse, error) {
	resp, err := r.gc.UpdateServiceNumberWithResponse(ctx, id, managerapi.UpdateServiceNumberJSONRequestBody(body))
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddTags adds tags to a service number and returns the updated number.
func (r *NumbersResource) AddTags(ctx context.Context, id string, tags []managerapi.Tag) (*managerapi.ServiceNumberItemResponse, error) {
	resp, err := r.gc.AddTagsToNumberWithResponse(ctx, id, managerapi.AddNumberTagsRequest{Tags: tags})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
