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
import manager "github.com/babelforce/manager-sdk-go"
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

_, err = mgr.Users.Create(ctx, manager.CreateManagedUserRequest{Email: "new.user@acme.com"})
```

### Authentication

- `manager.ClientCredentials(clientID, clientSecret)` — OAuth2 client_credentials grant with
  transparent refresh (recommended for server-to-server).
- `manager.Bearer(token)` — a token you already hold.
- `manager.Password(user, pass)` — OAuth2 password grant with transparent refresh.

### Errors

Non-2xx responses return a typed `*manager.APIError` (`Status`, `Code`, `Message`, `Body`).

### Custom host

```go
manager.Connect(ctx, manager.Options{
    BaseURL: "https://acme.babelforce.com",
    Auth:    manager.Bearer(token),
})
```

## License

Apache-2.0
