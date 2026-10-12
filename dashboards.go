package manager

import (
	"context"
	"iter"

	openapi_types "github.com/oapi-codegen/runtime/types"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// DashboardsResource is the reporting-dashboards namespace (/api/v2/dashboards).
type DashboardsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over dashboards, auto-paginating across pages. The Page field of
// params is managed by the iterator; any other filters (Q, Uuid, Sort, Order, Max) are honoured.
//
//	for dashboard, err := range mgr.Dashboards.List(ctx, managerapi.ListDashboardsParams{}) {
//	    if err != nil { return err }
//	    fmt.Println(dashboard.Name)
//	}
func (r *DashboardsResource) List(ctx context.Context, params managerapi.ListDashboardsParams) iter.Seq2[managerapi.Dashboard, error] {
	return func(yield func(managerapi.Dashboard, error) bool) {
		var zero managerapi.Dashboard

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListDashboardsWithResponse(ctx, &params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, d := range data.Items {
				if !yield(d, nil) {
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

// ListAll collects every dashboard into a slice (convenience over List).
func (r *DashboardsResource) ListAll(ctx context.Context, params managerapi.ListDashboardsParams) ([]managerapi.Dashboard, error) {
	var dashboards []managerapi.Dashboard
	for d, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, d)
	}
	return dashboards, nil
}

// Create creates a dashboard.
func (r *DashboardsResource) Create(ctx context.Context, body managerapi.DashboardCreateBody) (*managerapi.DashboardItemResponse, error) {
	resp, err := r.gc.CreateDashboardWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a dashboard by id.
func (r *DashboardsResource) Get(ctx context.Context, id string) (*managerapi.DashboardItemResponse, error) {
	resp, err := r.gc.GetDashboardWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a dashboard.
func (r *DashboardsResource) Update(ctx context.Context, id string, body managerapi.DashboardUpdateBody) (*managerapi.DashboardItemResponse, error) {
	resp, err := r.gc.UpdateDashboardWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a dashboard by id.
func (r *DashboardsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteDashboardWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// ListUsers lists the users allowed to access a dashboard.
func (r *DashboardsResource) ListUsers(ctx context.Context, id string) (*managerapi.DashboardUsersResponse, error) {
	resp, err := r.gc.ListDashboardUsersWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddUser grants a user (by email) access to a dashboard.
func (r *DashboardsResource) AddUser(ctx context.Context, id, email string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.AddDashboardUserWithResponse(ctx, id, managerapi.DashboardUserAddRequest{Email: openapi_types.Email(email)})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// RemoveUser revokes a user's (by id) access to a dashboard.
func (r *DashboardsResource) RemoveUser(ctx context.Context, id, userId string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.RemoveDashboardUserWithResponse(ctx, id, userId)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
