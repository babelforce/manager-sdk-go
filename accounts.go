package manager

import (
	"context"
	"net/http"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// AccountsResource is the sub-accounts namespace of a partner account (/api/v2/accounts). A partner
// account holds one direct sub-account per organisation it manages, each named by the partner's own
// orgKey.
//
// Only a service token of the partner account may call it (the [ClientCredentials] auth mode with an
// OAuth application holding account:manage); a person's token answers 403 insufficient_scope. Act
// inside a sub-account with [ManagerClient.ForSubAccount].
type AccountsResource struct {
	gc *managerapi.ClientWithResponses
}

// ListSubAccountsQuery filters [AccountsResource.List].
type ListSubAccountsQuery struct {
	// OrgKey selects the sub-account with this orgKey only.
	OrgKey *string
}

// SubAccountCreated is the answer of [AccountsResource.Create].
type SubAccountCreated struct {
	// UUID is the sub-account's account id.
	UUID string
	// Created is false when a sub-account with this orgKey and this name already existed and
	// nothing was created.
	Created bool
}

// List returns the direct sub-accounts of your partner account, oldest first. The list is not
// paged; q may be nil.
func (r *AccountsResource) List(ctx context.Context, q *ListSubAccountsQuery) ([]managerapi.SubAccount, error) {
	params := &managerapi.ListSubAccountsParams{}
	if q != nil {
		params.OrgKey = q.OrgKey
	}
	resp, err := r.gc.ListSubAccountsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// Get returns the sub-account with this orgKey, or nil when there is none.
func (r *AccountsResource) Get(ctx context.Context, orgKey string) (*managerapi.SubAccount, error) {
	items, err := r.List(ctx, &ListSubAccountsQuery{OrgKey: &orgKey})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

// Create creates a direct sub-account. It is safe to retry: the same orgKey with the same name
// answers the existing sub-account with Created false. A refusal is an [*APIError] whose Body names
// it: 409 org_key_taken, org_key_is_account_id or name_taken; 400 invalid_request or invalid_name;
// 403 insufficient_scope or not_a_partner_account.
func (r *AccountsResource) Create(ctx context.Context, body managerapi.CreateSubAccountRequest) (*SubAccountCreated, error) {
	resp, err := r.gc.CreateSubAccountWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	payload := resp.JSON201
	if payload == nil {
		payload = resp.JSON200
	}
	out, err := result(payload, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return &SubAccountCreated{UUID: out.Uuid, Created: resp.HTTPResponse.StatusCode == http.StatusCreated}, nil
}
