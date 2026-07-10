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

// Attempts lists the dialing attempts for a campaign.
func (r *CampaignsResource) Attempts(ctx context.Context, id string, params *managerapi.ListCampaignAttemptsParams) (*managerapi.PaginatedAttemptResponse, error) {
	resp, err := r.gc.ListCampaignAttemptsWithResponse(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Hopper returns the hopper (next leads to dial) for a campaign.
func (r *CampaignsResource) Hopper(ctx context.Context, id string) (*managerapi.CampaignHopperResponse, error) {
	resp, err := r.gc.GetCampaignHopperWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Leads lists the leads of a campaign.
func (r *CampaignsResource) Leads(ctx context.Context, id string) (*managerapi.PaginatedLeadResponse, error) {
	resp, err := r.gc.ListCampaignLeadsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ProcessedLeads lists the already-processed leads of a campaign.
func (r *CampaignsResource) ProcessedLeads(ctx context.Context, id string) (*managerapi.PaginatedLeadResponse, error) {
	resp, err := r.gc.ListCampaignProcessedLeadsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetList returns the lead list assigned to a campaign.
func (r *CampaignsResource) GetList(ctx context.Context, id string) (*managerapi.LeadListItemResponse, error) {
	resp, err := r.gc.GetCampaignListWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// SetList assigns a lead list to a campaign.
func (r *CampaignsResource) SetList(ctx context.Context, id string, body managerapi.SetCampaignListRequest) (*managerapi.CampaignItemResponse, error) {
	resp, err := r.gc.SetCampaignListWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UnsetList removes the lead list assigned to a campaign.
func (r *CampaignsResource) UnsetList(ctx context.Context, id string) (*managerapi.CampaignItemResponse, error) {
	resp, err := r.gc.UnsetCampaignListWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// SetListById assigns a lead list to a campaign by list id.
func (r *CampaignsResource) SetListById(ctx context.Context, id string, listID string) (*managerapi.CampaignItemResponse, error) {
	resp, err := r.gc.SetCampaignListByIdWithResponse(ctx, id, listID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// LogoutAllAgents logs out all agents from a campaign.
func (r *CampaignsResource) LogoutAllAgents(ctx context.Context, id string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.LogoutAllCampaignAgentsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Statistics returns aggregated statistics for a campaign.
func (r *CampaignsResource) Statistics(ctx context.Context, id string, params *managerapi.GetCampaignStatisticsParams) (*managerapi.CampaignStatisticsResponse, error) {
	resp, err := r.gc.GetCampaignStatisticsWithResponse(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Status returns the realtime status of a campaign.
func (r *CampaignsResource) Status(ctx context.Context, id string) (*managerapi.CampaignRealtimeStatusResponse, error) {
	resp, err := r.gc.GetCampaignStatusWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
