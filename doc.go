// Package manager is the babelforce manager SDK for Go.
//
// It provides an intuitive, hand-written client over the babelforce manager APIs — auth,
// user & agent management, call reporting, metrics, and task automations. Configure one
// [ManagerClient] with [Connect], authenticate once, and use its resource namespaces.
//
//	mgr, err := manager.Connect(ctx, manager.Options{
//	    Auth: manager.APIKey(accessID, accessToken),
//	})
//	for user, err := range mgr.Users.List(ctx, manager.ListUsersQuery{}) {
//	    if err != nil { return err }
//	    fmt.Println(user.Email)
//	}
//
// The low-level clients under gen/ are generated from the OpenAPI specs and are an internal
// detail; this package is the public surface.
package manager
