package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// BabeldeskResource is the babeldesk-dashboards namespace (/api/v2/babeldesk/dashboards), with a
// nested Widgets sub-resource.
type BabeldeskResource struct {
	gc *managerapi.ClientWithResponses
	// Widgets is the babeldesk-widgets sub-namespace (/api/v2/babeldesk/widgets).
	Widgets *BabeldeskWidgetsResource
}

// List returns an iterator over babeldesk dashboards, auto-paginating across pages.
func (r *BabeldeskResource) List(ctx context.Context, params managerapi.ListBabeldesksParams) iter.Seq2[managerapi.Babeldesk, error] {
	return func(yield func(managerapi.Babeldesk, error) bool) {
		var zero managerapi.Babeldesk
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListBabeldesksWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, b := range data.Items {
				if !yield(b, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// ListAll collects every dashboard into a slice (convenience over List).
func (r *BabeldeskResource) ListAll(ctx context.Context, params managerapi.ListBabeldesksParams) ([]managerapi.Babeldesk, error) {
	var out []managerapi.Babeldesk
	for b, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

// Create creates a dashboard.
func (r *BabeldeskResource) Create(ctx context.Context, body managerapi.RestCreateBabeldesk) (*managerapi.BabeldeskItemResponse, error) {
	resp, err := r.gc.CreateBabeldeskWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a dashboard by id.
func (r *BabeldeskResource) Get(ctx context.Context, id string) (*managerapi.BabeldeskItemResponse, error) {
	resp, err := r.gc.GetBabeldeskWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a dashboard.
func (r *BabeldeskResource) Update(ctx context.Context, id string, body managerapi.RestUpdateBabeldesk) (*managerapi.BabeldeskItemResponse, error) {
	resp, err := r.gc.UpdateBabeldeskWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a dashboard.
func (r *BabeldeskResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteBabeldeskWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// WidgetSettings returns the UI feature flags and type-specific settings for a widget type
// (GET /api/v2/widget/{type}/settings).
func (r *BabeldeskResource) WidgetSettings(ctx context.Context, widgetType string) (*managerapi.WidgetSettingsResponse, error) {
	resp, err := r.gc.GetWidgetSettingsWithResponse(ctx, widgetType)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BabeldeskWidgetsResource is the babeldesk-widgets namespace (/api/v2/babeldesk/widgets).
type BabeldeskWidgetsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over widgets, auto-paginating across pages.
func (r *BabeldeskWidgetsResource) List(ctx context.Context, params managerapi.ListBabeldeskWidgetsParams) iter.Seq2[managerapi.BabeldeskWidget, error] {
	return func(yield func(managerapi.BabeldeskWidget, error) bool) {
		var zero managerapi.BabeldeskWidget
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListBabeldeskWidgetsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, wdg := range data.Items {
				if !yield(wdg, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// ListAll collects every widget into a slice (convenience over List).
func (r *BabeldeskWidgetsResource) ListAll(ctx context.Context, params managerapi.ListBabeldeskWidgetsParams) ([]managerapi.BabeldeskWidget, error) {
	var out []managerapi.BabeldeskWidget
	for wdg, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, wdg)
	}
	return out, nil
}

// Create creates a widget.
func (r *BabeldeskWidgetsResource) Create(ctx context.Context, body managerapi.RestCreateBabeldeskWidget) (*managerapi.BabeldeskWidgetItemResponse, error) {
	resp, err := r.gc.CreateBabeldeskWidgetWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a widget by id.
func (r *BabeldeskWidgetsResource) Get(ctx context.Context, id string) (*managerapi.BabeldeskWidgetItemResponse, error) {
	resp, err := r.gc.GetBabeldeskWidgetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a widget.
func (r *BabeldeskWidgetsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateBabeldeskWidget) (*managerapi.BabeldeskWidgetItemResponse, error) {
	resp, err := r.gc.UpdateBabeldeskWidgetWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a widget.
func (r *BabeldeskWidgetsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteBabeldeskWidgetWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
