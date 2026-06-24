package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// TriggersResource is the workflow-triggers namespace (/api/v2/triggers).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type TriggersResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over triggers, auto-paginating across pages.
func (r *TriggersResource) List(ctx context.Context, params managerapi.ListTriggersParams) iter.Seq2[managerapi.Trigger, error] {
	return func(yield func(managerapi.Trigger, error) bool) {
		var zero managerapi.Trigger
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListTriggersWithResponse(ctx, &p)
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

// ListAll collects every trigger into a slice (convenience over List).
func (r *TriggersResource) ListAll(ctx context.Context, params managerapi.ListTriggersParams) ([]managerapi.Trigger, error) {
	var out []managerapi.Trigger
	for x, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

// Create creates a trigger.
func (r *TriggersResource) Create(ctx context.Context, body managerapi.RestCreateTrigger) (*managerapi.TriggerItemResponse, error) {
	resp, err := r.gc.CreateTriggerWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a trigger by id.
func (r *TriggersResource) Get(ctx context.Context, id string) (*managerapi.TriggerItemResponse, error) {
	resp, err := r.gc.GetTriggerWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a trigger.
func (r *TriggersResource) Update(ctx context.Context, id string, body managerapi.RestUpdateTrigger) (*managerapi.TriggerItemResponse, error) {
	resp, err := r.gc.UpdateTriggerWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a trigger.
func (r *TriggersResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteTriggerWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Clone clones a trigger and returns the new trigger.
func (r *TriggersResource) Clone(ctx context.Context, id string) (*managerapi.TriggerItemResponse, error) {
	resp, err := r.gc.CloneTriggerWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Test tests trigger conditions against a sample payload. testMode runs without side effects.
func (r *TriggersResource) Test(ctx context.Context, body managerapi.TestTriggersRequest, testMode bool) (*managerapi.TestTriggersResponse, error) {
	resp, err := r.gc.TestTriggersWithResponse(ctx, &managerapi.TestTriggersParams{TestMode: testMode}, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
