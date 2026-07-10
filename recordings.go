package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// RecordingsResource is the call-recordings namespace (/api/v2/recordings): listing, starting,
// fetching, updating, deleting, bulk actions and flagging of call recordings.
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type RecordingsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over recordings, auto-paginating across pages.
//
//	for rec, err := range mgr.Recordings.List(ctx, managerapi.ListRecordingsParams{}) {
//	    if err != nil { return err }
//	    fmt.Println(rec.Id)
//	}
func (r *RecordingsResource) List(ctx context.Context, params managerapi.ListRecordingsParams) iter.Seq2[managerapi.Recording, error] {
	return func(yield func(managerapi.Recording, error) bool) {
		var zero managerapi.Recording
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListRecordingsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, rec := range data.Items {
				if !yield(rec, nil) {
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

// ListAll collects every recording into a slice (convenience over List).
func (r *RecordingsResource) ListAll(ctx context.Context, params managerapi.ListRecordingsParams) ([]managerapi.Recording, error) {
	var out []managerapi.Recording
	for rec, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

// Start starts a recording for a call.
func (r *RecordingsResource) Start(ctx context.Context, body managerapi.RecordingStartRequest) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.StartRecordingWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Get returns a recording by id.
func (r *RecordingsResource) Get(ctx context.Context, id string) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.GetRecordingWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a recording (e.g. its tags) by id.
func (r *RecordingsResource) Update(ctx context.Context, id string, body managerapi.RecordingUpdateBody) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.UpdateRecordingWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a recording by id.
func (r *RecordingsResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteRecordingWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// BulkAction applies a bulk action (e.g. "delete", "flag") to several recordings by id.
func (r *RecordingsResource) BulkAction(ctx context.Context, action string, ids []string) (*managerapi.BulkActionResponse, error) {
	uuids, err := toUUIDs(ids)
	if err != nil {
		return nil, err
	}
	resp, err := r.gc.BulkRecordingActionWithResponse(ctx, action, managerapi.BulkIdsRequest{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// GetFlag returns a recording's flag state by id.
func (r *RecordingsResource) GetFlag(ctx context.Context, id string) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.GetRecordingFlagWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Flag flags a recording by id.
func (r *RecordingsResource) Flag(ctx context.Context, id string) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.FlagRecordingWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Unflag removes the flag from a recording by id.
func (r *RecordingsResource) Unflag(ctx context.Context, id string) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.UnflagRecordingWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ToggleFlag toggles the flag on a recording by id.
func (r *RecordingsResource) ToggleFlag(ctx context.Context, id string) (*managerapi.RecordingItemResponse, error) {
	resp, err := r.gc.ToggleRecordingFlagWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
