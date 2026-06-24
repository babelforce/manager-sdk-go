package manager

import (
	"context"
	"iter"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// UsersResource is the user-management namespace (/api/v2/users).
type UsersResource struct {
	gc *managerapi.ClientWithResponses
}

// ListUsersQuery filters a user listing.
type ListUsersQuery struct {
	// Email filters by email address.
	Email *string
}

// List returns an iterator over users, auto-paginating across pages.
//
//	for user, err := range mgr.Users.List(ctx, manager.ListUsersQuery{}) {
//	    if err != nil { return err }
//	    fmt.Println(user.Email)
//	}
func (r *UsersResource) List(ctx context.Context, q ListUsersQuery) iter.Seq2[managerapi.ManagedUser, error] {
	return func(yield func(managerapi.ManagedUser, error) bool) {
		var zero managerapi.ManagedUser

		params := &managerapi.ListUsersParams{}
		if q.Email != nil {
			var email managerapi.ListAgentsGroupIdsParameter
			if err := email.FromListAgentsGroupIdsParameter0(*q.Email); err != nil {
				yield(zero, err)
				return
			}
			params.Email = &email
		}

		for {
			resp, err := r.gc.ListUsersWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			page, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, u := range page.Items {
				if !yield(u, nil) {
					return
				}
			}
			if len(page.Items) == 0 || page.Pagination.Current >= page.Pagination.Pages {
				return
			}
		}
	}
}

// ListAll collects every user into a slice (convenience over List).
func (r *UsersResource) ListAll(ctx context.Context, q ListUsersQuery) ([]managerapi.ManagedUser, error) {
	var users []managerapi.ManagedUser
	for u, err := range r.List(ctx, q) {
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// Create creates a user.
func (r *UsersResource) Create(ctx context.Context, user managerapi.CreateManagedUserRequest) (*managerapi.ManagedUserItemResponse, error) {
	resp, err := r.gc.CreateUserWithResponse(ctx, user)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Enable enables the given users (by email).
func (r *UsersResource) Enable(ctx context.Context, emails []string) error {
	resp, err := r.gc.EnableUsersWithResponse(ctx, emailList(emails))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Disable disables the given users (by email).
func (r *UsersResource) Disable(ctx context.Context, emails []string) error {
	resp, err := r.gc.DisableUsersWithResponse(ctx, emailList(emails))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Delete deletes the given users (by email).
func (r *UsersResource) Delete(ctx context.Context, emails []string) error {
	resp, err := r.gc.DeleteUsersWithResponse(ctx, emailList(emails))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// ResetPasswords triggers a password-reset email for the given users (by email).
func (r *UsersResource) ResetPasswords(ctx context.Context, emails []string) error {
	resp, err := r.gc.ResetPasswordsWithResponse(ctx, emailList(emails))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// ListRoles lists the role names that can be assigned to users.
func (r *UsersResource) ListRoles(ctx context.Context) ([]managerapi.AccountRole, error) {
	resp, err := r.gc.ListAvailableRolesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	out, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// AddRoles grants the given roles to the given users (by email).
func (r *UsersResource) AddRoles(ctx context.Context, emails []string, roles []managerapi.AccountRole) error {
	resp, err := r.gc.AddRolesWithResponse(ctx, roleBinding(emails, roles))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// RemoveRoles revokes the given roles from the given users (by email).
func (r *UsersResource) RemoveRoles(ctx context.Context, emails []string, roles []managerapi.AccountRole) error {
	resp, err := r.gc.RemoveRolesWithResponse(ctx, roleBinding(emails, roles))
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

func emailList(emails []string) managerapi.EmailListRequest {
	out := make([]managerapi.ManagedUserEmail, len(emails))
	for i, e := range emails {
		out[i] = managerapi.ManagedUserEmail(e)
	}
	return managerapi.EmailListRequest{Emails: out}
}

func roleBinding(emails []string, roles []managerapi.AccountRole) managerapi.EmailRoleBinding {
	es := make([]managerapi.ManagedUserEmail, len(emails))
	for i, e := range emails {
		es[i] = managerapi.ManagedUserEmail(e)
	}
	return managerapi.EmailRoleBinding{Emails: &es, Roles: &roles}
}
