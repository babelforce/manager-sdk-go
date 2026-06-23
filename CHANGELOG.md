# Changelog

All notable changes to `github.com/babelforce/manager-sdk-go` are documented here. This project
adheres to [Semantic Versioning](https://semver.org/).

## v0.1.0

Initial release.

- `ManagerClient` (`manager.Connect`) facade with one-shot auth configuration.
- Auth modes: API key (`X-Auth-Access-Id` / `X-Auth-Access-Token`), bearer, and OAuth2 password
  grant with transparent token refresh.
- `Users` resource: auto-paginating `List` (Go 1.23 iterator) / `ListAll`, `Create`, `Enable`,
  `Disable`, `Delete`.
- Typed `*APIError` for non-2xx responses.
