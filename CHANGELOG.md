# Changelog

All notable changes to `github.com/babelforce/manager-sdk-go` are documented here. This project
adheres to [Semantic Versioning](https://semver.org/).

## v0.4.0

- Add the `Applications` resource (v2 IVR application management): `List` / `ListAll`
  (auto-paginating iterator), `Create`, `Get`, `Update`, `Delete`, `DeleteMany` (bulk),
  `ListModules`, `Dispatch`, plus a nested `Applications.Actions` sub-resource (local automations:
  `List` / `ListAll` / `Create` / `Get` / `Update` / `Delete`).
- Add the `Settings` resource (v2 global settings): a typed `Get` / `Update` per group across the
  `App`, `Telephony`, `Audit`, `Ui` and `Retention` scopes — e.g.
  `Settings.Telephony.AgentRecording.Get(ctx)` / `.Update(ctx, …)`. Read/write the data payload
  directly; the `{ scope, key }` envelope is handled for you.

## v0.3.0

- Add the `Agents` resource (v2 agent management): `List` / `ListAll` (auto-paginating iterator),
  `Create`, `Get`, `Update`, `Delete`, `UpdateStatus`, plus `Agents.Groups`
  (`List` / `ListAll` / `Create` / `Get` / `Update` / `Delete` / `AddAgent`).
- Add the `Calls.Reporting` resource (v2 call reporting): `List` / `ListAll` (detailed report),
  `Simple` / `SimpleAll`, and `SimpleByType` / `SimpleAllByType` — all auto-paginating iterators.
- Add the `Metrics` resource (v2 metrics): `ListIds`, `Get`, `Describe`, `Push`, `Reset`.

## v0.2.0

- Add the `Tasks` resource (v3 task automations): `Create`, `CreateFromTemplate`, `List` /
  `ListAll` (auto-paginating iterator), `Get`, `Update`, `Interrupt`, plus `Tasks.Schedules`
  (`List` / `Create` / `Get` / `Delete`).

## v0.1.0

Initial release.

- `ManagerClient` (`manager.Connect`) facade with one-shot auth configuration.
- Auth modes: API key (`X-Auth-Access-Id` / `X-Auth-Access-Token`), bearer, and OAuth2 password
  grant with transparent token refresh.
- `Users` resource: auto-paginating `List` (Go 1.23 iterator) / `ListAll`, `Create`, `Enable`,
  `Disable`, `Delete`.
- Typed `*APIError` for non-2xx responses.
