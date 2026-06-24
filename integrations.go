package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// IntegrationsResource is the integrations namespace (/api/v2/integrations).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type IntegrationsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over configured integrations, auto-paginating across pages.
func (r *IntegrationsResource) List(ctx context.Context, params managerapi.ListIntegrationsParams) iter.Seq2[managerapi.Integration, error] {
	return func(yield func(managerapi.Integration, error) bool) {
		var zero managerapi.Integration
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListIntegrationsWithResponse(ctx, &p)
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

// ListAll collects every integration into a slice (convenience over List).
func (r *IntegrationsResource) ListAll(ctx context.Context, params managerapi.ListIntegrationsParams) ([]managerapi.Integration, error) {
	var out []managerapi.Integration
	for x, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

// Create creates an integration.
func (r *IntegrationsResource) Create(ctx context.Context, body managerapi.IntegrationCreateRequest) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.CreateIntegrationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns an integration by id.
func (r *IntegrationsResource) Get(ctx context.Context, id string) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.GetIntegrationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates an integration.
func (r *IntegrationsResource) Update(ctx context.Context, id string, body managerapi.IntegrationUpdateRequest) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.UpdateIntegrationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes an integration.
func (r *IntegrationsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteIntegrationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Available lists the integration providers available to this account.
func (r *IntegrationsResource) Available(ctx context.Context) (*managerapi.IntegrationListAvailableIntegrationsResponse, error) {
	resp, err := r.gc.ListAvailableIntegrationsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddAssociation associates an integration action with an object.
func (r *IntegrationsResource) AddAssociation(ctx context.Context, integrationID, associationID, actionName string) (*managerapi.IntegrationAddAssociationResponse, error) {
	resp, err := r.gc.AddIntegrationAssociationWithResponse(ctx, integrationID, associationID, actionName)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// RemoveAssociation removes an integration action association.
func (r *IntegrationsResource) RemoveAssociation(ctx context.Context, integrationID, associationID, actionName string) (*managerapi.IntegrationRemoveAssociationResponse, error) {
	resp, err := r.gc.DeleteIntegrationAssociationWithResponse(ctx, integrationID, associationID, actionName)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ProviderLogo returns a provider's logo at a given size.
func (r *IntegrationsResource) ProviderLogo(ctx context.Context, providerName managerapi.IntegrationProvider, size string) (*managerapi.GetIntegrationProviderLogoResponse, error) {
	resp, err := r.gc.GetIntegrationProviderLogoWithResponse(ctx, providerName, size)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ProviderSessionVariables lists the session variables a provider's actions expose.
func (r *IntegrationsResource) ProviderSessionVariables(ctx context.Context, provider string) (*managerapi.IntegrationSessionVariableItemsResponse, error) {
	resp, err := r.gc.ListProviderSessionVariablesWithResponse(ctx, provider)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DispatchAction dispatches an integration action.
func (r *IntegrationsResource) DispatchAction(ctx context.Context, integrationID, action string, body managerapi.IntegrationDispatchActionRequest) (*managerapi.IntegrationDispatchActionResponse, error) {
	resp, err := r.gc.DispatchActionWithResponse(ctx, integrationID, action, nil, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ActionVariables lists the variables a single provider action exposes.
func (r *IntegrationsResource) ActionVariables(ctx context.Context, provider managerapi.IntegrationProvider, actionName string) (*managerapi.VariableDefinitionItemsResponse, error) {
	resp, err := r.gc.ListSingleActionSessionVariablesWithResponse(ctx, provider, actionName)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
