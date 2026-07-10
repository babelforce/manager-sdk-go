package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// SystemResource is the system/reference namespace: health checks, server time,
// timezones, push tokens, tags, and template exports.
type SystemResource struct {
	gc *managerapi.ClientWithResponses
}

// Echo returns the request echoed back (/api/v2/echo).
func (r *SystemResource) Echo(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.EchoWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Ping is a liveness check (/api/v2/ping).
func (r *SystemResource) Ping(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.PingWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ApiStatus returns the API status (/api/v2/status).
func (r *SystemResource) ApiStatus(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.GetApiStatusWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ServerTime returns the current server time (/api/v2/data/time).
func (r *SystemResource) ServerTime(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.GetServerTimeWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Timezones lists the available timezones (/api/v2/data/timezones).
func (r *SystemResource) Timezones(ctx context.Context, params managerapi.ListTimezonesParams) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.ListTimezonesWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// PushToken returns the push token (/api/v2/push-token).
func (r *SystemResource) PushToken(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.GetPushTokenWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Tags lists all tags (/api/v2/tags).
func (r *SystemResource) Tags(ctx context.Context) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.ListTagsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TagsByCategory lists tags within a category (/api/v2/tags/{category}).
func (r *SystemResource) TagsByCategory(ctx context.Context, category string) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.ListTagsByCategoryWithResponse(ctx, category)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ExportTemplates exports configuration templates of the given type
// (/api/v2/templates/export/{type}). The response is a free-form object.
func (r *SystemResource) ExportTemplates(ctx context.Context, templateType string) (map[string]any, error) {
	resp, err := r.gc.ExportTemplatesWithResponse(ctx, managerapi.ExportTemplatesParamsType(templateType))
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return *out, nil
}
