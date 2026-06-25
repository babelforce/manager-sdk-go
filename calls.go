package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// CallsResource is the call namespace (/api/v2/calls): call reporting plus call control.
type CallsResource struct {
	// Reporting is the call-reporting sub-namespace (/api/v2/calls/reporting).
	Reporting *ReportingResource

	gc *managerapi.ClientWithResponses
}

// Get returns a single call by id.
func (r *CallsResource) Get(ctx context.Context, id string) (*managerapi.CallItemResponse, error) {
	resp, err := r.gc.GetCallWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Hangup hangs up a live call and returns the updated call.
func (r *CallsResource) Hangup(ctx context.Context, id string) (*managerapi.CallItemResponse, error) {
	resp, err := r.gc.HangupCallWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// CreateTestCall creates an inbound test call.
func (r *CallsResource) CreateTestCall(ctx context.Context, body managerapi.CreateTestCall) (*managerapi.CallItemResponse, error) {
	resp, err := r.gc.CreateInboundTestCallWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// SetSessionVariables sets session variables on a call.
func (r *CallsResource) SetSessionVariables(ctx context.Context, id string, variables managerapi.SetCallSessionVariablesRequest) (*managerapi.SetCallSessionVariablesResponse, error) {
	resp, err := r.gc.SetCallSessionVariablesWithResponse(ctx, id, variables)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Cancel cancels a queued or ringing call and returns the updated call.
func (r *CallsResource) Cancel(ctx context.Context, id string) (*managerapi.CallItemResponse, error) {
	resp, err := r.gc.CancelCallWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListQueued returns an iterator over the calls currently queued in a queue,
// auto-paginating across pages. The Page field is managed by the auto-paginator,
// so leave it unset.
func (r *CallsResource) ListQueued(ctx context.Context, queueID string, params managerapi.ListQueuedCallsParams) iter.Seq2[managerapi.QueuedCall, error] {
	return func(yield func(managerapi.QueuedCall, error) bool) {
		var zero managerapi.QueuedCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListQueuedCallsWithResponse(ctx, queueID, &p)
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
			if len(data.Items) == 0 || data.Pagination == nil || data.Pagination.Current >= data.Pagination.Pages {
				return
			}
			page = data.Pagination.Current + 1
		}
	}
}

// QueueCallback enqueues a callback in a queue and returns the queued callback.
func (r *CallsResource) QueueCallback(ctx context.Context, queueID string, body managerapi.QueueCallbackRequest) (*managerapi.QueueCallbackResponse, error) {
	resp, err := r.gc.QueueCallbackWithResponse(ctx, queueID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ReportingResource is the call-reporting namespace (/api/v2/calls/reporting).
//
// The list methods take the generated parameter structs directly (every filter is an optional
// pointer field); the Page field is managed by the auto-paginator, so leave it unset.
type ReportingResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over the detailed call report, auto-paginating across pages.
//
//	for call, err := range mgr.Calls.Reporting.List(ctx, managerapi.ListReportingCallsParams{}) {
//	    if err != nil { return err }
//	    fmt.Println(call.Id)
//	}
func (r *ReportingResource) List(ctx context.Context, params managerapi.ListReportingCallsParams) iter.Seq2[managerapi.Call, error] {
	return func(yield func(managerapi.Call, error) bool) {
		var zero managerapi.Call
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListReportingCallsWithResponse(ctx, &p)
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

// ListAll collects every call from the detailed report into a slice (convenience over List).
func (r *ReportingResource) ListAll(ctx context.Context, params managerapi.ListReportingCallsParams) ([]managerapi.Call, error) {
	var calls []managerapi.Call
	for c, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, nil
}

// Simple returns an iterator over the simple call report across all report types
// (/api/v2/calls/reporting/simple), auto-paginating across pages.
func (r *ReportingResource) Simple(ctx context.Context, params managerapi.ListAllSimpleReportingCallsParams) iter.Seq2[managerapi.ReportingCall, error] {
	return func(yield func(managerapi.ReportingCall, error) bool) {
		var zero managerapi.ReportingCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListAllSimpleReportingCallsWithResponse(ctx, &p)
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

// SimpleAll collects every call from the simple report into a slice (convenience over Simple).
func (r *ReportingResource) SimpleAll(ctx context.Context, params managerapi.ListAllSimpleReportingCallsParams) ([]managerapi.ReportingCall, error) {
	var calls []managerapi.ReportingCall
	for c, err := range r.Simple(ctx, params) {
		if err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, nil
}

// InboundSimple returns an iterator over the simple inbound call report
// (/api/v2/calls/reporting/simple/inbound), auto-paginating across pages.
func (r *ReportingResource) InboundSimple(ctx context.Context, params managerapi.ListInboundSimpleReportingCallsParams) iter.Seq2[managerapi.ReportingCall, error] {
	return func(yield func(managerapi.ReportingCall, error) bool) {
		var zero managerapi.ReportingCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListInboundSimpleReportingCallsWithResponse(ctx, &p)
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

// InboundSimpleAll collects every call from the simple inbound report into a slice
// (convenience over InboundSimple).
func (r *ReportingResource) InboundSimpleAll(ctx context.Context, params managerapi.ListInboundSimpleReportingCallsParams) ([]managerapi.ReportingCall, error) {
	var calls []managerapi.ReportingCall
	for c, err := range r.InboundSimple(ctx, params) {
		if err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, nil
}
