package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// MetricsResource is the metrics namespace (/api/v2/metrics).
type MetricsResource struct {
	gc *managerapi.ClientWithResponses
}

// ListIds lists the available metric ids.
func (r *MetricsResource) ListIds(ctx context.Context) (*managerapi.MetricIdItemsResponse, error) {
	resp, err := r.gc.ListMetricIdsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Get returns a metric's current value by id.
func (r *MetricsResource) Get(ctx context.Context, id string) (*managerapi.MetricResponse, error) {
	resp, err := r.gc.GetMetricWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Definitions returns the definitions for every available metric
// (GET /api/v2/metrics/describe).
func (r *MetricsResource) Definitions(ctx context.Context) (*managerapi.MetricDefinitionListResponse, error) {
	resp, err := r.gc.GetAllMetricDefinitionsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Describe returns a metric's definition by id.
func (r *MetricsResource) Describe(ctx context.Context, id string) (*managerapi.MetricDefinitionItemResponse, error) {
	resp, err := r.gc.GetMetricDefinitionWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
