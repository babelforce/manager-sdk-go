package manager

import (
	"context"
	"errors"
	"net/http"

	authapi "github.com/babelforce/manager-sdk-go/gen/auth"
	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
	taskautomationapi "github.com/babelforce/manager-sdk-go/gen/taskautomation"
	taskscheduleapi "github.com/babelforce/manager-sdk-go/gen/taskschedule"
	userapi "github.com/babelforce/manager-sdk-go/gen/user"
)

// DefaultBaseURL is the babelforce API host used when [Options.BaseURL] is empty.
const DefaultBaseURL = "https://services.babelforce.com"

// OrgIDHeader is the header that makes a partner's service token act in one of its direct
// sub-accounts.
const OrgIDHeader = "X-Babelforce-Org-Id"

// Options configures a [ManagerClient].
type Options struct {
	// BaseURL is the base URL of the babelforce API. Defaults to [DefaultBaseURL].
	BaseURL string
	// Auth is how the client authenticates. Required.
	Auth Auth
	// HTTPClient is the underlying HTTP client. Defaults to http.DefaultClient.
	HTTPClient *http.Client
	// Retry tunes automatic retries. Nil uses sensible defaults (see [RetryPolicy]); set
	// &RetryPolicy{MaxRetries: 0} to disable.
	Retry *RetryPolicy
	// OrgID makes the client act in this direct sub-account of your partner account: every API
	// call carries the X-Babelforce-Org-Id header with this value (the sub-account's orgKey).
	// Honoured for a service token holding account:manage only ([ClientCredentials]). See also
	// [ManagerClient.ForSubAccount].
	OrgID string
}

// ManagerClient is the babelforce manager SDK client. Create one with [Connect].
type ManagerClient struct {
	// Users is the user-management namespace (/api/v2/users).
	Users *UsersResource
	// Me is the authenticated-principal namespace (/api/v2/user): current user, accounts.
	Me *MeResource
	// Agents is the agent-management namespace (/api/v2/agents).
	Agents *AgentsResource
	// Calls is the call namespace (/api/v2/calls): reporting and call control.
	Calls *CallsResource
	// Sms is the SMS-records namespace (/api/v2/sms).
	Sms *SmsResource
	// Numbers is the service-numbers namespace (/api/v2/numbers).
	Numbers *NumbersResource
	// Conferences is the conferences namespace (/api/v2/conferences).
	Conferences *ConferencesResource
	// Queues is the queues namespace (/api/v2/queues), including selections.
	Queues *QueuesResource
	// Routing is the routing-rules namespace (/api/v2/routings).
	Routing *RoutingResource
	// Triggers is the workflow-triggers namespace (/api/v2/triggers).
	Triggers *TriggersResource
	// Automations is the global-automations namespace (/api/v2/events/triggers).
	Automations *AutomationsResource
	// Integrations is the integrations namespace (/api/v2/integrations).
	Integrations *IntegrationsResource
	// Outbound is the outbound dialer-lists namespace (/api/v2/outbound/lists).
	Outbound *OutboundResource
	// Dialer is the dialer namespace (/api/v2/dialer): runtime info, queue control,
	// simple reporting and behaviours.
	Dialer *DialerResource
	// Phonebook is the phonebook-entries namespace (/api/v2/phonebook).
	Phonebook *PhonebookResource
	// Campaigns is the outbound-campaigns namespace (/api/v2/outbound/campaigns).
	Campaigns *CampaignsResource
	// BusinessHours is the business-hours namespace (/api/v2/business-hours).
	BusinessHours *BusinessHoursResource
	// Calendars is the calendars namespace (/api/v2/calendars).
	Calendars *CalendarsResource
	// Conversations is the conversations namespace (/api/v2/conversations).
	Conversations *ConversationsResource
	// Sessions is the call/automation-sessions namespace (/api/v2/sessions).
	Sessions *SessionsResource
	// Prompts is the audio-prompts namespace (/api/v2/prompts).
	Prompts *PromptsResource
	// Babeldesk is the babeldesk namespace (/api/v2/babeldesk): dashboards and widgets.
	Babeldesk *BabeldeskResource
	// Dashboards is the reporting-dashboards namespace (/api/v2/dashboards).
	Dashboards *DashboardsResource
	// Files is the stored-files namespace (/api/v2/files).
	Files *FilesResource
	// Recordings is the call-recordings namespace (/api/v2/recordings).
	Recordings *RecordingsResource
	// Events is the events namespace (/api/v2/events): definitions and custom events.
	Events *EventsResource
	// Logs is the logs namespace: request audit logs (/api/v2/audit) and live logs (/api/v2/logs).
	Logs *LogsResource
	// Expressions is the expressions namespace (/api/v2/expressions): catalog and evaluation.
	Expressions *ExpressionsResource
	// Metrics is the metrics namespace (/api/v2/metrics).
	Metrics *MetricsResource
	// System is the system/reference namespace: health checks (/api/v2/echo, /api/v2/ping,
	// /api/v2/status), server time and timezones (/api/v2/data), push tokens (/api/v2/push-token),
	// tags (/api/v2/tags), and template exports (/api/v2/templates/export).
	System *SystemResource
	// Applications is the application (IVR) management namespace (/api/v2/applications).
	Applications *ApplicationsResource
	// Settings is the global-settings namespace (/api/v2/settings).
	Settings *SettingsResource
	// Tasks is the task-automation namespace (/api/v3/tasks).
	Tasks *TasksResource
	// Auth is the OAuth 2.0 namespace (/oauth): authorize, token, and revoke. These endpoints
	// authenticate via OAuth client credentials in the request, independent of [Options.Auth].
	Auth *AuthResource
	// Accounts is the sub-accounts namespace of a partner account (/api/v2/accounts).
	Accounts *AccountsResource
	// OAuthApplications is the OAuth-applications namespace of the account (/oauth/applications).
	OAuthApplications *OAuthApplicationsResource

	// what the client was built from, kept so ForSubAccount can share it
	baseURL string
	hc      *http.Client
	edits   []editorFn
}

// Connect creates and configures a client.
func Connect(_ context.Context, opts Options) (*ManagerClient, error) {
	if opts.Auth == nil {
		return nil, errors.New("manager: Auth is required")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	// Wrap the transport with retries (unless disabled). Copy the client so we never mutate the
	// caller's (or the shared http.DefaultClient) — the retry-aware client also backs auth token fetches.
	policy := defaultRetryPolicy
	if opts.Retry != nil {
		policy = opts.Retry.effective()
	}
	if policy.MaxRetries > 0 {
		base := hc.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		hc = &http.Client{
			Transport:     newRetryTransport(base, policy),
			CheckRedirect: hc.CheckRedirect,
			Jar:           hc.Jar,
			Timeout:       hc.Timeout,
		}
	}

	edits := []editorFn{opts.Auth.editor(baseURL, hc), normalizeUUIDPath}
	if opts.OrgID != "" {
		edits = append(edits, orgIDEditor(opts.OrgID))
	}
	return build(baseURL, hc, edits)
}

// ForSubAccount returns a client that acts in one direct sub-account of your partner account, named
// by its orgKey: every API call carries the X-Babelforce-Org-Id header. It shares this client's
// credentials, token and retry settings, so one service token serves the partner account and every
// sub-account. This client is unchanged.
func (c *ManagerClient) ForSubAccount(orgID string) (*ManagerClient, error) {
	if orgID == "" {
		return nil, errors.New("manager: ForSubAccount needs a non-empty orgID")
	}
	// The new editor runs last; its Header.Set replaces any org header the parent carries.
	edits := append(append([]editorFn{}, c.edits...), orgIDEditor(orgID))
	return build(c.baseURL, c.hc, edits)
}

// orgIDEditor sets the X-Babelforce-Org-Id header. It runs after any earlier one, so the last value
// wins.
func orgIDEditor(orgID string) editorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set(OrgIDHeader, orgID)
		return nil
	}
}

// build wires every generated client to the same HTTP client and request editors.
func build(baseURL string, hc *http.Client, edits []editorFn) (*ManagerClient, error) {
	managerEditors := make([]managerapi.ClientOption, 0, len(edits)+1)
	taEditors := make([]taskautomationapi.ClientOption, 0, len(edits)+1)
	tsEditors := make([]taskscheduleapi.ClientOption, 0, len(edits)+1)
	userEditors := make([]userapi.ClientOption, 0, len(edits)+1)
	authEditors := make([]authapi.ClientOption, 0, len(edits)+1)
	managerEditors = append(managerEditors, managerapi.WithHTTPClient(hc))
	taEditors = append(taEditors, taskautomationapi.WithHTTPClient(hc))
	tsEditors = append(tsEditors, taskscheduleapi.WithHTTPClient(hc))
	userEditors = append(userEditors, userapi.WithHTTPClient(hc))
	authEditors = append(authEditors, authapi.WithHTTPClient(hc))
	for _, e := range edits {
		managerEditors = append(managerEditors, managerapi.WithRequestEditorFn(e))
		taEditors = append(taEditors, taskautomationapi.WithRequestEditorFn(e))
		tsEditors = append(tsEditors, taskscheduleapi.WithRequestEditorFn(e))
		userEditors = append(userEditors, userapi.WithRequestEditorFn(e))
		authEditors = append(authEditors, authapi.WithRequestEditorFn(e))
	}

	gc, err := managerapi.NewClientWithResponses(baseURL, managerEditors...)
	if err != nil {
		return nil, err
	}
	tac, err := taskautomationapi.NewClientWithResponses(baseURL, taEditors...)
	if err != nil {
		return nil, err
	}
	tsc, err := taskscheduleapi.NewClientWithResponses(baseURL, tsEditors...)
	if err != nil {
		return nil, err
	}
	uc, err := userapi.NewClientWithResponses(baseURL, userEditors...)
	if err != nil {
		return nil, err
	}
	au, err := authapi.NewClientWithResponses(baseURL, authEditors...)
	if err != nil {
		return nil, err
	}

	return &ManagerClient{
		Users:         &UsersResource{gc: gc},
		Me:            &MeResource{uc: uc},
		Agents:        &AgentsResource{gc: gc, Groups: &AgentGroupsResource{gc: gc}},
		Calls:         &CallsResource{gc: gc, Reporting: &ReportingResource{gc: gc}},
		Sms:           &SmsResource{gc: gc},
		Numbers:       &NumbersResource{gc: gc},
		Conferences:   &ConferencesResource{gc: gc},
		Queues:        &QueuesResource{gc: gc, Selections: &QueueSelectionsResource{gc: gc}},
		Routing:       &RoutingResource{gc: gc},
		Triggers:      &TriggersResource{gc: gc},
		Automations:   &AutomationsResource{gc: gc},
		Integrations:  &IntegrationsResource{gc: gc},
		Outbound:      &OutboundResource{gc: gc},
		Dialer:        &DialerResource{gc: gc, Behaviours: &DialerBehavioursResource{gc: gc}},
		Phonebook:     &PhonebookResource{gc: gc},
		Campaigns:     &CampaignsResource{gc: gc},
		BusinessHours: &BusinessHoursResource{gc: gc},
		Calendars:     &CalendarsResource{gc: gc},
		Conversations: &ConversationsResource{gc: gc},
		Sessions:      &SessionsResource{gc: gc},
		Prompts:       &PromptsResource{gc: gc},
		Babeldesk:     &BabeldeskResource{gc: gc, Widgets: &BabeldeskWidgetsResource{gc: gc}},
		Dashboards:    &DashboardsResource{gc: gc},
		Files:         &FilesResource{gc: gc},
		Recordings:    &RecordingsResource{gc: gc},
		Events:        &EventsResource{gc: gc},
		Logs:          &LogsResource{gc: gc},
		Expressions:   &ExpressionsResource{gc: gc},
		Metrics:       &MetricsResource{gc: gc},
		System:        &SystemResource{gc: gc},
		Applications:  &ApplicationsResource{gc: gc, Actions: &AppActionsResource{gc: gc}},
		Settings:      newSettingsResource(gc),
		Tasks: &TasksResource{
			ta:              tac,
			Schedules:       &TaskSchedulesResource{ts: tsc},
			Scripts:         &TaskScriptsResource{ta: tac},
			Secrets:         &TaskSecretsResource{ta: tac},
			SelectionConfig: &TaskSelectionConfigResource{ta: tac},
			Metrics:         &TaskMetricsResource{ta: tac},
		},
		Auth:              &AuthResource{au: au},
		Accounts:          &AccountsResource{gc: gc},
		OAuthApplications: &OAuthApplicationsResource{au: au},
		baseURL:           baseURL,
		hc:                hc,
		edits:             edits,
	}, nil
}
