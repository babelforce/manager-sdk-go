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

// Providers lists the integration providers known to the manager.
func (r *IntegrationsResource) Providers(ctx context.Context) (*managerapi.IntegrationObjectListResponse, error) {
	resp, err := r.gc.ListIntegrationProvidersWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ProviderTemplate returns the configuration template for a provider.
func (r *IntegrationsResource) ProviderTemplate(ctx context.Context, provider string) (*map[string]any, error) {
	resp, err := r.gc.GetIntegrationProviderTemplateWithResponse(ctx, provider)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Template returns the configuration template for a given type and provider.
func (r *IntegrationsResource) Template(ctx context.Context, typ, provider string) (*map[string]any, error) {
	resp, err := r.gc.GetIntegrationTemplateWithResponse(ctx, typ, provider)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Authorize completes the authorization step for an integration.
func (r *IntegrationsResource) Authorize(ctx context.Context, id string, body map[string]any) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.AuthorizeIntegrationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Clone clones an existing integration.
func (r *IntegrationsResource) Clone(ctx context.Context, id string) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.CloneIntegrationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Integrate runs the integrate step for an integration.
func (r *IntegrationsResource) Integrate(ctx context.Context, id string, body map[string]any) (*managerapi.IntegrationItemResponse, error) {
	resp, err := r.gc.IntegrateIntegrationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkUpdate updates several integrations in one request.
func (r *IntegrationsResource) BulkUpdate(ctx context.Context, body managerapi.BulkUpdateIntegrationsRequest) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.BulkUpdateIntegrationsWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkDelete deletes several integrations by id.
func (r *IntegrationsResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.BulkDeleteIntegrationsWithResponse(ctx, managerapi.IntegrationBulkIdsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TypeActions lists the actions available for an integration type.
func (r *IntegrationsResource) TypeActions(ctx context.Context, typ string) (*managerapi.IntegrationObjectListResponse, error) {
	resp, err := r.gc.ListIntegrationTypeActionsWithResponse(ctx, typ)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DispatchTypeAction dispatches an action on an integration by type.
func (r *IntegrationsResource) DispatchTypeAction(ctx context.Context, typ, id, action string, body map[string]any) (*map[string]any, error) {
	resp, err := r.gc.DispatchIntegrationTypeActionWithResponse(ctx, typ, id, action, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// APIProxyGet proxies a GET request to an integration's upstream API and returns the raw response body.
func (r *IntegrationsResource) APIProxyGet(ctx context.Context, integrationID, uri string) ([]byte, error) {
	resp, err := r.gc.IntegrationApiProxyGetWithResponse(ctx, integrationID, uri)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// APIProxyPost proxies a POST request to an integration's upstream API and returns the raw response body.
func (r *IntegrationsResource) APIProxyPost(ctx context.Context, integrationID, uri string, body map[string]any) ([]byte, error) {
	resp, err := r.gc.IntegrationApiProxyPostWithResponse(ctx, integrationID, uri, body)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ListTokens lists the OAuth tokens stored for an integration.
func (r *IntegrationsResource) ListTokens(ctx context.Context, id string) (*managerapi.IntegrationTokenListResponse, error) {
	resp, err := r.gc.ListIntegrationTokensWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetToken returns a single integration token by id.
func (r *IntegrationsResource) GetToken(ctx context.Context, id, tokenID string) (*managerapi.IntegrationTokenItemResponse, error) {
	resp, err := r.gc.GetIntegrationTokenWithResponse(ctx, id, tokenID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeleteToken deletes an integration token by id.
func (r *IntegrationsResource) DeleteToken(ctx context.Context, id, tokenID string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DeleteIntegrationTokenWithResponse(ctx, id, tokenID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// RefreshToken refreshes an integration token by id.
func (r *IntegrationsResource) RefreshToken(ctx context.Context, id, tokenID string) (*managerapi.IntegrationTokenItemResponse, error) {
	resp, err := r.gc.RefreshIntegrationTokenWithResponse(ctx, id, tokenID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListActions lists the actions available across providers, optionally filtered by type.
func (r *IntegrationsResource) ListActions(ctx context.Context, params managerapi.ListActionsParams) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.ListActionsWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListActionParams lists the parameters a single provider action accepts.
func (r *IntegrationsResource) ListActionParams(ctx context.Context, providerName, providerActionName string) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.ListActionParamsWithResponse(ctx, providerName, providerActionName)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ExecuteAction executes an action by type and name with a free-form request body.
func (r *IntegrationsResource) ExecuteAction(ctx context.Context, actionType, actionName string, body map[string]any) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.ExecuteActionWithResponse(ctx, actionType, actionName, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DispatchActionGet dispatches an integration action via GET, optionally scoped to a call or session.
func (r *IntegrationsResource) DispatchActionGet(ctx context.Context, integrationID, action string, params managerapi.DispatchActionGetParams) (*managerapi.IntegrationDispatchActionResponse, error) {
	resp, err := r.gc.DispatchActionGetWithResponse(ctx, integrationID, action, &params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
