package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// SmsResource is the SMS-records namespace (/api/v2/sms).
//
// List takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
type SmsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns an iterator over SMS records, auto-paginating across pages.
func (r *SmsResource) List(ctx context.Context, params managerapi.ListSmssParams) iter.Seq2[managerapi.Sms, error] {
	return func(yield func(managerapi.Sms, error) bool) {
		var zero managerapi.Sms
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ListSmssWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, s := range data.Items {
				if !yield(s, nil) {
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

// ListAll collects every SMS record into a slice (convenience over List).
func (r *SmsResource) ListAll(ctx context.Context, params managerapi.ListSmssParams) ([]managerapi.Sms, error) {
	var out []managerapi.Sms
	for s, err := range r.List(ctx, params) {
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// Get returns a single SMS record by id.
func (r *SmsResource) Get(ctx context.Context, id string) (*managerapi.SmsItemResponse, error) {
	resp, err := r.gc.GetSmsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Send sends an SMS message.
func (r *SmsResource) Send(ctx context.Context, body managerapi.SmsSendRequest) (*managerapi.SmsItemResponse, error) {
	resp, err := r.gc.SendSmsWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes an SMS record by id, returning the deleted record.
func (r *SmsResource) Delete(ctx context.Context, id string) (*managerapi.SmsItemResponse, error) {
	resp, err := r.gc.DeleteSmsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Report returns an iterator over the SMS reporting records, auto-paginating across pages.
//
// Params takes the generated parameter struct directly; the Page field is managed by the
// auto-paginator, so leave it unset.
func (r *SmsResource) Report(ctx context.Context, params managerapi.ReportSmsParams) iter.Seq2[managerapi.Sms, error] {
	return func(yield func(managerapi.Sms, error) bool) {
		var zero managerapi.Sms
		p := params
		page := 1
		for {
			p.Page = &page
			resp, err := r.gc.ReportSmsWithResponse(ctx, &p)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, s := range data.Items {
				if !yield(s, nil) {
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

// TestInbound simulates an inbound SMS message for testing.
func (r *SmsResource) TestInbound(ctx context.Context, body managerapi.SmsSendRequest) (*managerapi.SmsItemResponse, error) {
	resp, err := r.gc.TestInboundSmsWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}
