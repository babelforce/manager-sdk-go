package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// QueuesResource is the queues namespace (/api/v2/queues), with a nested Selections sub-resource.
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type QueuesResource struct {
	gc *managerapi.ClientWithResponses
	// Selections is the queue-selections sub-namespace (/api/v2/queues/{queueId}/selections).
	Selections *QueueSelectionsResource
}

// List returns an iterator over queues, auto-paginating across pages.
func (r *QueuesResource) List(ctx context.Context, params managerapi.ListQueuesParams) iter.Seq2[managerapi.Queue, error] {
	return func(yield func(managerapi.Queue, error) bool) {
		var zero managerapi.Queue
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListQueuesWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, q := range data.Items {
				if !yield(q, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// ListAll collects every queue into a slice (convenience over List).
func (r *QueuesResource) ListAll(ctx context.Context, params managerapi.ListQueuesParams) ([]managerapi.Queue, error) {
	var out []managerapi.Queue
	for q, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, nil
}

// Create creates a queue.
func (r *QueuesResource) Create(ctx context.Context, body managerapi.RestCreateQueue) (*managerapi.QueueItemResponse, error) {
	resp, err := r.gc.CreateQueueWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a queue by id.
func (r *QueuesResource) Get(ctx context.Context, id string) (*managerapi.QueueItemResponse, error) {
	resp, err := r.gc.GetQueueWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a queue.
func (r *QueuesResource) Update(ctx context.Context, id string, body managerapi.RestUpdateQueue) (*managerapi.QueueItemResponse, error) {
	resp, err := r.gc.UpdateQueueWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a queue.
func (r *QueuesResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteQueueWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// ListTriggers lists the triggers attached to a queue.
func (r *QueuesResource) ListTriggers(ctx context.Context, queueID string) (*managerapi.QueueTriggerListResponse, error) {
	resp, err := r.gc.ListQueueTriggersWithResponse(ctx, queueID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkUpdate updates multiple queues in a single request (each item must include its id).
func (r *QueuesResource) BulkUpdate(ctx context.Context, body managerapi.QueueBulkUpdateRequest) (*managerapi.QueueListResponse, error) {
	resp, err := r.gc.BulkUpdateQueuesWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GlobalSelections lists queue selections across all queues.
func (r *QueuesResource) GlobalSelections(ctx context.Context) (*managerapi.GlobalQueueSelectionListResponse, error) {
	resp, err := r.gc.ListGlobalQueueSelectionsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// QueueSelectionsResource is queue selections (routing rules) and their agent/group/tag membership.
type QueueSelectionsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over a queue's selections, auto-paginating across pages.
func (r *QueueSelectionsResource) List(ctx context.Context, queueID string, params managerapi.ListQueueSelectionsParams) iter.Seq2[managerapi.QueueSelection, error] {
	return func(yield func(managerapi.QueueSelection, error) bool) {
		var zero managerapi.QueueSelection
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListQueueSelectionsWithResponse(ctx, queueID, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, s := range data.Items {
				if !yield(s, nil) {
					return
				}
			}
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// ListAll collects every selection for a queue into a slice (convenience over List).
func (r *QueueSelectionsResource) ListAll(ctx context.Context, queueID string, params managerapi.ListQueueSelectionsParams) ([]managerapi.QueueSelection, error) {
	var out []managerapi.QueueSelection
	for s, err := range r.List(ctx, queueID, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// SetPriority sets the priority of the given selections on a queue.
func (r *QueueSelectionsResource) SetPriority(ctx context.Context, queueID string, items []managerapi.QueueSelectionPriorityItem) (*managerapi.QueueSelectionListResponse, error) {
	resp, err := r.gc.SetQueueSelectionsPriorityWithResponse(ctx, queueID, items)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Create creates a selection on a queue.
func (r *QueueSelectionsResource) Create(ctx context.Context, queueID string, body managerapi.RestCreateQueueSelection) (*managerapi.QueueSelectionItemResponse, error) {
	resp, err := r.gc.CreateQueueSelectionWithResponse(ctx, queueID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a selection.
func (r *QueueSelectionsResource) Get(ctx context.Context, queueID, id string) (*managerapi.QueueSelectionItemResponse, error) {
	resp, err := r.gc.GetQueueSelectionWithResponse(ctx, queueID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a selection.
func (r *QueueSelectionsResource) Update(ctx context.Context, queueID, id string, body managerapi.RestUpdateQueueSelection) (*managerapi.QueueSelectionItemResponse, error) {
	resp, err := r.gc.UpdateQueueSelectionWithResponse(ctx, queueID, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a selection.
func (r *QueueSelectionsResource) Delete(ctx context.Context, queueID, id string) error {
	resp, err := r.gc.DeleteQueueSelectionWithResponse(ctx, queueID, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// SelectAgents resolves the agents currently selected for a queue.
func (r *QueueSelectionsResource) SelectAgents(ctx context.Context, queueID string) (*managerapi.QueueSelectionResponse, error) {
	resp, err := r.gc.GetAgentsForQueueSelectionWithResponse(ctx, queueID, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddAgent adds an agent to a selection.
func (r *QueueSelectionsResource) AddAgent(ctx context.Context, queueID, selectionID, agentID string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.AddAgentToQueueSelectionWithResponse(ctx, queueID, selectionID, managerapi.QueueSelectionAddition{Id: &agentID})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// RemoveAgent removes an agent from a selection.
func (r *QueueSelectionsResource) RemoveAgent(ctx context.Context, queueID, selectionID, id string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.RemoveAgentFromQueueSelectionWithResponse(ctx, queueID, selectionID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddGroup adds an agent group to a selection.
func (r *QueueSelectionsResource) AddGroup(ctx context.Context, queueID, selectionID, groupID string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.AddGroupToQueueSelectionWithResponse(ctx, queueID, selectionID, managerapi.QueueSelectionAddition{Id: &groupID})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// RemoveGroup removes an agent group from a selection.
func (r *QueueSelectionsResource) RemoveGroup(ctx context.Context, queueID, selectionID, id string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.RemoveGroupFromQueueSelectionWithResponse(ctx, queueID, selectionID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddTag adds a tag to a selection.
func (r *QueueSelectionsResource) AddTag(ctx context.Context, queueID, selectionID, tagID string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.AddTagToQueueSelectionWithResponse(ctx, queueID, selectionID, managerapi.QueueSelectionAddition{Id: &tagID})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// RemoveTag removes a tag from a selection.
func (r *QueueSelectionsResource) RemoveTag(ctx context.Context, queueID, selectionID, id string) (*managerapi.QueueSelectionModificationResponse, error) {
	resp, err := r.gc.RemoveTagFromQueueSelectionWithResponse(ctx, queueID, selectionID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
