package manager

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"time"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// ApplicationView holds the fields every IVR application variant shares, regardless of its module.
// managerapi.Application is a oneOf union (a distinct shape per module), so it carries no directly
// addressable fields; use ApplicationViewOf to read the common ones. For module-specific fields
// (routings, settings, …) use the generated app.As<Module>Application() accessors or
// app.ValueByDiscriminator().
type ApplicationView struct {
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	Module      string           `json:"module"`
	Enabled     bool             `json:"enabled"`
	DateCreated time.Time        `json:"dateCreated"`
	LastUpdated time.Time        `json:"lastUpdated"`
	Tags        []managerapi.Tag `json:"tags"`
}

// ApplicationViewOf extracts the fields common to every Application variant. It also works on the
// Application returned inside ApplicationItemResponse.Item (Get/Create/Update).
func ApplicationViewOf(app managerapi.Application) (ApplicationView, error) {
	raw, err := app.MarshalJSON()
	if err != nil {
		return ApplicationView{}, err
	}
	var v ApplicationView
	if err := json.Unmarshal(raw, &v); err != nil {
		return ApplicationView{}, err
	}
	return v, nil
}

// ApplicationsResource is the application (IVR) management namespace (/api/v2/applications).
type ApplicationsResource struct {
	gc *managerapi.ClientWithResponses
	// Actions is the per-application actions (local automations) sub-namespace
	// (/api/v2/applications/{applicationId}/actions).
	Actions *AppActionsResource
}

// ListApplicationsQuery filters an application listing.
type ListApplicationsQuery struct {
	// PageSize is the page size (the API's max). Zero uses the server default.
	PageSize int
}

// List returns an iterator over applications, auto-paginating across pages.
//
//	for app, err := range mgr.Applications.List(ctx, manager.ListApplicationsQuery{}) {
//	    if err != nil { return err }
//	    v, _ := manager.ApplicationViewOf(app)
//	    fmt.Println(v.Id, v.Name, v.Module)
//	}
func (r *ApplicationsResource) List(ctx context.Context, q ListApplicationsQuery) iter.Seq2[managerapi.Application, error] {
	return func(yield func(managerapi.Application, error) bool) {
		var zero managerapi.Application

		params := &managerapi.ListApplicationsParams{}
		if q.PageSize > 0 {
			params.Max = &q.PageSize
		}

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListApplicationsWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, a := range data.Items {
				if !yield(a, nil) {
					return
				}
			}
			// Pagination is optional on this envelope: a missing block means a single page.
			if data.Pagination == nil || len(data.Items) == 0 || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
		}
	}
}

// ListAll collects every application into a slice (convenience over List).
func (r *ApplicationsResource) ListAll(ctx context.Context, q ListApplicationsQuery) ([]managerapi.Application, error) {
	var apps []managerapi.Application
	for a, err := range r.List(ctx, q) {
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

// Create creates an application.
func (r *ApplicationsResource) Create(ctx context.Context, body managerapi.ApplicationCreateBody) (*managerapi.ApplicationItemResponse, error) {
	resp, err := r.gc.CreateApplicationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns an application by id.
func (r *ApplicationsResource) Get(ctx context.Context, id string) (*managerapi.ApplicationItemResponse, error) {
	resp, err := r.gc.GetApplicationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates an application.
func (r *ApplicationsResource) Update(ctx context.Context, id string, body managerapi.ApplicationUpdateBody) (*managerapi.ApplicationItemResponse, error) {
	resp, err := r.gc.UpdateApplicationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes an application by id.
func (r *ApplicationsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteApplicationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// DeleteMany bulk-deletes applications by id.
func (r *ApplicationsResource) DeleteMany(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DeleteManyApplicationsWithResponse(ctx, managerapi.DeleteManyApplicationsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListModules lists the available IVR modules (the building blocks of applications).
func (r *ApplicationsResource) ListModules(ctx context.Context) (*managerapi.ListModulesResponse, error) {
	resp, err := r.gc.ListModulesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Dispatch dispatches the local automations configured at a position in an application. The body is
// optional: pass nil to send no request payload, or a non-nil *LocalAutomationDispatch to send one.
func (r *ApplicationsResource) Dispatch(ctx context.Context, id, position string, async bool, body *managerapi.LocalAutomationDispatch) (*managerapi.DispatchLocalAutomationsResponse, error) {
	params := &managerapi.DispatchLocalAutomationsParams{Async: async}
	if body == nil {
		// No body: route through the WithBody variant and strip the Content-Type header the generator
		// adds unconditionally, so the wire request carries no payload (parity with the TS SDK).
		stripContentType := func(_ context.Context, req *http.Request) error {
			req.Header.Del("Content-Type")
			return nil
		}
		resp, err := r.gc.DispatchLocalAutomationsWithBodyWithResponse(ctx, id, position, params, "", http.NoBody, stripContentType)
		if err != nil {
			return nil, err
		}
		return result(resp.JSON200, resp.HTTPResponse, resp.Body)
	}
	resp, err := r.gc.DispatchLocalAutomationsWithResponse(ctx, id, position, params, *body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AppActionsResource is the per-application actions (local automations) namespace
// (/api/v2/applications/{applicationId}/actions).
type AppActionsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over an application's actions, auto-paginating across pages.
func (r *AppActionsResource) List(ctx context.Context, applicationID string, pageSize int) iter.Seq2[managerapi.LocalAutomation, error] {
	return func(yield func(managerapi.LocalAutomation, error) bool) {
		var zero managerapi.LocalAutomation

		params := &managerapi.ListLocalAutomationsParams{}
		if pageSize > 0 {
			params.Max = &pageSize
		}

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListLocalAutomationsWithResponse(ctx, applicationID, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, a := range data.Items {
				if !yield(a, nil) {
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

// ListAll collects every action of an application into a slice (convenience over List).
func (r *AppActionsResource) ListAll(ctx context.Context, applicationID string, pageSize int) ([]managerapi.LocalAutomation, error) {
	var actions []managerapi.LocalAutomation
	for a, err := range r.List(ctx, applicationID, pageSize) {
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

// Create creates an action in an application.
func (r *AppActionsResource) Create(ctx context.Context, applicationID string, body managerapi.RestCreateLocalAutomation) (*managerapi.LocalAutomationItemResponse, error) {
	resp, err := r.gc.CreateLocalAutomationWithResponse(ctx, applicationID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns one of an application's actions by id.
func (r *AppActionsResource) Get(ctx context.Context, applicationID, id string) (*managerapi.LocalAutomationItemResponse, error) {
	resp, err := r.gc.GetLocalAutomationWithResponse(ctx, applicationID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates one of an application's actions.
func (r *AppActionsResource) Update(ctx context.Context, applicationID, id string, body managerapi.RestUpdateLocalAutomation) (*managerapi.LocalAutomationItemResponse, error) {
	resp, err := r.gc.UpdateLocalAutomationWithResponse(ctx, applicationID, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes one of an application's actions.
func (r *AppActionsResource) Delete(ctx context.Context, applicationID, id string) error {
	resp, err := r.gc.DeleteLocalAutomationWithResponse(ctx, applicationID, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
