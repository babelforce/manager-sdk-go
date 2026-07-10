package manager

import (
	"context"
	"iter"

	"github.com/google/uuid"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// BusinessHoursResource is the business-hours namespace (/api/v2/business-hours).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type BusinessHoursResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over business-hours definitions, auto-paginating across pages.
func (r *BusinessHoursResource) List(ctx context.Context, params managerapi.ListBusinessHoursParams) iter.Seq2[managerapi.BusinessHour, error] {
	return func(yield func(managerapi.BusinessHour, error) bool) {
		var zero managerapi.BusinessHour
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListBusinessHoursWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// ListAll collects every business-hours definition into a slice (convenience over List).
func (r *BusinessHoursResource) ListAll(ctx context.Context, params managerapi.ListBusinessHoursParams) ([]managerapi.BusinessHour, error) {
	var out []managerapi.BusinessHour
	for b, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

// Create creates a business-hours definition.
func (r *BusinessHoursResource) Create(ctx context.Context, body managerapi.RestCreateBusinessHour) (*managerapi.BusinessHourItemResponse, error) {
	resp, err := r.gc.CreateBusinessHourWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a business-hours definition by id.
func (r *BusinessHoursResource) Get(ctx context.Context, id string) (*managerapi.BusinessHourItemResponse, error) {
	resp, err := r.gc.GetBusinessHourWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a business-hours definition.
func (r *BusinessHoursResource) Update(ctx context.Context, id string, body managerapi.RestUpdateBusinessHour) (*managerapi.BusinessHourItemResponse, error) {
	resp, err := r.gc.UpdateBusinessHourWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a business-hours definition.
func (r *BusinessHoursResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteBusinessHourWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// AddRanges adds weekly time ranges to a business-hours definition.
func (r *BusinessHoursResource) AddRanges(ctx context.Context, id string, body managerapi.BusinessHourRangesBody) (*managerapi.BusinessHourItemResponse, error) {
	resp, err := r.gc.AddBusinessHourRangesWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListRanges returns the weekly time ranges of a business-hours definition.
func (r *BusinessHoursResource) ListRanges(ctx context.Context, id string) (*managerapi.BusinessHourRangeListResponse, error) {
	resp, err := r.gc.ListBusinessHourRangesWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetRange returns a single weekly time range of a business-hours definition by id.
func (r *BusinessHoursResource) GetRange(ctx context.Context, id, rangeID string) (*managerapi.BusinessHourRangeItemResponse, error) {
	resp, err := r.gc.GetBusinessHourRangeWithResponse(ctx, id, rangeID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// RemoveRange removes a weekly time range from a business-hours definition.
func (r *BusinessHoursResource) RemoveRange(ctx context.Context, id, rangeID string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.RemoveBusinessHourRangeWithResponse(ctx, id, rangeID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkDelete deletes the given business-hours definitions by id.
func (r *BusinessHoursResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	uuids := make([]managerapi.ObjectUuid, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsed)
	}
	resp, err := r.gc.BulkDeleteBusinessHoursWithResponse(ctx, managerapi.BulkDeleteBusinessHoursJSONRequestBody{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkUpdate applies the given updates to multiple business-hours definitions at once.
func (r *BusinessHoursResource) BulkUpdate(ctx context.Context, body managerapi.BusinessHourBulkUpdateRequest) (*managerapi.BusinessHourListResponse, error) {
	resp, err := r.gc.BulkUpdateBusinessHoursWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
