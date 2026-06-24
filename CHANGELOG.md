# Changelog

All notable changes to `github.com/babelforce/manager-sdk-go` are documented here. This project
adheres to [Semantic Versioning](https://semver.org/).

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
