package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// RoutingResource is the routing-rules namespace (/api/v2/routings).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type RoutingResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over routings, auto-paginating across pages.
func (r *RoutingResource) List(ctx context.Context, params managerapi.ListRoutingsParams) iter.Seq2[managerapi.Routing, error] {
	return func(yield func(managerapi.Routing, error) bool) {
		var zero managerapi.Routing
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListRoutingsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, x := range data.Items {
				if !yield(x, nil) {
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

// ListAll collects every routing into a slice (convenience over List).
func (r *RoutingResource) ListAll(ctx context.Context, params managerapi.ListRoutingsParams) ([]managerapi.Routing, error) {
	var out []managerapi.Routing
	for x, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

// Create creates a routing.
func (r *RoutingResource) Create(ctx context.Context, body managerapi.RestCreateRouting) (*managerapi.RoutingItemResponse, error) {
	resp, err := r.gc.CreateRoutingWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a routing by id.
func (r *RoutingResource) Get(ctx context.Context, id string) (*managerapi.RoutingItemResponse, error) {
	resp, err := r.gc.GetRoutingWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a routing.
func (r *RoutingResource) Update(ctx context.Context, id string, body managerapi.RestUpdateRouting) (*managerapi.RoutingItemResponse, error) {
	resp, err := r.gc.UpdateRoutingWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a routing.
func (r *RoutingResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteRoutingWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
