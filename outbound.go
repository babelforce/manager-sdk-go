package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// OutboundResource is the outbound dialer-lists namespace (/api/v2/outbound/lists), with leads.
type OutboundResource struct {
	gc *managerapi.ClientWithResponses
}

// Lists returns all outbound lists.
func (r *OutboundResource) Lists(ctx context.Context) ([]managerapi.OutboundList, error) {
	resp, err := r.gc.ListOutboundListsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// CreateList creates an outbound list.
func (r *OutboundResource) CreateList(ctx context.Context, body managerapi.CreateOutboundListRequest) (*managerapi.OutboundListItemResponse, error) {
	resp, err := r.gc.CreateOutboundListWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ClearList removes all leads from an outbound list and returns the (now empty) list.
func (r *OutboundResource) ClearList(ctx context.Context, id string) (*managerapi.OutboundListItemResponse, error) {
	resp, err := r.gc.ClearOutboundListWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddLead adds a lead to an outbound list.
func (r *OutboundResource) AddLead(ctx context.Context, listID string, body managerapi.AddOutboundLeadRequest) (*managerapi.OutboundLeadItemResponse, error) {
	resp, err := r.gc.AddOutboundLeadWithResponse(ctx, listID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateLead updates a lead in an outbound list.
func (r *OutboundResource) UpdateLead(ctx context.Context, listID, leadID string, body managerapi.AddOutboundLeadRequest) (*managerapi.OutboundLeadItemResponse, error) {
	resp, err := r.gc.UpdateOutboundLeadWithResponse(ctx, listID, leadID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeleteLead removes a lead from an outbound list.
func (r *OutboundResource) DeleteLead(ctx context.Context, listID, leadID string) error {
	resp, err := r.gc.DeleteOutboundLeadWithResponse(ctx, listID, leadID)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
