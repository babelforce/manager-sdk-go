package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// DialerResource is the dialer namespace (/api/v2/dialer): runtime info, queue control,
// simple call reporting, plus the dialer-behaviours sub-namespace.
type DialerResource struct {
	gc *managerapi.ClientWithResponses
	// Behaviours is the dialer-behaviours sub-namespace (/api/v2/outbound/dialer-behaviours).
	Behaviours *DialerBehavioursResource
}

// Info returns the current dialer runtime information.
func (r *DialerResource) Info(ctx context.Context) (*managerapi.GenericItemResponse, error) {
	resp, err := r.gc.GetDialerInfoWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Flush flushes queued dialer tasks. Pass a task id to flush a single task, or all=true to
// flush every queued task.
func (r *DialerResource) Flush(ctx context.Context, id *string, all *bool) (*managerapi.DefaultV2MessageResponse, error) {
	params := &managerapi.FlushDialerParams{Id: id, All: all}
	resp, err := r.gc.FlushDialerWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// SimpleReporting returns an iterator over the dialer's simple call report
// (/api/v2/calls/reporting/simple/dialer), auto-paginating across pages.
//
//	for call, err := range mgr.Dialer.SimpleReporting(ctx, managerapi.ListDialerSimpleReportingCallsParams{}) {
//	    if err != nil { return err }
//	    fmt.Println(call.Id)
//	}
func (r *DialerResource) SimpleReporting(ctx context.Context, params managerapi.ListDialerSimpleReportingCallsParams) iter.Seq2[managerapi.ReportingCall, error] {
	return func(yield func(managerapi.ReportingCall, error) bool) {
		var zero managerapi.ReportingCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListDialerSimpleReportingCallsWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || page >= data.Pagination.Pages {
				return
			}
			page++
		}
	}
}

// SimpleReportingAll collects every call from the dialer's simple report into a slice
// (convenience over SimpleReporting).
func (r *DialerResource) SimpleReportingAll(ctx context.Context, params managerapi.ListDialerSimpleReportingCallsParams) ([]managerapi.ReportingCall, error) {
	var calls []managerapi.ReportingCall
	for c, err := range r.SimpleReporting(ctx, params) {
		if err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, nil
}

// DialerBehavioursResource is the dialer-behaviours namespace (/api/v2/outbound/dialer-behaviours).
type DialerBehavioursResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over dialer behaviours, auto-paginating across pages.
func (r *DialerBehavioursResource) List(ctx context.Context, pageSize int) iter.Seq2[managerapi.DialerBehaviour, error] {
	return func(yield func(managerapi.DialerBehaviour, error) bool) {
		var zero managerapi.DialerBehaviour

		params := &managerapi.ListDialerBehavioursParams{}
		if pageSize > 0 {
			params.Max = &pageSize
		}

		page := 1
		for {
			params.Page = &page
			resp, err := r.gc.ListDialerBehavioursWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, b := range data.Items {
				if !yield(b, nil) {
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

// ListAll collects every dialer behaviour into a slice (convenience over List).
func (r *DialerBehavioursResource) ListAll(ctx context.Context, pageSize int) ([]managerapi.DialerBehaviour, error) {
	var behaviours []managerapi.DialerBehaviour
	for b, err := range r.List(ctx, pageSize) {
		if err != nil {
			return nil, err
		}
		behaviours = append(behaviours, b)
	}
	return behaviours, nil
}

// Create creates a dialer behaviour.
func (r *DialerBehavioursResource) Create(ctx context.Context, body managerapi.DialerBehaviourWriteBody) (*managerapi.DialerBehaviourItemResponse, error) {
	resp, err := r.gc.CreateDialerBehaviourWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Get returns a dialer behaviour by id.
func (r *DialerBehavioursResource) Get(ctx context.Context, id string) (*managerapi.DialerBehaviourItemResponse, error) {
	resp, err := r.gc.GetDialerBehaviourWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a dialer behaviour.
func (r *DialerBehavioursResource) Update(ctx context.Context, id string, body managerapi.DialerBehaviourWriteBody) (*managerapi.DialerBehaviourItemResponse, error) {
	resp, err := r.gc.UpdateDialerBehaviourWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a dialer behaviour by id.
func (r *DialerBehavioursResource) Delete(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteDialerBehaviourWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
