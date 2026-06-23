package manager

import (
	"context"

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
	AgentOutbound  Setting[managerapi.SettingsTelephonyAgentOutbound, managerapi.SettingsTelephonyAgentOutboundRequestData]
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
type SettingsResource struct {
	App       AppSettings
	Telephony TelephonySettings
	Audit     AuditSettings
	Ui        UiSettings
	Retention RetentionSettings
}

func newSettingsResource(gc *managerapi.ClientWithResponses) *SettingsResource {
	return &SettingsResource{
		App: AppSettings{
			CustomerLogging: Setting[managerapi.SettingsAppCustomerLogging, managerapi.SettingsAppCustomerLoggingRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppCustomerLogging, error) {
					resp, err := gc.GetSettingsForAppCustomerLoggingWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsAppCustomerLoggingRequestData) (*managerapi.SettingsAppCustomerLogging, error) {
					body := managerapi.SettingsAppCustomerLoggingRequest{
						Scope: managerapi.SettingsAppCustomerLoggingRequestScope("app"),
						Key:   managerapi.SettingsAppCustomerLoggingRequestKey("customer.logging"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForAppCustomerLoggingWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			Conversations: Setting[managerapi.SettingsAppConversations, managerapi.SettingsAppConversationsRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppConversations, error) {
					resp, err := gc.GetSettingsForAppConversationsWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsAppConversationsRequestData) (*managerapi.SettingsAppConversations, error) {
					body := managerapi.SettingsAppConversationsRequest{
						Scope: managerapi.SettingsAppConversationsRequestScope("app"),
						Key:   managerapi.SettingsAppConversationsRequestKey("conversations"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForAppConversationsWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			Integrations: Setting[managerapi.SettingsAppIntegrations, managerapi.SettingsAppIntegrationsRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppIntegrations, error) {
					resp, err := gc.GetSettingsForAppIntegrationsWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsAppIntegrationsRequestData) (*managerapi.SettingsAppIntegrations, error) {
					body := managerapi.SettingsAppIntegrationsRequest{
						Scope: managerapi.SettingsAppIntegrationsRequestScope("app"),
						Key:   managerapi.SettingsAppIntegrationsRequestKey("integrations"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForAppIntegrationsWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			AgentStatus: Setting[managerapi.SettingsAppAgentStatus, managerapi.SettingsAppAgentStatusRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsAppAgentStatus, error) {
					resp, err := gc.GetSettingsForAppAgentStatusWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsAppAgentStatusRequestData) (*managerapi.SettingsAppAgentStatus, error) {
					body := managerapi.SettingsAppAgentStatusRequest{
						Scope: managerapi.SettingsAppAgentStatusRequestScope("app"),
						Key:   managerapi.SettingsAppAgentStatusRequestKey("agent.status"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForAppAgentStatusWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
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
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentInboundRequestData) (*managerapi.SettingsTelephonyAgentInbound, error) {
					body := managerapi.SettingsTelephonyAgentInboundRequest{
						Scope: managerapi.SettingsTelephonyAgentInboundRequestScope("telephony"),
						Key:   managerapi.SettingsTelephonyAgentInboundRequestKey("agent.inbound"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForTelephonyAgentInboundWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			AgentOutbound: Setting[managerapi.SettingsTelephonyAgentOutbound, managerapi.SettingsTelephonyAgentOutboundRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentOutbound, error) {
					resp, err := gc.GetSettingsForTelephonyAgentOutboundWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentOutboundRequestData) (*managerapi.SettingsTelephonyAgentOutbound, error) {
					body := managerapi.SettingsTelephonyAgentOutboundRequest{
						Scope: managerapi.SettingsTelephonyAgentOutboundRequestScope("telephony"),
						Key:   managerapi.SettingsTelephonyAgentOutboundRequestKey("agent.outbound"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForTelephonyAgentOutboundWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			AgentRecording: Setting[managerapi.SettingsTelephonyAgentRecording, managerapi.SettingsTelephonyAgentRecordingRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentRecording, error) {
					resp, err := gc.GetSettingsForTelephonyAgentRecordingWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentRecordingRequestData) (*managerapi.SettingsTelephonyAgentRecording, error) {
					body := managerapi.SettingsTelephonyAgentRecordingRequest{
						Scope: managerapi.SettingsTelephonyAgentRecordingRequestScope("telephony"),
						Key:   managerapi.SettingsTelephonyAgentRecordingRequestKey("agent.recording"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForTelephonyAgentRecordingWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			AgentWrapup: Setting[managerapi.SettingsTelephonyAgentWrapup, managerapi.SettingsTelephonyAgentWrapupRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyAgentWrapup, error) {
					resp, err := gc.GetSettingsForTelephonyAgentWrapupWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyAgentWrapupRequestData) (*managerapi.SettingsTelephonyAgentWrapup, error) {
					body := managerapi.SettingsTelephonyAgentWrapupRequest{
						Scope: managerapi.SettingsTelephonyAgentWrapupRequestScope("telephony"),
						Key:   managerapi.SettingsTelephonyAgentWrapupRequestKey("agent.wrapup"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForTelephonyAgentWrapupWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
			PostCall: Setting[managerapi.SettingsTelephonyPostCall, managerapi.SettingsTelephonyPostCallRequestData]{
				get: func(ctx context.Context) (*managerapi.SettingsTelephonyPostCall, error) {
					resp, err := gc.GetSettingsForTelephonyPostCallWithResponse(ctx)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsTelephonyPostCallRequestData) (*managerapi.SettingsTelephonyPostCall, error) {
					body := managerapi.SettingsTelephonyPostCallRequest{
						Scope: managerapi.SettingsTelephonyPostCallRequestScope("telephony"),
						Key:   managerapi.SettingsTelephonyPostCallRequestKey("post-call"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForTelephonyPostCallWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
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
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsAuditDefaultRequestData) (*managerapi.SettingsAuditDefault, error) {
					body := managerapi.SettingsAuditDefaultRequest{
						Scope: managerapi.SettingsAuditDefaultRequestScope("audit"),
						Key:   managerapi.SettingsAuditDefaultRequestKey("default"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForAuditDefaultWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
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
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsUiI18nRequestData) (*managerapi.SettingsUiI18n, error) {
					body := managerapi.SettingsUiI18nRequest{
						Scope: managerapi.SettingsUiI18nRequestScope("ui"),
						Key:   managerapi.SettingsUiI18nRequestKey("i18n"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForUiI18nWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
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
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
				update: func(ctx context.Context, data managerapi.SettingsRetentionPeriodsRequestData) (*managerapi.SettingsRetentionPeriods, error) {
					body := managerapi.SettingsRetentionPeriodsRequest{
						Scope: managerapi.SettingsRetentionPeriodsRequestScope("retention"),
						Key:   managerapi.SettingsRetentionPeriodsRequestKey("periods"),
						Data:  data,
					}
					resp, err := gc.UpdateSettingsForRetentionPeriodsWithResponse(ctx, body)
					if err != nil {
						return nil, err
					}
					r, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
					if err != nil {
						return nil, err
					}
					return &r.Item.Data, nil
				},
			},
		},
	}
}
