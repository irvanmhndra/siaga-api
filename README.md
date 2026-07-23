# undangin-api

Backend for **Undangin** — a digital wedding-invitation SaaS where each couple
customizes their own event (template + fill). Go + Echo v5 + Postgres, layered
(handler → service → repository), token sessions. Companion: `undangin-web`
(React: couple dashboard + public shareable invitation page).

## Run

```bash
cp .env.example .env      # adjust POSTGRES_* if needed
createdb undangin_db
make run                  # migrations run automatically on boot
```

Server listens on `:8088`. Envelope: `{ success, message?, data?, error_code? }`.

## Model

- **invitations** — one wedding. Queryable columns (`slug` unique, `template`,
  `title`, `status` draft/published) + a flexible **`content` JSONB** document
  the template-fill editor owns (couple, events, gift, gallery, story, section
  toggles, music, …).
- **guests** — invited guests with a short `code` for personalized links
  (`/i/<slug>?to=<code>` shows "Kepada Yth: <name>").
- **wishes** — public guest submissions: RSVP (`attending` yes/no/maybe +
  `guest_count`) plus a message (ucapan).

## Endpoints (`/api/v1`)

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/auth/register` `/auth/login` `/auth/refresh` | – | couple accounts |
| POST | `/auth/logout` · GET `/me` | ✓ | |
| POST/GET | `/invitations` | ✓ | create (`title`,`template`) / list mine |
| GET/PUT/DELETE | `/invitations/:id` | ✓ | edit full doc (`slug`,`template`,`title`,`status`,`content`) |
| POST/GET | `/invitations/:id/guests` | ✓ | add / list guests (each gets a `code`) |
| DELETE | `/invitations/:id/guests/:guestId` | ✓ | |
| GET | `/invitations/:id/wishes` | ✓ | responses + attendance stats |
| GET | `/public/:slug?to=<code>` | – | published invitation + resolved guest name |
| GET | `/public/:slug/wishes` | – | public ucapan wall (published only) |
| POST | `/public/:slug/wishes` | – | guest RSVP + ucapan |

Only `published` invitations are reachable publicly. Auth routes use
`Authorization: Bearer <access_token>`.

## Single-origin SPA + WhatsApp previews

Set `WEB_DIST` to the built `undangin-web/dist` and the API serves the SPA on
the same origin. On `GET /i/:slug` it injects Open Graph / Twitter meta
(couple names as title, cover photo as `og:image`) so shared links get a rich
preview in WhatsApp and social apps — the crawler reads meta without running JS.
Leave `WEB_DIST` empty to run API-only (web on its own dev server).

## Later (needs Cloudflare R2)

Photo uploads (cover/prewedding gallery). v1 stores image **URLs** inside
`content`; wire presigned direct-to-R2 upload once the bucket/CDN/keys exist.
