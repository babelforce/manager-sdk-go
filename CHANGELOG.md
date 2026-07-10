# Changelog

All notable changes to `github.com/babelforce/manager-sdk-go` are documented here. This project
adheres to [Semantic Versioning](https://semver.org/).

## v0.45.0

- **Typed OAuth grant errors (behavioral change for message-string matchers).** A failed OAuth2
  token grant — `PasswordGrant`, `ClientCredentialsGrant`, `AuthorizationCodeGrant`,
  `RefreshTokenGrant`, and the transparent token fetch behind the grant-based `Auth` modes — now
  returns the same typed `*APIError` (`Status`, `Code`, `Message`, `Body`; `errors.As` compatible)
  as every other non-2xx response, instead of an opaque `fmt.Errorf` with the status embedded in
  the message. Callers that matched the old `"manager: <grant> grant failed (status N)"` strings
  should branch on `Status` / `Code` instead.

## v0.44.1

- No Go changes — lockstep version bump. Maintenance release: a cross-SDK behavioral conformance
  harness now exercises the real TS/Go/Rust facades against one scripted mock server in CI (retry
  and auth/token scenario families, advisory job). The published module is unchanged.

## v0.44.0

- `PasswordWithClientID(user, pass, clientID)`: password grant against a custom OAuth2 application
  instead of the default `"manager"` client id.
- `RefreshTokenWithSecret(refreshToken, clientID, clientSecret)`: refresh-token auth for
  confidential clients — the client secret is sent with every rotating exchange. Existing
  `Password`/`RefreshToken` constructors are unchanged. Both mirror the TypeScript SDK's optional
  `clientId`/`clientSecret`.
- Spec-inventory drift guard: a monorepo-only test pins the `go:generate` directives to
  `specs/sources.json` (skips outside the monorepo; no consumer-facing change).

## v0.43.0

- Token manager: OAuth grants no longer run while holding the client-wide mutex. Refreshes are
  single-flight, waiting callers honor their own context deadlines, and a hung token endpoint can
  no longer stall every request on the client.
- `result()`: a 2xx response whose status differs from the spec-declared code (e.g. a live 200
  where the spec says 201) now decodes the payload instead of being misreported as an `*APIError`
  carrying a 2xx status.
- Auto-pagination hardening: all list iterators (33 loops across 23 resources) advance by the
  locally requested page instead of the server-echoed cursor, so a misreporting server can no
  longer cause an infinite duplicate-yielding loop.
- `Users.List` issues a single request: the users endpoint exposes no page parameter; the response
  is yielded as served (the `Email` filter is unchanged).

## v0.42.3

- No Go changes — lockstep version bump. Maintenance release: crates.io publishing now runs
  automatically on tag push (like npm and the Go module), plus internal workflow-docs housekeeping.
  The published module is unchanged.

## v0.42.2

- No Go changes — lockstep version bump. This release fixes a Rust-only bug: path parameters now
  normalize a UUID to the API's unhyphenated form, so a `list` result round-trips cleanly into
  `get`/`update`/`delete` instead of 404ing.

## v0.42.1

- No Go changes — lockstep version bump. This release is the Rust SDK's first crates.io publish
  plus a pre-publish hardening pass (secret-redacting `Debug`, a de-panicked authorize-URL builder,
  and hardened auto-pagination).

## v0.42.0

- No Go changes — lockstep version bump. This release ships Rust-only facade additions (raw
  application create/clone where the typed decode fails on populated success payloads, and an
  in-memory lead upload).

## v0.41.0

- No Go changes — lockstep version bump. This release ships Rust-only facade additions
  (single-page listing + CDR report filter, raw-JSON reads where the generated models are stricter
  than the live API, paged SMS reporting, and a raw call hangup).

## v0.40.0

- Added first-class **Authorization Code + PKCE** (RFC 7636) support: a new `manager.RefreshToken`
  auth mode (transparent refresh with refresh-token rotation) plus the helpers `manager.GeneratePKCE`,
  `manager.BuildAuthorizeURL`, `manager.AuthorizationCodeGrant`, and `manager.RefreshTokenGrant`.
  `TokenResponse` gained a `RefreshToken` field.
- Consolidated the docs: the OAuth guide is merged into a single **Authentication** guide (overview +
  one section per flow + when-to-use). Client credentials is documented with a security-review/limited-
  availability caveat.

## v0.39.0

- Refreshed the vendored manager API spec (richer endpoint descriptions/examples) and regenerated
  the client. No operations added or removed — coverage stays at 397/397.
- **BREAKING (type rename):** `Settings.Telephony.AgentOutbound.Update` now takes
  `SettingsTelephonyAgentOutboundRequestData` (was `SettingsTelephonyAgentInboundRequestData`); the
  fields are identical.
- Schema correction: the routing `number` field is now typed as a phone-number reference (it
  previously aliased the application schema).

## v0.38.0

- **BREAKING:** Removed `APIKey` auth (the `X-Auth-Access-Id` / `X-Auth-Access-Token` header pair).
  The babelforce API no longer accepts those headers.
- Added `ClientCredentials(clientID, clientSecret)` — an OAuth2 `client_credentials` grant (lazy
  fetch + transparent refresh) as the server-to-server replacement for `APIKey`.

## v0.37.0

- Complete manager API parity (356/356 operations; 397/397 across all specs). Final operations:
  `Users.Me`/`Users.GetByEmail`, `Prompts.Uses`, `Metrics.Definitions`, `Babeldesk.WidgetSettings`,
  and `Applications.AllLocalAutomations`. Re-enabled the `full` coverage gate.

## v0.36.0

- Add a new `System` resource for system/reference endpoints: `Echo`, `Ping`, `ApiStatus`,
  `ServerTime`, `Timezones`, `PushToken`, `Tags`, `TagsByCategory`, `ExportTemplates` — 9 new
  operations.

## v0.35.0

- Extend `Integrations` with the action catalog and execution: `ListActions`, `ListActionParams`,
  `ExecuteAction`, and `DispatchActionGet` — 4 new operations.

## v0.34.0

- Extend `Queues` (`BulkUpdate`, `GlobalSelections`, `Selections.SetPriority`), `Outbound`
  (`GetLead`, `ListLeads`), `Phonebook` (`BulkDelete`), and `Numbers` (`Update`) — 7 new
  operations.

## v0.33.0

- Extend `Calls` (`Cancel`, `ListQueued`, `QueueCallback`, and `Reporting.InboundSimpleAll`) and
  `Logs` (`EnableLive`, `DisableLive`, `Write`) — 7 new operations.

## v0.32.0

- Extend `Settings` with generic scope/key operations (`ListAll`, `ListInScope`, `Clear`,
  `ClearInScope`, `ClearAll`) and `Applications` (`Clone`, `BulkUpdate`, `ListActions`,
  `ListErrors`) — 9 new operations.

## v0.31.0

- Extend `Conversations` (`AddEvent`, `Open`, `Close`, `FirstEvent`, `LatestEvent`, `AllEvents`)
  and `Sms` (`Send`, `Delete`, `Report`, `TestInbound`) — 10 new operations.

## v0.30.0

- Extend `BusinessHours` (range `AddRanges`/`ListRanges`/`GetRange`/`RemoveRange`,
  `BulkUpdate`/`BulkDelete`) and `Calendars` (individual-date `GetDate`/`UpdateDate`/`RemoveDate`,
  `TestDate`, `BulkUpdate`/`BulkDelete`) — 12 new operations.

## v0.29.0

- Extend `Agents` with bulk actions (`BulkAction`), CSV `Export`/`Import`/`ValidateImport` +
  `GetImportJob`, per-agent and global activity `Logs`/`AllLogs`, `Push`, `UpdatePassword`, and
  group membership (`Groups.RemoveAgent`, `Groups.ListAgents`, `Groups.BulkDelete`) — 12 new
  operations.

## v0.28.0

- Extend `Agents` with named presence CRUD (`Presences`, `GetPresence`, `CreatePresence`,
  `UpdatePresence`, `DeletePresence`), status reads (`GetStatus`, `AvailableStatuses`), and
  lifecycle controls (`Enable`, `Disable`, `HangupCall`) — 10 new operations.

## v0.27.0

- Extend `Triggers` (expression/operator catalogs, `Conditions`/`SetConditions`, `Uses`,
  `BulkAction`), `Automations` (`Clone`, `Dispatch`, `BulkUpdate`/`BulkDelete`), and `Queues`
  (`ListTriggers`) — 11 new operations.

## v0.26.0

- Extend `Outbound` with outbound-list management, lead/attempt browsing, `CreateAgentCall`,
  bulk lead deletion, and CSV lead uploads (11 new operations).

## v0.25.0

- Extend `Campaigns` with realtime `Status`/`Statistics`, `Hopper`, `Leads`/`ProcessedLeads`,
  `Attempts`, lead-list assignment, and `LogoutAllAgents` (11 new operations).

## v0.24.0

- Extend `Integrations` with the provider catalog, OAuth token management, `Clone`,
  `Authorize`, `Integrate`, type actions, the API proxy, templates, and bulk update/delete
  (16 new operations).

## v0.23.0

- Add the `Recordings` resource (v2 manager API): `List`/`ListAll`, `Start`, `Get`, `Update`,
  `Delete`, `BulkAction`, plus recording flags (`GetFlag`, `Flag`, `Unflag`, `ToggleFlag`).

## v0.22.0

- Add the `Dialer` resource (v2 manager API): `Info`, `Flush`, `SimpleReporting` /
  `SimpleReportingAll`, plus a nested `Dialer.Behaviours` sub-resource (CRUD).

## v0.21.0

- Add the `Files` resource (v2 manager API): `List`/`ListAll` (auto-paginated), `ListByType`,
  `Backups`, `Recordings`, `Prompts`, `Get`, `Delete`, `Download`, `BulkDelete`, `BulkDownload`,
  `BulkDownloadPost`.

## v0.20.0

- Add the `Dashboards` resource (v2 manager API): `List` / `ListAll` (auto-paginated), `Create`,
  `Get`, `Update`, `Delete`, plus dashboard-user access (`ListUsers`, `AddUser`, `RemoveUser`).

## v0.19.0

- Add the `Auth` resource — the **OAuth 2.0** endpoints (`/oauth/authorize`, `/oauth/token`,
  `/oauth/revoke`): `mgr.Auth.Token`, `mgr.Auth.Revoke`, `mgr.Auth.Authorize`. These authenticate
  via client credentials in the request, independent of the SDK's api-key/bearer auth.

## v0.18.0

- **Refreshed the manager OpenAPI spec** (a large upstream expansion: 187 → 356 operations) and
  **vendored the new `auth` (OAuth 2.0) spec**. The low-level generated clients now cover the new
  surface; ergonomic facade methods for the new operations land in follow-up releases.
- **Breaking — removed operations no longer in the spec:** `Conversations.GetEvent`,
  `Calls.Reporting.SimpleByType` / `SimpleAllByType`, `Metrics.Push`, `Metrics.Reset`.
- **Breaking — changed signatures:** `Calendars.AddDate(ctx, id, body)` now takes a
  `CalendarDateBody` and returns `*CalendarDateItemResponse`; `Calendars.GetDates` returns
  `[]CalendarDate`.
- Coverage is tracked against the new 397-operation scope; the SDK currently wraps ~56% while the
  newly-added operations are wrapped over the coming releases.

## v0.17.0

- Extend `Tasks` with `Tasks.Metrics` (task journal, agent journal, agent interaction durations),
  `Usage` / `UsageTypes`, `Logs` (customer logs), `AgentAction`, `SetAgentLock`, `ChangeState`, and
  `TestAction`.
- **The SDK now wraps 100% of every babelforce manager spec** (manager v2, user, task-automation,
  task-schedule); the only excluded operation is the Zendesk webhook receiver, which is not a client call.

## v0.16.0

- Extend `Tasks` (v3 task automation) with `Tasks.Scripts` (script CRUD by type), `Tasks.Secrets`
  (prefixes, keys, create/patch/delete), and `Tasks.SelectionConfig` (account task-selection
  configuration: read/create/update/delete).

## v0.15.0

- Add the `Events` resource (`List`, `CreateCustom`, `DeleteCustom`), the `Logs` resource (`Audit` /
  `AuditAll` auto-paginated, `Live`), and the `Expressions` resource (`List`, `Evaluate`).
- Extend `Integrations` with `DispatchAction` and `ActionVariables`.
- **The Go SDK now wraps 100% of the manager (v2) API.**

## v0.14.0

- Add the `Prompts` resource (v2 audio prompts): `List` / `ListAll`, `Get`, `Upload`, `Update`,
  `Delete`.
- Add the `Babeldesk` resource (v2 babeldesk dashboards): `List` / `ListAll`, `Create`, `Get`,
  `Update`, `Delete`, plus a nested `Babeldesk.Widgets` sub-resource (widget CRUD).

## v0.13.0

- Add the `Conversations` resource (v2 conversations): `List` / `ListAll`, `Create`, `Get`,
  `Update`, `Delete`, plus `Events`, `GetEvent`, `GetSession`, `UpdateSession`.
- Add the `Sessions` resource (v2 sessions): `Create`, `Get`, `UpdateVariables`.

## v0.12.0

- Add the `BusinessHours` resource (v2 business hours): `List` / `ListAll`, `Create`, `Get`,
  `Update`, `Delete`.
- Add the `Calendars` resource (v2 calendars): `List` / `ListAll`, `Create`, `Get`, `Update`,
  `Delete`, plus `GetDates` / `AddDate`.

## v0.11.0

- Add the `Outbound` resource (v2 dialer lists & leads): `Lists`, `CreateList`, `ClearList`,
  `AddLead`, `UpdateLead`, `DeleteLead`.
- Add the `Phonebook` resource (v2 phonebook): `List` / `ListAll`, `Create`, `Get`, `Update`,
  `Delete`, plus bulk `Download` / `Upload`.
- Add the `Campaigns` resource (v2 outbound campaigns): `List`, `Create`, `Get`, `Update`, `Delete`.

## v0.10.0

- Add the `Integrations` resource (v2 integrations): `List` / `ListAll`, `Create`, `Get`, `Update`,
  `Delete`, `Available`, `AddAssociation` / `RemoveAssociation`, `ProviderLogo`, and
  `ProviderSessionVariables`.

## v0.9.0

- Add the `Routing` resource (v2 routing rules): `List` / `ListAll`, `Create`, `Get`, `Update`,
  `Delete`.
- Add the `Triggers` resource (v2 workflow triggers): CRUD plus `Clone` and `Test`.
- Add the `Automations` resource (v2 global automations / event triggers): `List` / `ListAll`,
  `Create`, `Get`, `Update`, `Delete`.

## v0.8.0

- Add the `Queues` resource (v2 queues): `List` / `ListAll`, `Create`, `Get`, `Update`, `Delete`,
  plus a nested `Queues.Selections` sub-resource — selection CRUD, agent/group/tag membership
  (`AddAgent` / `RemoveAgent` / `AddGroup` / `RemoveGroup` / `AddTag` / `RemoveTag`), and
  `SelectAgents` to resolve a queue's selected agents.

## v0.7.0

- Extend the `Calls` resource with call control: `Get`, `Hangup`, `CreateTestCall`, and
  `SetSessionVariables`.
- Add the `Sms` resource (v2 SMS records): `List` / `ListAll` (auto-paginating) and `Get`.
- Add the `Numbers` resource (v2 service numbers): `List` / `ListAll`, `Get`, and `AddTags`.
- Add the `Conferences` resource (v2 conferences): `List` / `ListAll` and `Get`.

## v0.6.0

- Extend the `Users` resource: `ListRoles` (assignable role names), `AddRoles(emails, roles)` /
  `RemoveRoles(emails, roles)`, and `ResetPasswords(emails)`.
- Add the `Me` resource (the authenticated principal, v2 user API): `Get` (current user),
  `Customer` (current user + account info), `Accounts` (accounts the principal can access), and
  `ResetPassword`.

## v0.5.0

- **Breaking:** removed `Options.Environment` and the `Environment`/`Production` API. The client now
  targets `https://services.babelforce.com` by default — set `Options.BaseURL` to point at another
  host. The default is exported as `DefaultBaseURL`.
- Add automatic retries via the new `Options.Retry *RetryPolicy`, on by default with conservative
  settings (set `&RetryPolicy{MaxRetries: 0}` to disable). Transient failures — network errors and
  `429`/`502`/`503`/`504` — are retried with exponential backoff and jitter, honouring `Retry-After`;
  non-idempotent requests are only retried on `429`.
- **Breaking:** `Applications.Dispatch` now takes the request body as an optional
  `*managerapi.LocalAutomationDispatch` (previously a value). Pass `nil` to send no request body —
  parity with the TypeScript SDK, which previously differed by always sending an empty `{}`.
- Add `manager.ApplicationViewOf(app)` (and the `ApplicationView` type) to read the fields every IVR
  application variant shares (`Id`, `Name`, `Module`, `Enabled`, `DateCreated`, `LastUpdated`,
  `Tags`). `managerapi.Application` is a `oneOf` union with no directly addressable fields; for
  module-specific fields use `app.As<Module>Application()` or `app.ValueByDiscriminator()`.

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
