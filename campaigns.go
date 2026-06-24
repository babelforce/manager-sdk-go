package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// CampaignsResource is the outbound-campaigns namespace (/api/v2/outbound/campaigns).
type CampaignsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns all outbound campaigns.
func (r *CampaignsResource) List(ctx context.Context) ([]managerapi.OutboundCampaign, error) {
	resp, err := r.gc.ListCampaignsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Create creates a campaign.
func (r *CampaignsResource) Create(ctx context.Context, body managerapi.CreateCampaignRequest) (*managerapi.OutboundCampaignItemResponse, error) {
	resp, err := r.gc.CreateCampaignWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a campaign by id.
func (r *CampaignsResource) Get(ctx context.Context, id string) (*managerapi.OutboundCampaignItemResponse, error) {
	resp, err := r.gc.GetCampaignWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a campaign.
func (r *CampaignsResource) Update(ctx context.Context, id string, body managerapi.UpdateCampaignRequest) (*managerapi.OutboundCampaignItemResponse, error) {
	resp, err := r.gc.UpdateCampaignWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Delete deletes a campaign.
func (r *CampaignsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteCampaignWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
