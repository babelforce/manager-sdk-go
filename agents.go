package manager

import (
	"context"
	"io"
	"iter"

	"github.com/google/uuid"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// AgentsResource is the agent-management namespace (/api/v2/agents).
type AgentsResource struct {
	gc *managerapi.ClientWithResponses
	// Groups is the agent-groups sub-namespace (/api/v2/agents/groups).
	Groups *AgentGroupsResource
}

// ListAgentsQuery filters an agent listing.
type ListAgentsQuery struct {
	// Q searches name, group name, number, email, sourceId and integration label at once.
	Q *string
	// Enabled restricts to enabled (true) or disabled (false) agents.
	Enabled *bool
	// Name filters by agent name.
	Name *string
	// Number filters by the agent's number.
	Number *string
	// SourceId filters by integration source id.
	SourceId *string
	// State filters by line status.
	State *managerapi.AgentLineStatus
	// Source filters by source integration.
	Source *string
	// GroupIds restricts to agents in these group id(s).
	GroupIds []string
	// PageSize is the page size (the API's max). Zero uses the server default.
	PageSize int
}

// List returns an iterator over agents, auto-paginating across pages.
//
//	for agent, err := range mgr.Agents.List(ctx, manager.ListAgentsQuery{}) {
//	    if err != nil { return err }
//	    fmt.Println(agent.Name)
//	}
func (r *AgentsResource) List(ctx context.Context, q ListAgentsQuery) iter.Seq2[managerapi.Agent, error] {
	return func(yield func(managerapi.Agent, error) bool) {
		var zero managerapi.Agent

		params := &managerapi.ListAgentsParams{
			Q:        q.Q,
			Enabled:  q.Enabled,
			Name:     q.Name,
			Number:   q.Number,
			SourceId: q.SourceId,
			State:    q.State,
			Source:   q.Source,
		}
		if len(q.GroupIds) > 0 {
			var groupIds managerapi.ListAgentsGroupIdsParameter
			if err := groupIds.FromListAgentsGroupIdsParameter1(q.GroupIds); err != nil {
				yield(zero, err)
				return
			}
			params.GroupIds = &groupIds
		}
		if q.PageSize > 0 {
			params.Max = &q.PageSize
		}

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListAgentsWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, a := range data.Items {
				if !yield(a, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// ListAll collects every agent into a slice (convenience over List).
func (r *AgentsResource) ListAll(ctx context.Context, q ListAgentsQuery) ([]managerapi.Agent, error) {
	var agents []managerapi.Agent
	for a, err := range r.List(ctx, q) {
		if err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// Create creates an agent.
func (r *AgentsResource) Create(ctx context.Context, body managerapi.RestCreateAgent) (*managerapi.AgentItemResponse, error) {
	resp, err := r.gc.CreateAgentWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns an agent by id.
func (r *AgentsResource) Get(ctx context.Context, id string) (*managerapi.AgentItemResponse, error) {
	resp, err := r.gc.GetAgentWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates an agent.
func (r *AgentsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateAgent) (*managerapi.AgentItemResponse, error) {
	resp, err := r.gc.UpdateAgentWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes an agent by id.
func (r *AgentsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteAgentWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// UpdateStatus updates an agent's status (enabled flag and/or presence).
func (r *AgentsResource) UpdateStatus(ctx context.Context, id string, status managerapi.UpdateAgentStatusRequest) (*managerapi.AgentStatus, error) {
	resp, err := r.gc.UpdateAgentStatusWithResponse(ctx, id, status)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Presences lists the available agent presences.
func (r *AgentsResource) Presences(ctx context.Context) (*managerapi.AgentPresenceListResponse, error) {
	resp, err := r.gc.ListAgentPresencesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetPresence returns an available agent presence by name.
func (r *AgentsResource) GetPresence(ctx context.Context, name string) (*managerapi.AgentPresenceItemResponse, error) {
	resp, err := r.gc.GetAgentPresenceWithResponse(ctx, name)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// CreatePresence creates an available agent presence.
func (r *AgentsResource) CreatePresence(ctx context.Context, body managerapi.AgentPresenceWriteBody) (*managerapi.AgentPresenceItemResponse, error) {
	resp, err := r.gc.CreateAgentPresenceWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdatePresence updates an available agent presence by name.
func (r *AgentsResource) UpdatePresence(ctx context.Context, name string, body managerapi.AgentPresenceWriteBody) (*managerapi.AgentPresenceItemResponse, error) {
	resp, err := r.gc.UpdateAgentPresenceWithResponse(ctx, name, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeletePresence deletes an available agent presence by name.
func (r *AgentsResource) DeletePresence(ctx context.Context, name string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DeleteAgentPresenceWithResponse(ctx, name)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetStatus returns the total status of an agent by id.
func (r *AgentsResource) GetStatus(ctx context.Context, id string) (*managerapi.AgentTotalStatusResponse, error) {
	resp, err := r.gc.GetAgentStatusWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AvailableStatuses lists the agent availability statuses.
func (r *AgentsResource) AvailableStatuses(ctx context.Context) (*managerapi.AgentAvailabilityListResponse, error) {
	resp, err := r.gc.ListAvailableAgentStatusesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Enable enables an agent by id.
func (r *AgentsResource) Enable(ctx context.Context, id string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.EnableAgentWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Disable disables an agent by id.
func (r *AgentsResource) Disable(ctx context.Context, id string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DisableAgentWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// HangupCall hangs up the active call of an agent by id.
func (r *AgentsResource) HangupCall(ctx context.Context, id string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.HangupAgentCallWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkAction applies a bulk action (e.g. "enable", "disable", "delete") to the given agent ids.
func (r *AgentsResource) BulkAction(ctx context.Context, action string, body managerapi.AgentBulkRequest) (*managerapi.AgentBulkResponse, error) {
	resp, err := r.gc.BulkAgentActionWithResponse(ctx, managerapi.BulkAgentActionParamsBulkAction(action), body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Export returns all agents as a raw export in the given format (e.g. "csv").
func (r *AgentsResource) Export(ctx context.Context, format string) ([]byte, error) {
	params := &managerapi.ExportAgentsParams{Format: managerapi.ExportAgentsParamsFormat(format)}
	resp, err := r.gc.ExportAgentsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Import provisions agents from an upload stream (e.g. a CSV body). contentType is e.g. "text/csv".
func (r *AgentsResource) Import(ctx context.Context, contentType string, body io.Reader, params *managerapi.ImportAgentsParams) (*managerapi.AgentImportValidationResults, error) {
	resp, err := r.gc.ImportAgentsWithBodyWithResponse(ctx, params, contentType, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ValidateImport validates an agent import upload stream without applying it. contentType is e.g. "text/csv".
func (r *AgentsResource) ValidateImport(ctx context.Context, contentType string, body io.Reader) (*managerapi.AgentImportValidationResults, error) {
	resp, err := r.gc.ValidateAgentImportWithBodyWithResponse(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetImportJob returns an asynchronous agent-import job by id.
func (r *AgentsResource) GetImportJob(ctx context.Context, id string) (*managerapi.AgentImportJobItemResponse, error) {
	resp, err := r.gc.GetAgentImportJobWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Logs returns an iterator over a single agent's log entries, auto-paginating across pages.
//
// The Page field of params is managed by the auto-paginator, so leave it unset.
func (r *AgentsResource) Logs(ctx context.Context, id string, params managerapi.ListAgentLogsParams) iter.Seq2[managerapi.AgentLogEntry, error] {
	return func(yield func(managerapi.AgentLogEntry, error) bool) {
		var zero managerapi.AgentLogEntry
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListAgentLogsWithResponse(ctx, id, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, e := range data.Items {
				if !yield(e, nil) {
					return
				}
			}
			if len(data.Items) == 0 || data.Pagination == nil || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// AllLogs returns an iterator over every agent's log entries, auto-paginating across pages.
//
// The Page field of params is managed by the auto-paginator, so leave it unset.
func (r *AgentsResource) AllLogs(ctx context.Context, params managerapi.ListAllAgentLogsParams) iter.Seq2[managerapi.AgentLogEntry, error] {
	return func(yield func(managerapi.AgentLogEntry, error) bool) {
		var zero managerapi.AgentLogEntry
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListAllAgentLogsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, e := range data.Items {
				if !yield(e, nil) {
					return
				}
			}
			if len(data.Items) == 0 || data.Pagination == nil || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// Push sends a push notification / event to an agent.
func (r *AgentsResource) Push(ctx context.Context, body managerapi.AgentPushRequest) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.PushToAgentWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdatePassword updates the password of an agent by id.
func (r *AgentsResource) UpdatePassword(ctx context.Context, id string, body managerapi.AgentPasswordUpdateRequest) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.UpdateAgentPasswordWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AgentGroupsResource is the agent-groups namespace (/api/v2/agents/groups).
type AgentGroupsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over agent groups, auto-paginating across pages.
func (r *AgentGroupsResource) List(ctx context.Context, pageSize int) iter.Seq2[managerapi.AgentGroup, error] {
	return func(yield func(managerapi.AgentGroup, error) bool) {
		var zero managerapi.AgentGroup

		params := &managerapi.ListAgentGroupsParams{}
		if pageSize > 0 {
			params.Max = &pageSize
		}

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListAgentGroupsWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, g := range data.Items {
				if !yield(g, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// ListAll collects every agent group into a slice (convenience over List).
func (r *AgentGroupsResource) ListAll(ctx context.Context, pageSize int) ([]managerapi.AgentGroup, error) {
	var groups []managerapi.AgentGroup
	for g, err := range r.List(ctx, pageSize) {
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Create creates an agent group.
func (r *AgentGroupsResource) Create(ctx context.Context, body managerapi.RestCreateAgentGroup) (*managerapi.AgentGroupItemResponse, error) {
	resp, err := r.gc.CreateAgentGroupWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns an agent group by id.
func (r *AgentGroupsResource) Get(ctx context.Context, id string) (*managerapi.AgentGroupItemResponse, error) {
	resp, err := r.gc.GetAgentGroupWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates an agent group.
func (r *AgentGroupsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateAgentGroup) (*managerapi.AgentGroupItemResponse, error) {
	resp, err := r.gc.UpdateAgentGroupWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes an agent group by id.
func (r *AgentGroupsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteAgentGroupWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// AddAgent adds an agent (by id) to a group.
func (r *AgentGroupsResource) AddAgent(ctx context.Context, groupID, agentID string) (*managerapi.AgentGroupAdditionResponse, error) {
	id, err := uuid.Parse(agentID)
	if err != nil {
		return nil, err
	}
	resp, err := r.gc.AddAgentToGroupWithResponse(ctx, groupID, managerapi.AddAgentToGroupRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// RemoveAgent removes an agent (by id) from a group.
func (r *AgentGroupsResource) RemoveAgent(ctx context.Context, groupID, agentID string) (*managerapi.AgentGroupItemResponse, error) {
	resp, err := r.gc.RemoveAgentFromGroupWithResponse(ctx, groupID, agentID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListAgents returns an iterator over the agents in a group, auto-paginating across pages.
//
// The Page field of params is managed by the auto-paginator, so leave it unset.
func (r *AgentGroupsResource) ListAgents(ctx context.Context, groupID string, params managerapi.ListAgentsInGroupParams) iter.Seq2[managerapi.Agent, error] {
	return func(yield func(managerapi.Agent, error) bool) {
		var zero managerapi.Agent
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListAgentsInGroupWithResponse(ctx, groupID, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, a := range data.Items {
				if !yield(a, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// BulkDelete deletes the given agent groups by id.
func (r *AgentGroupsResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	uuids := make([]managerapi.ObjectUuid, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsed)
	}
	resp, err := r.gc.BulkDeleteAgentGroupsWithResponse(ctx, managerapi.BulkDeleteAgentGroupsJSONRequestBody{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
