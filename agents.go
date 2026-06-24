package manager

import (
	"context"
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
			if len(data.Items) == 0 || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
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
			if len(data.Items) == 0 || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
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
