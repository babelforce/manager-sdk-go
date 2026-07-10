package manager

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// FilesResource is the stored-files namespace (/api/v2/files) — listing, fetching, downloading,
// and (bulk) deleting stored files (recordings, prompts, backups, …).
type FilesResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over stored files, auto-paginating across pages. Page is managed by the
// iterator; other filters on params (Type, State, Filename, Q, Sort, Order, Max) are honoured.
func (r *FilesResource) List(ctx context.Context, params managerapi.ListFilesParams) iter.Seq2[managerapi.StoredFile, error] {
	return func(yield func(managerapi.StoredFile, error) bool) {
		var zero managerapi.StoredFile

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListFilesWithResponse(ctx, &params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, f := range data.Items {
				if !yield(f, nil) {
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

// ListAll collects every stored file into a slice (convenience over List).
func (r *FilesResource) ListAll(ctx context.Context, params managerapi.ListFilesParams) ([]managerapi.StoredFile, error) {
	var files []managerapi.StoredFile
	for f, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// ListByType lists the stored files of a given storage type.
func (r *FilesResource) ListByType(ctx context.Context, fileType managerapi.StorageType) ([]managerapi.StoredFile, error) {
	resp, err := r.gc.ListFilesByTypeWithResponse(ctx, fileType)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Backups lists the account's backup files.
func (r *FilesResource) Backups(ctx context.Context) ([]managerapi.StoredFile, error) {
	resp, err := r.gc.ListBackupFilesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Recordings lists the account's recording files.
func (r *FilesResource) Recordings(ctx context.Context) ([]managerapi.StoredFile, error) {
	resp, err := r.gc.ListRecordingFilesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Prompts lists the account's prompt files.
func (r *FilesResource) Prompts(ctx context.Context) ([]managerapi.StoredFile, error) {
	resp, err := r.gc.ListPromptFilesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Get returns a stored file's metadata by id.
func (r *FilesResource) Get(ctx context.Context, id string) (*managerapi.StoredFileItemResponse, error) {
	resp, err := r.gc.GetFileWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a stored file by id.
func (r *FilesResource) Delete(ctx context.Context, id string) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DeleteFileWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Download returns the raw bytes of a stored file.
func (r *FilesResource) Download(ctx context.Context, id string) ([]byte, error) {
	resp, err := r.gc.DownloadFileWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// BulkDelete deletes several stored files by id.
func (r *FilesResource) BulkDelete(ctx context.Context, ids []string) (*managerapi.DefaultV2MessageResponse, error) {
	uuids, err := toUUIDs(ids)
	if err != nil {
		return nil, err
	}
	resp, err := r.gc.BulkDeleteFilesWithResponse(ctx, managerapi.FileBulkDeleteRequest{Ids: uuids})
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkDownload downloads several stored files as a ZIP (ids passed as a query parameter).
func (r *FilesResource) BulkDownload(ctx context.Context, ids []string) ([]byte, error) {
	resp, err := r.gc.GetBulkFileDownloadWithResponse(ctx, &managerapi.GetBulkFileDownloadParams{Ids: strings.Join(ids, ",")})
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// BulkDownloadPost downloads several stored files as a ZIP (ids in the request body).
func (r *FilesResource) BulkDownloadPost(ctx context.Context, ids []string) ([]byte, error) {
	uuids, err := toUUIDs(ids)
	if err != nil {
		return nil, err
	}
	resp, err := r.gc.PostBulkFileDownloadWithResponse(ctx, managerapi.FileBulkDownloadRequest{Ids: uuids})
	if err != nil {
		return nil, err
	}
	if err := resultVoid(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func toUUIDs(ids []string) ([]openapi_types.UUID, error) {
	out := make([]openapi_types.UUID, len(ids))
	for i, s := range ids {
		u, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid file id %q: %w", s, err)
		}
		out[i] = u
	}
	return out, nil
}
