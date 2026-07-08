package manager

import (
	"context"

	userapi "github.com/babelforce/manager-sdk-go/gen/user"
)

// MeResource is the authenticated-principal namespace (/api/v2/user): the current user, the
// accounts they can access, and self-service password reset.
type MeResource struct {
	uc *userapi.ClientWithResponses
}

// Get returns the current user.
func (r *MeResource) Get(ctx context.Context) (*userapi.UserItemResponse, error) {
	resp, err := r.uc.GetUserWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Customer returns the current user together with their account (customer) information.
func (r *MeResource) Customer(ctx context.Context) (*userapi.UserCustomerItemResponse, error) {
	resp, err := r.uc.GetUserCustomerWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Accounts lists the accounts the current user can access.
func (r *MeResource) Accounts(ctx context.Context) ([]userapi.Account, error) {
	resp, err := r.uc.ListAccountsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// ResetPassword requests a password-reset email for the current user.
func (r *MeResource) ResetPassword(ctx context.Context) error {
	resp, err := r.uc.ResetPasswordWithResponse(ctx)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
