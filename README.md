# IsAIorNot API

Backend Go for IsAIorNot: media scan pipeline, auth (Clerk), subscriptions (Stripe), realtime (Centrifugo), and object storage (MinIO).

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP | [Fiber v3](https://gofiber.io/) |
| ORM / DB | GORM + PostgreSQL 16 |
| Auth | Clerk (JWT + webhooks) |
| Billing | Stripe (Checkout, Customer Portal, webhooks) |
| Storage | MinIO (S3-compatible) |
| Messaging | RabbitMQ |
| Realtime | Centrifugo |
| Dev | Docker Compose, Air, Make |

## Architecture

```
cmd/            Entry points (api, cli, dispatcher, metadata, heuristic, aimodel)
handler/        HTTP handlers + middleware
presenter/      JSON responses
usecase/        Business logic
domain/         Entities + repository interfaces
repository/     GORM implementations
infrastructure/ External services (Stripe, Clerk, MinIO, Centrifugo, RabbitMQ…)
```

## Prerequisites

- Docker + Docker Compose
- Make
- Accounts / keys for: Clerk, Stripe (test), optional ngrok for local webhooks

## Quick start

```bash
# 1. Env
cp .env.dist .env
# Fill Clerk, Stripe, MinIO bucket, RabbitMQ, etc.

# 2. Start the stack
make dev

# 3. Migrate DB
make migrate

# 4. API
# http://localhost:4000
```

## Day-to-day commands

| Command | Description |
|---|---|
| `make dev` | Start all Compose services in background |
| `make migrate` | Build CLI in the `api` container and run GORM migrations |
| `make lint` | Run golangci-lint (with `--fix`) inside `api` |

Useful Docker helpers:

```bash
docker-compose logs -f api          # API logs
docker-compose logs -f dispatcher   # Pipeline dispatcher
docker-compose ps                   # Service status
docker-compose restart api          # Restart API only
docker-compose down                 # Stop everything
docker-compose exec api sh          # Shell in API container
```

Local (without Docker), from the host:

```bash
go run ./cmd/api
go run ./cmd/cli migrate
```

Requires a reachable Postgres and the env vars from `.env`.

## Compose services

| Service | Host port(s) | Role |
|---|---|---|
| `api` | `4000` → `3000` | HTTP API (Air hot reload) |
| `database` | `9543` → `5432` | PostgreSQL |
| `minio` | `9000`, console `9001` | Object storage |
| `mc` | — | Creates buckets + MinIO webhook to API |
| `rabbitmq` | `5000` (AMQP), `5001` (UI) | Message broker |
| `centrifugo` | `8000` | Realtime websocket / pub-sub |
| `dispatcher` | — | Dispatches scan jobs |
| `metadata` / `heuristic` / `aimodel` | — | Pipeline workers |
| `ngrok` | `4040` | Public tunnel for Clerk/Stripe webhooks |

Default credentials (dev): see `.env.dist` / Compose (MinIO `devuser` / `devpassword`, RabbitMQ from env).

## Environment

```bash
cp .env.dist .env
```

Main groups:

| Group | Variables |
|---|---|
| App / CORS | `PORT`, `GO_ENV`, `CORS_*`, `RATE_LIMIT_MAX` |
| Postgres | `DATABASE_URL`, `POSTGRES_*` |
| Clerk | `CLERK_SECRET_KEY`, `CLERK_FRONTEND_API`, `CLERK_WEBHOOK_SECRET` |
| MinIO | `STORAGE_*`, `MINIO_WEBHOOK_SECRET` |
| RabbitMQ | `RABBITMQ_*`, queue / exchange names |
| Centrifugo | `CENTRIFUGO_*` |
| Stripe | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `REDIRECT_SUCCESS_URL`, `REDIRECT_CANCEL_URL`, `REDIRECT_PORTAL_URL` |
| Sightengine | `SIGHTENGINE_*` (heuristics / external checks) |
| ngrok | `NGROK_AUTHTOKEN` |

`STORAGE_INTERNAL_ENDPOINT` is overridden in Compose for in-network MinIO (`http://minio:9000`).

## Migrations

Entities are registered in `cmd/cli/command/migrate.go`. After model changes:

```bash
make migrate
```

Order matters for FKs (quota → plan → subscription → user, then media-related tables).

## API overview

Base URL (dev): `http://localhost:4000`

### Health

| Method | Path |
|---|---|
| `GET` | `/livez` |
| `GET` | `/readyz` |
| `GET` | `/startupz` |

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/plans` | List plans (+ quotas) |

### Protected (`Authorization: Bearer <clerk_jwt>`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users/me` | Current user |
| `GET` | `/api/subscription` | Subscription + effective plan |
| `GET` | `/api/quota` | Current quota usage |
| `POST` | `/api/subscriptions/preview` | Preview plan change / proration (`{ planId }`) before confirm |
| `POST` | `/api/subscriptions` | Confirm: checkout if new (`{ url, updated:false }`), or prorated update (`{ updated:true }`). Pass `prorationDate` from preview when upgrading. |
| `GET` | `/api/subscriptions/portal` | Stripe Customer Portal URL (requires `stripeCustomerId`) |
| `GET` | `/api/invoices` | List invoices (paginated) |
| `GET` | `/api/realtime/connection` | Centrifugo connection info |
| `POST` | `/api/scans/presign-upload-url` | Presigned upload URL |
| `GET` | `/api/scans` | List scans |
| `GET` | `/api/scans/statistics` | Stats |
| `GET` | `/api/scans/:id` | Scan detail |
| `GET` | `/api/medias/:id/thumbnail` | Thumbnail |

### Webhooks

| Method | Path | Auth |
|---|---|---|
| `POST` | `/webhooks/clerk` | Svix (`CLERK_WEBHOOK_SECRET`) |
| `POST` | `/webhooks/stripe` | Stripe signature (`STRIPE_WEBHOOK_SECRET`) |
| `POST` | `/webhooks/minio/object-created` | `MINIO_WEBHOOK_SECRET` |

Stripe events handled: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`, `invoice.payment_succeeded`, `invoice.payment_failed`.

`invoice.payment_succeeded` / `invoice.payment_failed` also upsert a local `invoices` row (PDF/hosted URL, amounts, status, etc.).

Processing is async (goroutine); always respond `200` after signature validation.

## Subscriptions & quotas

- New users get a **free** subscription (`CreateFreeSubscription`).
- Paid checkout via Stripe; portal for manage/cancel.
- **Effective plan**: if status ≠ `active`, entitlements fall back to free; `plan` stays the subscribed plan.
- **Quota**: `quotaPeriodStart` is an anniversary anchor. Usage is counted over a rolling month from that date (SQL on `medias`).
- Anchor resets only on **free ↔ paid** (not on paid → paid upgrades).

Centrifugo channel `users:{userId}` receives e.g. `subscription_updated`, `payment_succeeded`, `payment_failed`.

## Local webhooks (ngrok)

1. Set `NGROK_AUTHTOKEN` in `.env`.
2. Align the ngrok `command` in `compose.yaml` with your domain (or drop `--url` for a random URL).
3. Point Clerk / Stripe webhook endpoints to the ngrok host.
4. Inspector: `http://localhost:4040`.

Stripe Customer Portal must be enabled in the Stripe Dashboard (Settings → Billing → Customer portal).

## Entry points

| Binary / cmd | Role |
|---|---|
| `cmd/api` | HTTP API |
| `cmd/cli` | Ops (`migrate`) |
| `cmd/dispatcher` | Job routing |
| `cmd/metadata` | Metadata worker |
| `cmd/heuristic` | Heuristics worker |
| `cmd/aimodel` | AI model worker |

## Production build

```bash
docker build --target production -t isaiornot-api:prod .
docker run -p 3000:3000 --env-file .env isaiornot-api:prod
```

## Troubleshooting

| Problem | What to check |
|---|---|
| DB connection refused | `docker-compose ps` — `database` healthy? Port `9543`? |
| `401` JWT | `CLERK_FRONTEND_API` matches Clerk Frontend API / issuer |
| Clerk webhook signature | `CLERK_WEBHOOK_SECRET` |
| Stripe webhook signature | `STRIPE_WEBHOOK_SECRET` + raw body |
| Portal / checkout fails | Keys, `REDIRECT_*_URL`, Stripe portal enabled, customer id present |
| MinIO upload not processed | `mc` ran, bucket name, `MINIO_WEBHOOK_SECRET`, API logs |
| Quota shows `0` after free → paid | Expected: `quotaPeriodStart` reset on free → paid |
| Port `4000` in use | Change host mapping in `compose.yaml` |
