package manager

import (
	"bytes"
	"context"
	"io"
	"iter"
	"mime/multipart"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// OutboundResource is the outbound dialer-lists namespace (/api/v2/outbound/lists), with leads.
type OutboundResource struct {
	gc *managerapi.ClientWithResponses
}

// Lists returns all outbound lists.
func (r *OutboundResource) Lists(ctx context.Context) ([]managerapi.OutboundList, error) {
	resp, err := r.gc.ListOutboundListsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// CreateList creates an outbound list.
func (r *OutboundResource) CreateList(ctx context.Context, body managerapi.CreateOutboundListRequest) (*managerapi.OutboundListItemResponse, error) {
	resp, err := r.gc.CreateOutboundListWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ClearList removes all leads from an outbound list and returns the (now empty) list.
func (r *OutboundResource) ClearList(ctx context.Context, id string) (*managerapi.OutboundListItemResponse, error) {
	resp, err := r.gc.ClearOutboundListWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AddLead adds a lead to an outbound list.
func (r *OutboundResource) AddLead(ctx context.Context, listID string, body managerapi.AddOutboundLeadRequest) (*managerapi.OutboundLeadItemResponse, error) {
	resp, err := r.gc.AddOutboundLeadWithResponse(ctx, listID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateLead updates a lead in an outbound list.
func (r *OutboundResource) UpdateLead(ctx context.Context, listID, leadID string, body managerapi.AddOutboundLeadRequest) (*managerapi.OutboundLeadItemResponse, error) {
	resp, err := r.gc.UpdateOutboundLeadWithResponse(ctx, listID, leadID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeleteLead removes a lead from an outbound list.
func (r *OutboundResource) DeleteLead(ctx context.Context, listID, leadID string) error {
	resp, err := r.gc.DeleteOutboundLeadWithResponse(ctx, listID, leadID)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// GetLead returns a single lead from an outbound list.
//
// Wraps generated GetLeadInList.
func (r *OutboundResource) GetLead(ctx context.Context, listID, leadID string) (*managerapi.LeadItemResponse, error) {
	resp, err := r.gc.GetLeadInListWithResponse(ctx, listID, leadID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListLeads returns the leads of an outbound list (single page; optionally filtered by status).
//
// Wraps generated ListLeadsInList.
func (r *OutboundResource) ListLeads(ctx context.Context, listID string, params managerapi.ListLeadsInListParams) (*managerapi.PaginatedLeadResponse, error) {
	resp, err := r.gc.ListLeadsInListWithResponse(ctx, listID, &params)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// CreateAgentCall starts an outbound call from an agent (by id) to a destination.
//
// Wraps generated CreateAgentOutboundCall.
func (r *OutboundResource) CreateAgentCall(ctx context.Context, id string, body managerapi.AgentOutboundCallRequest) (*managerapi.CallItemResponse, error) {
	resp, err := r.gc.CreateAgentOutboundCallWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// SimpleReporting returns an iterator over outbound simple-reporting calls, auto-paginating across
// pages. Page is managed by the iterator; leave params.Page unset.
//
// Wraps generated ListOutboundSimpleReportingCalls.
func (r *OutboundResource) SimpleReporting(ctx context.Context, params managerapi.ListOutboundSimpleReportingCallsParams) iter.Seq2[managerapi.ReportingCall, error] {
	return func(yield func(managerapi.ReportingCall, error) bool) {
		var zero managerapi.ReportingCall
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListOutboundSimpleReportingCallsWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// SimpleReportingAll collects every outbound simple-reporting call into a slice (convenience over
// SimpleReporting).
func (r *OutboundResource) SimpleReportingAll(ctx context.Context, params managerapi.ListOutboundSimpleReportingCallsParams) ([]managerapi.ReportingCall, error) {
	var out []managerapi.ReportingCall
	for c, err := range r.SimpleReporting(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Attempts returns an iterator over outbound call attempts, auto-paginating across pages. Page is
// managed by the iterator; leave params.Page unset.
//
// Wraps generated ListOutboundAttempts.
func (r *OutboundResource) Attempts(ctx context.Context, params managerapi.ListOutboundAttemptsParams) iter.Seq2[managerapi.CallAttempt, error] {
	return func(yield func(managerapi.CallAttempt, error) bool) {
		var zero managerapi.CallAttempt
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListOutboundAttemptsWithResponse(ctx, &p)
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
			if len(data.Items) == 0 || page >= pageCount(data.Pagination.Pages) {
				return
			}
			page++
		}
	}
}

// AttemptsAll collects every outbound call attempt into a slice (convenience over Attempts).
func (r *OutboundResource) AttemptsAll(ctx context.Context, params managerapi.ListOutboundAttemptsParams) ([]managerapi.CallAttempt, error) {
	var out []managerapi.CallAttempt
	for a, err := range r.Attempts(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// Leads returns an iterator over outbound leads, auto-paginating across pages. Page is managed by
// the iterator; leave params.Page unset.
//
// Wraps generated ListOutboundLeads.
func (r *OutboundResource) Leads(ctx context.Context, params managerapi.ListOutboundLeadsParams) iter.Seq2[managerapi.Lead, error] {
	return func(yield func(managerapi.Lead, error) bool) {
		var zero managerapi.Lead
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListOutboundLeadsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, l := range data.Items {
				if !yield(l, nil) {
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

// LeadsAll collects every outbound lead into a slice (convenience over Leads).
func (r *OutboundResource) LeadsAll(ctx context.Context, params managerapi.ListOutboundLeadsParams) ([]managerapi.Lead, error) {
	var out []managerapi.Lead
	for l, err := range r.Leads(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// ProcessedLeads returns an iterator over processed outbound leads, auto-paginating across pages.
// Page is managed by the iterator; leave params.Page unset.
//
// Wraps generated ListProcessedOutboundLeads.
func (r *OutboundResource) ProcessedLeads(ctx context.Context, params managerapi.ListProcessedOutboundLeadsParams) iter.Seq2[managerapi.Lead, error] {
	return func(yield func(managerapi.Lead, error) bool) {
		var zero managerapi.Lead
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListProcessedOutboundLeadsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, l := range data.Items {
				if !yield(l, nil) {
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

// ProcessedLeadsAll collects every processed outbound lead into a slice (convenience over
// ProcessedLeads).
func (r *OutboundResource) ProcessedLeadsAll(ctx context.Context, params managerapi.ListProcessedOutboundLeadsParams) ([]managerapi.Lead, error) {
	var out []managerapi.Lead
	for l, err := range r.ProcessedLeads(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// GetList returns an outbound list by id.
//
// Wraps generated GetOutboundList.
func (r *OutboundResource) GetList(ctx context.Context, id string) (*managerapi.LeadListItemResponse, error) {
	resp, err := r.gc.GetOutboundListWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UpdateList updates an outbound list by id.
//
// Wraps generated UpdateOutboundList.
func (r *OutboundResource) UpdateList(ctx context.Context, id string, body managerapi.CreateListRequest) (*managerapi.LeadListItemResponse, error) {
	resp, err := r.gc.UpdateOutboundListWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeleteList deletes an outbound list by id.
//
// Wraps generated DeleteOutboundList.
func (r *OutboundResource) DeleteList(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteOutboundListWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// BulkDeleteLeads bulk-deletes leads from an outbound list.
//
// Wraps generated BulkDeleteOutboundLeads.
func (r *OutboundResource) BulkDeleteLeads(ctx context.Context, id string, body managerapi.LeadBulkDeleteRequest) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.BulkDeleteOutboundLeadsWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// BulkDeleteLeadsAlt bulk-deletes leads from an outbound list via the alternate endpoint.
//
// Wraps generated BulkDeleteOutboundLeadsAlt.
func (r *OutboundResource) BulkDeleteLeadsAlt(ctx context.Context, id string, body managerapi.LeadBulkDeleteRequest) (*managerapi.DefaultV2MessageResponse, error) {
	resp, err := r.gc.BulkDeleteOutboundLeadsAltWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UploadLeadsOptions are the optional fields accepted alongside the CSV file by UploadLeads.
type UploadLeadsOptions struct {
	// Clear clears the list before importing.
	Clear *bool
	// Mapping is a JSON string mapping CSV columns to lead fields.
	Mapping *string
	// Separator is the CSV column separator (default ",").
	Separator *string
}

// UploadLeads imports leads into an outbound list from a CSV file (multipart upload). filename is
// the form file name (e.g. "leads.csv") and file streams the CSV contents.
//
// Wraps generated UploadOutboundLeads.
func (r *OutboundResource) UploadLeads(ctx context.Context, id, filename string, file io.Reader, opts UploadLeadsOptions) (*managerapi.LeadUploadResponse, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if opts.Clear != nil {
		v := "false"
		if *opts.Clear {
			v = "true"
		}
		if err := mw.WriteField("clear", v); err != nil {
			return nil, err
		}
	}
	if opts.Mapping != nil {
		if err := mw.WriteField("mapping", *opts.Mapping); err != nil {
			return nil, err
		}
	}
	if opts.Separator != nil {
		if err := mw.WriteField("separator", *opts.Separator); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	resp, err := r.gc.UploadOutboundLeadsWithBodyWithResponse(ctx, id, mw.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}
