package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// ConversationsResource is the conversations namespace (/api/v2/conversations), with events and
// session variables.
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type ConversationsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over conversations, auto-paginating across pages.
func (r *ConversationsResource) List(ctx context.Context, params managerapi.ListConversationsParams) iter.Seq2[managerapi.Conversation, error] {
	return func(yield func(managerapi.Conversation, error) bool) {
		var zero managerapi.Conversation
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListConversationsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, c := range data.Items {
				if !yield(c, nil) {
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

// ListAll collects every conversation into a slice (convenience over List).
func (r *ConversationsResource) ListAll(ctx context.Context, params managerapi.ListConversationsParams) ([]managerapi.Conversation, error) {
	var out []managerapi.Conversation
	for c, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Create creates a conversation.
func (r *ConversationsResource) Create(ctx context.Context, body managerapi.RestCreateConversation) (*managerapi.ConversationItemResponse, error) {
	resp, err := r.gc.CreateConversationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a conversation by id.
func (r *ConversationsResource) Get(ctx context.Context, id string) (*managerapi.ConversationItemResponse, error) {
	resp, err := r.gc.GetConversationWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a conversation.
func (r *ConversationsResource) Update(ctx context.Context, id string, body managerapi.RestUpdateConversation) (*managerapi.ConversationItemResponse, error) {
	resp, err := r.gc.UpdateConversationWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a conversation.
func (r *ConversationsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteConversationWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Events returns a conversation's events.
func (r *ConversationsResource) Events(ctx context.Context, conversationID string) ([]managerapi.ConversationEvent, error) {
	resp, err := r.gc.ListConversationEventsWithResponse(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// GetEvent returns a single conversation event.
func (r *ConversationsResource) GetEvent(ctx context.Context, conversationID, id string) (*managerapi.ConversationEventItemResponse, error) {
	resp, err := r.gc.GetConversationEventWithResponse(ctx, conversationID, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetSession returns a conversation's session variables.
func (r *ConversationsResource) GetSession(ctx context.Context, conversationID string) (*managerapi.ConversationSessionVariablesItemResponse, error) {
	resp, err := r.gc.GetConversationSessionWithResponse(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateSession updates a conversation's session variables.
func (r *ConversationsResource) UpdateSession(ctx context.Context, conversationID string, variables managerapi.ConversationSessionVariables) (*managerapi.ConversationSessionVariablesItemResponse, error) {
	resp, err := r.gc.UpdateConversationSessionWithResponse(ctx, conversationID, variables)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
