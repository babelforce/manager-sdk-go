package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// CallsResource is the call namespace (/api/v2/calls). Today it exposes call reporting; call
// control may follow.
type CallsResource struct {
	// Reporting is the call-reporting sub-namespace (/api/v2/calls/reporting).
	Reporting *ReportingResource
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

// SimpleByType returns an iterator over the simple call report for a single report type
// (/api/v2/calls/reporting/simple/{reportType}), auto-paginating across pages.
func (r *ReportingResource) SimpleByType(ctx context.Context, reportType managerapi.SimpleReportingReportType, params managerapi.ListSimpleReportingCallsParams) iter.Seq2[managerapi.ReportingCall, error] {
	return func(yield func(managerapi.ReportingCall, error) bool) {
		var zero managerapi.ReportingCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListSimpleReportingCallsWithResponse(ctx, reportType, &p)
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

// SimpleAllByType collects every call from a single report type into a slice.
func (r *ReportingResource) SimpleAllByType(ctx context.Context, reportType managerapi.SimpleReportingReportType, params managerapi.ListSimpleReportingCallsParams) ([]managerapi.ReportingCall, error) {
	var calls []managerapi.ReportingCall
	for c, err := range r.SimpleByType(ctx, reportType, params) {
		if err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, nil
}
