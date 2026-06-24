package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// AutomationsResource is the global-automations namespace (event triggers, /api/v2/events/triggers).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type AutomationsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over global automations, auto-paginating across pages.
func (r *AutomationsResource) List(ctx context.Context, params managerapi.ListGlobalAutomationsParams) iter.Seq2[managerapi.GlobalAutomation, error] {
	return func(yield func(managerapi.GlobalAutomation, error) bool) {
		var zero managerapi.GlobalAutomation
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListGlobalAutomationsWithResponse(ctx, &p)
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

// ListAll collects every global automation into a slice (convenience over List).
func (r *AutomationsResource) ListAll(ctx context.Context, params managerapi.ListGlobalAutomationsParams) ([]managerapi.GlobalAutomation, error) {
	var out []managerapi.GlobalAutomation
	for x, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

// Create creates a global automation.
func (r *AutomationsResource) Create(ctx context.Context, body managerapi.RestCreateGlobalAutomation) (*managerapi.GlobalAutomationItemResponse, error) {
	resp, err := r.gc.CreateGlobalAutomationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a global automation by id.
func (r *AutomationsResource) Get(ctx context.Context, id string) (*managerapi.GlobalAutomationItemResponse, error) {
	resp, err := r.gc.GetGlobalAutomationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a global automation.
func (r *AutomationsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateGlobalAutomation) (*managerapi.GlobalAutomationItemResponse, error) {
	resp, err := r.gc.UpdateGlobalAutomationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a global automation.
func (r *AutomationsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteGlobalAutomationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
