package manager

import (
	"context"
	"io"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// PromptsResource is the audio-prompts namespace (/api/v2/prompts).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type PromptsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over prompts, auto-paginating across pages.
func (r *PromptsResource) List(ctx context.Context, params managerapi.ListPromptsParams) iter.Seq2[managerapi.Prompt, error] {
	return func(yield func(managerapi.Prompt, error) bool) {
		var zero managerapi.Prompt
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListPromptsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, pr := range data.Items {
				if !yield(pr, nil) {
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

// ListAll collects every prompt into a slice (convenience over List).
func (r *PromptsResource) ListAll(ctx context.Context, params managerapi.ListPromptsParams) ([]managerapi.Prompt, error) {
	var out []managerapi.Prompt
	for pr, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, pr)
	}
	return out, nil
}

// Get returns a prompt by id.
func (r *PromptsResource) Get(ctx context.Context, id string) (*managerapi.PromptItemResponse, error) {
	resp, err := r.gc.GetPromptWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Uses lists the objects (e.g. applications) that reference a prompt
// (GET /api/v2/prompts/{id}/uses).
func (r *PromptsResource) Uses(ctx context.Context, id string) (*managerapi.ObjectListResponse, error) {
	resp, err := r.gc.GetPromptUsesWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Upload uploads a new audio prompt from a stream. contentType is e.g. "audio/wav".
func (r *PromptsResource) Upload(ctx context.Context, contentType string, body io.Reader) (*managerapi.PromptItemResponse, error) {
	resp, err := r.gc.UploadPromptWithBodyWithResponse(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Update updates a prompt's metadata.
func (r *PromptsResource) Update(ctx context.Context, id string, body managerapi.RestUpdatePrompt) (*managerapi.PromptItemResponse, error) {
	resp, err := r.gc.UpdatePromptWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a prompt.
func (r *PromptsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeletePromptWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
