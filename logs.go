package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// LogsResource is the logs namespace: request audit logs (/api/v2/audit) and live logs (/api/v2/logs).
type LogsResource struct {
	gc *managerapi.ClientWithResponses
}

// Audit returns an iterator over request audit-log entries, auto-paginating across pages.
func (r *LogsResource) Audit(ctx context.Context, params managerapi.ListAuditLogsParams) iter.Seq2[managerapi.AuditLog, error] {
	return func(yield func(managerapi.AuditLog, error) bool) {
		var zero managerapi.AuditLog
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListAuditLogsWithResponse(ctx, &p)
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

// AuditAll collects every audit-log entry into a slice (convenience over Audit).
func (r *LogsResource) AuditAll(ctx context.Context, params managerapi.ListAuditLogsParams) ([]managerapi.AuditLog, error) {
	var out []managerapi.AuditLog
	for a, err := range r.Audit(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// Live returns the current live logs.
func (r *LogsResource) Live(ctx context.Context) ([]managerapi.LiveLog, error) {
	resp, err := r.gc.ListLiveLogsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// EnableLive turns on live logging and returns the acknowledgement message.
func (r *LogsResource) EnableLive(ctx context.Context) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.EnableLiveLoggingWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DisableLive turns off live logging and returns the acknowledgement message.
func (r *LogsResource) DisableLive(ctx context.Context) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.DisableLiveLoggingWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Write appends an entry to the live log and returns the created item.
func (r *LogsResource) Write(ctx context.Context, body managerapi.WriteLogRequest) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.WriteLogWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
