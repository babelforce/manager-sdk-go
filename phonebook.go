package manager

import (
	"context"
	"io"
	"iter"

	"github.com/google/uuid"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// PhonebookResource is the phonebook-entries namespace (/api/v2/phonebook), with bulk CSV
// download/upload.
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type PhonebookResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over phonebook entries, auto-paginating across pages.
func (r *PhonebookResource) List(ctx context.Context, params managerapi.ListPhonebookEntrysParams) iter.Seq2[managerapi.PhonebookEntry, error] {
	return func(yield func(managerapi.PhonebookEntry, error) bool) {
		var zero managerapi.PhonebookEntry
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListPhonebookEntrysWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
		}
	}
}

// ListAll collects every phonebook entry into a slice (convenience over List).
func (r *PhonebookResource) ListAll(ctx context.Context, params managerapi.ListPhonebookEntrysParams) ([]managerapi.PhonebookEntry, error) {
	var out []managerapi.PhonebookEntry
	for e, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// Create creates a phonebook entry.
func (r *PhonebookResource) Create(ctx context.Context, body managerapi.RestCreatePhonebookEntry) (*managerapi.PhonebookEntryItemResponse, error) {
	resp, err := r.gc.CreatePhonebookEntryWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a phonebook entry by id.
func (r *PhonebookResource) Get(ctx context.Context, id string) (*managerapi.PhonebookEntryItemResponse, error) {
	resp, err := r.gc.GetPhonebookEntryWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a phonebook entry.
func (r *PhonebookResource) Update(ctx context.Context, id string, body managerapi.RestUpdatePhonebookEntry) (*managerapi.PhonebookEntryItemResponse, error) {
	resp, err := r.gc.UpdatePhonebookEntryWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a phonebook entry.
func (r *PhonebookResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeletePhonebookEntryWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// BulkDelete deletes the given phonebook entries by id.
func (r *PhonebookResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	uuids := make([]managerapi.ObjectUuid, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsed)
	}
	resp, err := r.gc.BulkDeletePhonebookEntriesWithResponse(ctx, managerapi.BulkDeletePhonebookEntriesJSONRequestBody{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Download returns all phonebook entries as a raw CSV export (bulk).
func (r *PhonebookResource) Download(ctx context.Context) ([]byte, error) {
	resp, err := r.gc.DownloadPhonebookEntriesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Upload imports phonebook entries from a CSV stream (bulk). contentType is e.g. "text/csv".
func (r *PhonebookResource) Upload(ctx context.Context, contentType string, body io.Reader) error {
	resp, err := r.gc.UploadPhonebookEntriesWithBodyWithResponse(ctx, nil, contentType, body)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
