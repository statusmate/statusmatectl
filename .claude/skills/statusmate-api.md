# Statusmate API Reference

Use this skill whenever implementing or debugging API calls in this codebase. The live schema is at `https://devstatusmate.ru/api/schema/`.

## Base URL & Auth

- **Dev server:** `https://devstatusmate.ru`
- **Prod server:** `https://statusmate.top` (default `--server` flag value)
- **Auth header:** `Authorization: Token <key>` (token obtained from `/api/auth/signin/`)
- All endpoints except signin/signup require token auth.

---

## Authentication

### POST /api/auth/signin/
Request:
```json
{ "username": "email@example.com", "password": "string" }
```
Response 200 — token auth:
```json
{ "key": "<40-char token>", "created": "<datetime>", "user": 123 }
```
Response 201 — 2FA required:
```json
{ "token": "<2fa-session-token>" }
```

### POST /api/auth/two_factor_verify/
```json
{ "code": "123456", "token": "<2fa-session-token>" }
```
Returns same `Token` object on success.

### GET /api/auth/me/
Returns current user profile.

---

## Status Pages

### GET /api/pages/
Query params: `ordering`, `page`, `size`, `search`, `status_page` (int)
Returns: `PaginatedStatusPageList { count, results: [StatusPage] }`

### POST /api/pages/
### GET/PUT/PATCH/DELETE /api/pages/{slug}/

**StatusPage key fields:**
| Field | Type | Notes |
|---|---|---|
| `id` | int | read-only |
| `uuid` | uuid | read-only |
| `slug` | string | `^[-a-zA-Z0-9_]+$`, max 150 |
| `name` | string | max 255 |
| `description` | string\|null | |
| `impact` | ImpactEnum | current page-level impact |
| `timezone` | TimezoneEnum | |
| `custom_domain` | string\|null | read-only |
| `absolute_url` | string | read-only |

### GET /api/pages/{slug}/overview/
Returns page overview with components and active incidents.

---

## Incidents

### GET /api/incident/
Query params:
- `status` (array): `incident_investigating`, `incident_identified`, `incident_monitoring`, `incident_resolved`
- `status_page` (int)
- `impact` (array): see ImpactEnum
- `affect_uptime` (bool)
- `show_on_top` (bool)
- `start_at_after`, `start_at_before`, `end_at_after`, `end_at_before` (datetime)
- `ordering`, `page`, `size`, `search`

Returns: `PaginatedIncidentList { count, results: [Incident] }`

### POST /api/incident/
Uses `IncidentCreate` schema. Required: `title`, `status`, `status_page`, `components[]`, `description` (writeOnly), `start_at`.

**IncidentCreate request body:**
```json
{
  "title": "string",
  "description": "string (write-only, becomes first update)",
  "status": "incident_investigating",
  "status_page": 1,
  "start_at": "2024-01-01T00:00:00Z",
  "end_at": null,
  "notify": true,
  "show_on_top": true,
  "affect_uptime": true,
  "show_on_page": true,
  "postmortem": false,
  "private_note": null,
  "worst_impact": null,
  "components": [
    { "component": 42, "impact": "partial_outage" }
  ]
}
```
Returns 201 + `Incident` object.

### GET /api/incident/{uuid}/
### PUT /api/incident/{uuid}/
### PATCH /api/incident/{uuid}/
Uses `PatchedIncident` — all fields optional. `status` is **read-only** on the Incident object; to change status, post an incident update.
### DELETE /api/incident/{uuid}/ → 204

**Incident response fields:**
| Field | Type | Notes |
|---|---|---|
| `id` | int | read-only |
| `uuid` | uuid | read-only |
| `title` | string | |
| `status` | Status3a3Enum | read-only (set via updates) |
| `worst_impact` | WorstImpactEnum | read-only |
| `description` | string | not in response, write-only at creation |
| `updates` | string | read-only summary |
| `last_update_at` | string | read-only |
| `absolute_url` | string | read-only |
| `start_at` | datetime | |
| `end_at` | datetime\|null | |
| `notify` | bool | |
| `show_on_top` | bool | |
| `affect_uptime` | bool | |
| `show_on_page` | bool | |
| `private_note` | string\|null | |
| `status_page` | int | read-only |
| `created_at` / `updated_at` | datetime | read-only |

---

## Components

### GET /api/component/
Query params: `status_page` (int), `ordering`, `page`, `size`, `search`

### POST /api/component/
### GET/PUT/PATCH/DELETE /api/component/{uuid}/

**Component key fields:**
| Field | Type | Notes |
|---|---|---|
| `id` | int | read-only |
| `uuid` | uuid | read-only |
| `name` | string | max 255, required |
| `description` | string\|null | |
| `impact` | ImpactEnum | current component status |
| `status_page` | int | required |
| `parent` | int\|null | parent component id |
| `index` | int | display order (0–32767) |
| `histogram` | bool | show uptime histogram |
| `collapse` | bool | collapsed group |
| `enabled` | bool | |
| `private` | bool | |
| `uptime` | decimal | computed |

### POST /api/component/batch_update/
Bulk update multiple components at once.

---

## Maintenance

### GET /api/maintenance/
Query params: `status_page` (int), `status` (array), `ordering`, `page`, `size`

### POST /api/maintenance/
Uses `MaintenanceCreate`. Required: `title`, `status_page`, `components[]`, `description` (write-only), `start_at`.

**MaintenanceCreate request body:**
```json
{
  "title": "string",
  "description": "string (write-only)",
  "status_page": 1,
  "start_at": "2024-01-01T00:00:00Z",
  "end_at": null,
  "notify": true,
  "notify_before": false,
  "notify_before_minutes": null,
  "auto_start": false,
  "auto_end": false,
  "notify_auto_start": false,
  "notify_auto_end": false,
  "affect_uptime": true,
  "show_on_page": true,
  "components": [
    { "component": 42, "impact": "under_maintenance" }
  ]
}
```

**MaintenanceStatus (StatusAd8Enum):**
- `maintenance_not_started` — Planned
- `maintenance_in_progress` — In Progress
- `maintenance_completed` — Completed

### GET/PUT/PATCH/DELETE /api/maintenance/{uuid}/

---

## Enums

### ImpactEnum (component/incident impact)
| Value | Label |
|---|---|
| `operational` | Operational |
| `under_maintenance` | Under Maintenance |
| `degraded_performance` | Degraded Performance |
| `partial_outage` | Partial Outage |
| `major_outage` | Major Outage |

**Shorthands** (used in CLI `--components` flag, parsed by `pkg/api/impact.go`):
`o`/`op` → operational · `u`/`um` → under_maintenance · `d`/`dp` → degraded_performance · `p`/`po` → partial_outage · `m`/`mo` → major_outage

### IncidentStatus (Status3a3Enum)
`incident_investigating` · `incident_identified` · `incident_monitoring` · `incident_resolved`

Active statuses (used for default list filter): first three.

---

## Pagination pattern

All list endpoints return:
```json
{ "count": 100, "results": [...] }
```
Paginated with `page` (1-based) and `size`. To fetch all: use `size=1000, page=1` — see `NewAllPaginatedRequest()` in `pkg/api/paginated.go`.

---

## API Tokens

### GET/POST /api/api_token/
### GET/PUT/PATCH/DELETE /api/api_token/{uuid}/
Manage per-page API tokens. Token object: `{ key, created, user, status_page }`.

---

## Other endpoints (available, not yet wired in CLI)

| Endpoint | Purpose |
|---|---|
| `GET/POST /api/release/` | Release notes |
| `GET/POST /api/release_page/` | Release pages |
| `GET/POST /api/subscriber/` | Page subscribers |
| `GET/POST /api/incoming_webhook/` | Grafana/Prometheus/PagerDuty webhooks |
| `GET /api/incoming_webhook/{uuid}/logs/` | Webhook logs |
| `GET /api/logs/` | Audit log |
| `GET/POST /api/tag/` | Component tags |
| `GET/POST /api/teams/` | Team management |
| `GET/POST /api/team_invite/` | Team invites |
| `GET /api/balance/` | Account balance |
| `GET /api/invoice/` | Invoices |
