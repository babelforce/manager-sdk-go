# manager-sdk-go

Go SDK for the babelforce **manager APIs** — auth, user & agent management, call reporting,
metrics, and task automations.

One client, configured once, exposes resource namespaces over the API. Authentication, paging, and
error handling are handled for you.

📖 **Docs:** https://babelforce.github.io/manager-sdk/

## Install

```bash
go get github.com/babelforce/manager-sdk-go
```

```go
import (
    manager "github.com/babelforce/manager-sdk-go"
    managerapi "github.com/babelforce/manager-sdk-go/gen/manager" // request & model types
)
```

## Usage

```go
mgr, err := manager.Connect(ctx, manager.Options{
    Auth: manager.ClientCredentials(clientID, clientSecret), // or manager.Password(user, pass)
    // BaseURL defaults to https://services.babelforce.com
})
if err != nil {
    log.Fatal(err)
}

// list users (auto-paginated)
for user, err := range mgr.Users.List(ctx, manager.ListUsersQuery{}) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user.Email)
}

_, err = mgr.Users.Create(ctx, managerapi.CreateManagedUserRequest{Email: "new.user@acme.com"})
```

### Authentication

- `manager.RefreshToken(refreshToken, clientID)` — a refresh token from the Authorization Code +
  PKCE flow (helpers: `manager.GeneratePKCE`, `manager.BuildAuthorizeURL`,
  `manager.AuthorizationCodeGrant`); transparent refresh with rotation. Best for apps acting on
  behalf of a user.
- `manager.ClientCredentials(clientID, clientSecret)` — OAuth2 client_credentials grant with
  transparent refresh, for server-to-server use (credential issuance is in security review — see the
  [Authentication guide](https://babelforce.github.io/manager-sdk/guides/authentication)).
- `manager.Bearer(token)` — a token you already hold.
- `manager.Password(user, pass)` — OAuth2 password grant (legacy) with transparent refresh.

### Errors

Non-2xx responses return a typed `*manager.APIError` (`Status`, `Code`, `Message`, `Body`). Failed
OAuth2 token grants (e.g. invalid client credentials) return the same typed error (`errors.As`
compatible) — branch on `Status`/`Code` instead of the message string.

### Custom host

```go
manager.Connect(ctx, manager.Options{
    BaseURL: "https://acme.babelforce.com",
    Auth:    manager.Bearer(token),
})
```

## License

Apache-2.0
