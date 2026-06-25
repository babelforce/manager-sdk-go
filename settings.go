package manager

import (
	"context"
	"net/http"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// Setting is one global-settings group: read its full value with Get, replace it with Update.
// TGet is the returned value type; TUpd is the (partial, all-optional) update payload type.
type Setting[TGet any, TUpd any] struct {
	get    func(context.Context) (*TGet, error)
	update func(context.Context, TUpd) (*TGet, error)
}

// Get reads the current value of this settings group.
func (s Setting[TGet, TUpd]) Get(ctx context.Context) (*TGet, error) { return s.get(ctx) }

// Update replaces this settings group and returns the new value.
func (s Setting[TGet, TUpd]) Update(ctx context.Context, data TUpd) (*TGet, error) {
	return s.update(ctx, data)
}

// unwrapItem applies result() to a generated settings response and returns a pointer to the item's
// data payload. It centralizes the {item:{data}} envelope handling shared by every settings group;
// the per-group data closure only names the response type (generics cannot reach r.Item.Data).
func unwrapItem[W any, T any](j200 *W, httpResp *http.Response, body []byte, data func(*W) *T) (*T, error) {
	r, err := result(j200, httpResp, body)
	if err != nil {
		return nil, err
	}
	return data(r), nil
}

// AppSettings groups the `app` settings.
type AppSettings struct {
	CustomerLogging Setting[managerapi.SettingsAppCustomerLogging, managerapi.SettingsAppCustomerLoggingRequestData]
	Conversations   Setting[managerapi.SettingsAppConversations, managerapi.SettingsAppConversationsRequestData]
	Integrations    Setting[managerapi.SettingsAppIntegrations, managerapi.SettingsAppIntegrationsRequestData]
	AgentStatus     Setting[managerapi.SettingsAppAgentStatus, managerapi.SettingsAppAgentStatusRequestData]
}

// TelephonySettings groups the `telephony` settings.
type TelephonySettings struct {
	AgentInbound   Setting[managerapi.SettingsTelephonyAgentInbound, managerapi.SettingsTelephonyAgentInboundRequestData]
	AgentOutbound  Setting[managerapi.SettingsTelephonyAgentOutbound, managerapi.SettingsTelephonyAgentInboundRequestData]
	AgentRecording Setting[managerapi.SettingsTelephonyAgentRecording, managerapi.SettingsTelephonyAgentRecordingRequestData]
	AgentWrapup    Setting[managerapi.SettingsTelephonyAgentWrapup, managerapi.SettingsTelephonyAgentWrapupRequestData]
	PostCall       Setting[managerapi.SettingsTelephonyPostCall, managerapi.SettingsTelephonyPostCallRequestData]
}

// AuditSettings groups the `audit` settings.
type AuditSettings struct {
	Default Setting[managerapi.SettingsAuditDefault, managerapi.SettingsAuditDefaultRequestData]
}

// UiSettings groups the `ui` settings.
type UiSettings struct {
	I18n Setting[managerapi.SettingsUiI18n, managerapi.SettingsUiI18nRequestData]
}

// RetentionSettings groups the `retention` settings.
type RetentionSettings struct {
	Periods Setting[managerapi.SettingsRetentionPeriods, managerapi.SettingsRetentionPeriodsRequestData]
}

// SettingsResource is the global-settings namespace (/api/v2/settings), grouped by scope.
//
// The typed section accessors (App, Telephony, …) read and replace individual scope/key groups.
// The generic methods (ListAll, ListInScope, Clear, ClearInScope, ClearAll) operate across the
// whole settings collection or an entire scope at once.
type SettingsResource struct {
	gc        *managerapi.ClientWithResponses
	App       AppSettings
	Telephony TelephonySettings
	Audit     AuditSettings
	Ui        UiSettings
	Retention RetentionSettings
}

// ListAll lists every customer setting across all scopes (GET /api/v2/settings).
func (r *SettingsResource) ListAll(ctx context.Context) (*managerapi.SettingsListResponse, error) {
	resp, err := r.gc.ListAllSettingsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListInScope lists every customer setting in a scope (GET /api/v2/settings/{scope}).
func (r *SettingsResource) ListInScope(ctx context.Context, scope string) (*managerapi.SettingsListResponse, error) {
	resp, err := r.gc.ListSettingsInScopeWithResponse(ctx, scope)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Clear resets a single setting to its default and returns the cleared item
// (DELETE /api/v2/settings/{scope}/{key}).
func (r *SettingsResource) Clear(ctx context.Context, scope, key string) (*managerapi.SettingItemResponse, error) {
	resp, err := r.gc.ClearSettingWithResponse(ctx, scope, key)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ClearInScope resets every setting in a scope to its default and returns the remaining list
// (DELETE /api/v2/settings/{scope}).
func (r *SettingsResource) ClearInScope(ctx context.Context, scope string) (*managerapi.SettingsListResponse, error) {
	resp, err := r.gc.ClearSettingsInScopeWithResponse(ctx, scope)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ClearAll resets every customer setting across all scopes to its default and returns the remaining
// list (DELETE /api/v2/settings).
func (r *SettingsResource) ClearAll(ctx context.Context) (*managerapi.SettingsListResponse, error) {
	resp, err := r.gc.ClearAllSettingsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

func newSettingsResource(gc *managerapi.ClientWithResponses) *SettingsResource {
	return &SettingsResource{
		gc: gc,
		App: AppSettings{
			CustomerLogging: Setting[managerapi.SettingsAppCustomerLogging, managerapi.SettingsAppCustomerLoggingRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppCustomerLogging, error) {
					resp, err := gc.GetSettingsForAppCustomerLoggingWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppCustomerLoggingResponse) *managerapi.SettingsAppCustomerLogging {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsAppCustomerLoggingRequestData) (*managerapi.SettingsAppCustomerLogging, error) {
					resp, err := gc.UpdateSettingsForAppCustomerLoggingWithResponse(ctx, managerapi.SettingsAppCustomerLoggingRequest{
						Scope: "app",
						Key:   "customer.logging",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppCustomerLoggingResponse) *managerapi.SettingsAppCustomerLogging {
						return &r.Item.Data
					})
				},
			},
			Conversations: Setting[managerapi.SettingsAppConversations, managerapi.SettingsAppConversationsRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppConversations, error) {
					resp, err := gc.GetSettingsForAppConversationsWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppConversationsResponse) *managerapi.SettingsAppConversations {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsAppConversationsRequestData) (*managerapi.SettingsAppConversations, error) {
					resp, err := gc.UpdateSettingsForAppConversationsWithResponse(ctx, managerapi.SettingsAppConversationsRequest{
						Scope: "app",
						Key:   "conversations",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppConversationsResponse) *managerapi.SettingsAppConversations {
						return &r.Item.Data
					})
				},
			},
			Integrations: Setting[managerapi.SettingsAppIntegrations, managerapi.SettingsAppIntegrationsRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppIntegrations, error) {
					resp, err := gc.GetSettingsForAppIntegrationsWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppIntegrationsResponse) *managerapi.SettingsAppIntegrations {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsAppIntegrationsRequestData) (*managerapi.SettingsAppIntegrations, error) {
					resp, err := gc.UpdateSettingsForAppIntegrationsWithResponse(ctx, managerapi.SettingsAppIntegrationsRequest{
						Scope: "app",
						Key:   "integrations",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppIntegrationsResponse) *managerapi.SettingsAppIntegrations {
						return &r.Item.Data
					})
				},
			},
			AgentStatus: Setting[managerapi.SettingsAppAgentStatus, managerapi.SettingsAppAgentStatusRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppAgentStatus, error) {
					resp, err := gc.GetSettingsForAppAgentStatusWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppAgentStatusResponse) *managerapi.SettingsAppAgentStatus {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsAppAgentStatusRequestData) (*managerapi.SettingsAppAgentStatus, error) {
					resp, err := gc.UpdateSettingsForAppAgentStatusWithResponse(ctx, managerapi.SettingsAppAgentStatusRequest{
						Scope: "app",
						Key:   "agent.status",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAppAgentStatusResponse) *managerapi.SettingsAppAgentStatus {
						return &r.Item.Data
					})
				},
			},
		},
		Telephony: TelephonySettings{
			AgentInbound: Setting[managerapi.SettingsTelephonyAgentInbound, managerapi.SettingsTelephonyAgentInboundRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentInbound, error) {
					resp, err := gc.GetSettingsForTelephonyAgentInboundWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentInboundResponse) *managerapi.SettingsTelephonyAgentInbound {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentInboundRequestData) (*managerapi.SettingsTelephonyAgentInbound, error) {
					resp, err := gc.UpdateSettingsForTelephonyAgentInboundWithResponse(ctx, managerapi.SettingsTelephonyAgentInboundRequest{
						Scope: "telephony",
						Key:   "agent.inbound",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentInboundResponse) *managerapi.SettingsTelephonyAgentInbound {
						return &r.Item.Data
					})
				},
			},
			AgentOutbound: Setting[managerapi.SettingsTelephonyAgentOutbound, managerapi.SettingsTelephonyAgentInboundRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentOutbound, error) {
					resp, err := gc.GetSettingsForTelephonyAgentOutboundWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentOutboundResponse) *managerapi.SettingsTelephonyAgentOutbound {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentInboundRequestData) (*managerapi.SettingsTelephonyAgentOutbound, error) {
					resp, err := gc.UpdateSettingsForTelephonyAgentOutboundWithResponse(ctx, managerapi.SettingsTelephonyAgentOutboundRequest{
						Scope: "telephony",
						Key:   "agent.outbound",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentOutboundResponse) *managerapi.SettingsTelephonyAgentOutbound {
						return &r.Item.Data
					})
				},
			},
			AgentRecording: Setting[managerapi.SettingsTelephonyAgentRecording, managerapi.SettingsTelephonyAgentRecordingRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentRecording, error) {
					resp, err := gc.GetSettingsForTelephonyAgentRecordingWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentRecordingResponse) *managerapi.SettingsTelephonyAgentRecording {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentRecordingRequestData) (*managerapi.SettingsTelephonyAgentRecording, error) {
					resp, err := gc.UpdateSettingsForTelephonyAgentRecordingWithResponse(ctx, managerapi.SettingsTelephonyAgentRecordingRequest{
						Scope: "telephony",
						Key:   "agent.recording",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentRecordingResponse) *managerapi.SettingsTelephonyAgentRecording {
						return &r.Item.Data
					})
				},
			},
			AgentWrapup: Setting[managerapi.SettingsTelephonyAgentWrapup, managerapi.SettingsTelephonyAgentWrapupRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentWrapup, error) {
					resp, err := gc.GetSettingsForTelephonyAgentWrapupWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentWrapupResponse) *managerapi.SettingsTelephonyAgentWrapup {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentWrapupRequestData) (*managerapi.SettingsTelephonyAgentWrapup, error) {
					resp, err := gc.UpdateSettingsForTelephonyAgentWrapupWithResponse(ctx, managerapi.SettingsTelephonyAgentWrapupRequest{
						Scope: "telephony",
						Key:   "agent.wrapup",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyAgentWrapupResponse) *managerapi.SettingsTelephonyAgentWrapup {
						return &r.Item.Data
					})
				},
			},
			PostCall: Setting[managerapi.SettingsTelephonyPostCall, managerapi.SettingsTelephonyPostCallRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyPostCall, error) {
					resp, err := gc.GetSettingsForTelephonyPostCallWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyPostCallResponse) *managerapi.SettingsTelephonyPostCall {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyPostCallRequestData) (*managerapi.SettingsTelephonyPostCall, error) {
					resp, err := gc.UpdateSettingsForTelephonyPostCallWithResponse(ctx, managerapi.SettingsTelephonyPostCallRequest{
						Scope: "telephony",
						Key:   "post-call",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsTelephonyPostCallResponse) *managerapi.SettingsTelephonyPostCall {
						return &r.Item.Data
					})
				},
			},
		},
		Audit: AuditSettings{
			Default: Setting[managerapi.SettingsAuditDefault, managerapi.SettingsAuditDefaultRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAuditDefault, error) {
					resp, err := gc.GetSettingsForAuditDefaultWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAuditDefaultResponse) *managerapi.SettingsAuditDefault { return &r.Item.Data })
				},
				update: func(ctx context.Context, data managerapi.SettingsAuditDefaultRequestData) (*managerapi.SettingsAuditDefault, error) {
					resp, err := gc.UpdateSettingsForAuditDefaultWithResponse(ctx, managerapi.SettingsAuditDefaultRequest{
						Scope: "audit",
						Key:   "default",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsAuditDefaultResponse) *managerapi.SettingsAuditDefault { return &r.Item.Data })
				},
			},
		},
		Ui: UiSettings{
			I18n: Setting[managerapi.SettingsUiI18n, managerapi.SettingsUiI18nRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsUiI18n, error) {
					resp, err := gc.GetSettingsForUiI18nWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsUiI18nResponse) *managerapi.SettingsUiI18n { return &r.Item.Data })
				},
				update: func(ctx context.Context, data managerapi.SettingsUiI18nRequestData) (*managerapi.SettingsUiI18n, error) {
					resp, err := gc.UpdateSettingsForUiI18nWithResponse(ctx, managerapi.SettingsUiI18nRequest{
						Scope: "ui",
						Key:   "i18n",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsUiI18nResponse) *managerapi.SettingsUiI18n { return &r.Item.Data })
				},
			},
		},
		Retention: RetentionSettings{
			Periods: Setting[managerapi.SettingsRetentionPeriods, managerapi.SettingsRetentionPeriodsRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsRetentionPeriods, error) {
					resp, err := gc.GetSettingsForRetentionPeriodsWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsRetentionPeriodsResponse) *managerapi.SettingsRetentionPeriods {
						return &r.Item.Data
					})
				},
				update: func(ctx context.Context, data managerapi.SettingsRetentionPeriodsRequestData) (*managerapi.SettingsRetentionPeriods, error) {
					resp, err := gc.UpdateSettingsForRetentionPeriodsWithResponse(ctx, managerapi.SettingsRetentionPeriodsRequest{
						Scope: "retention",
						Key:   "periods",
						Data:  data,
					})
					if err != nil {
						return nil, err
					}
					return unwrapItem(resp.JSON200, resp.HTTPResponse, resp.Body, func(r *managerapi.SettingsRetentionPeriodsResponse) *managerapi.SettingsRetentionPeriods {
						return &r.Item.Data
					})
				},
			},
		},
	}
}
