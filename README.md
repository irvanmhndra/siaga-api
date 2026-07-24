# siaga-api

Backend for **Siaga** — a PagerDuty-style on-call & incident management SaaS.
Go + Echo v5 + Postgres, layered (handler → service → repository), opaque-token
sessions, multi-tenant by organization. Companion: `siaga-web` (React dashboard).

## Run

```bash
cp .env.example .env       # adjust POSTGRES_* if needed
createdb siaga_db
make run                   # migrations run automatically on boot
```

Server listens on `:8089`. Response envelope: `{ success, message?, data?, error_code? }`.
Register creates an organization + owner; every domain row is scoped by `organization_id`.

## What it does

- **Services** — monitored systems, each with a unique `integration_key`.
- **Events API** — `POST /api/v1/events/enqueue` (authenticated by `routing_key`
  in the body). Actions `trigger` / `acknowledge` / `resolve` with a `dedup_key`:
  at most one open incident per (service, dedup_key).
- **Incidents** — lifecycle triggered → acknowledged → resolved, with a timeline.
- **Escalation policies** — ordered rules → targets (a user or an on-call
  schedule), repeated up to `repeat_count` times.
- **On-call schedules** — rotation layers + overrides; the resolver computes who
  is on-call at any instant (overrides win, else the highest-position layer).
- **Escalation engine** — a background worker advances unacknowledged incidents
  when `next_escalation_at` passes, and auto-resolves past a service's window.
- **Notifications** — in-app is always recorded; email (SMTP), SMS + voice
  (Twilio) and Slack (webhook) are **gated by config** and disabled when unset.

### Send an event

```bash
curl -X POST http://localhost:8089/api/v1/events/enqueue \
  -H 'Content-Type: application/json' \
  -d '{"routing_key":"<integration_key>","event_action":"trigger",
       "dedup_key":"disk-full-01",
       "payload":{"summary":"Disk 95% full","severity":"critical"}}'
```

## Endpoints (`/api/v1`)

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/auth/register` `/auth/login` `/auth/refresh` | – | org + user |
| POST `/auth/logout` · GET `/me` | | ✓ | |
| POST | `/events/enqueue` | routing_key | the Events API |
| GET/POST | `/users` · PUT `/profile` | ✓ | roster / invite member / prefs |
| CRUD | `/teams` (+ `/:id/members`) | ✓ | |
| CRUD | `/schedules` (+ `/:id/on-call`, `/:id/overrides`) | ✓ | |
| CRUD | `/escalation-policies` | ✓ | |
| CRUD | `/services` | ✓ | |
| GET | `/incidents` `/incidents/:id` | ✓ | |
| POST | `/incidents/:id/acknowledge` `/resolve` | ✓ | |
| GET | `/notifications` · POST `/:id/read` `/read-all` | ✓ | in-app inbox |

Auth routes use `Authorization: Bearer <access_token>`.

## Tests

```bash
go test ./...
```

Unit tests (on-call resolver, urgency mapping) run anywhere. Integration tests
(`tests/integration`) exercise the Events API + escalation against Postgres via
`TEST_DATABASE_DSN` (defaults to a local `siaga_test` DB); they skip cleanly when
no database is reachable. CI provisions a Postgres service and sets the DSN.

## Single-origin deployment

Set `WEB_DIST` to the built `siaga-web/dist` and the API serves the SPA on the
same origin (client-side routes fall back to `index.html`); leave it empty to run
API-only with the web dev server.

Production mirrors the standard setup: multi-stage `Dockerfile` → GHCR via GitHub
Actions on push to `main` → `docker compose` on the VPS (Postgres + api, web dist
mounted at `/app/web`) behind nginx. See `docker-compose.yml` and
`.github/workflows/deploy.yml`. Required Actions secrets: `VPS_HOST`, `VPS_USER`,
`VPS_SSH_KEY`. Live email/SMS/voice/Slack need the `SMTP_*` / `TWILIO_*` /
`SLACK_DEFAULT_WEBHOOK_URL` env vars in the server `.env`.
