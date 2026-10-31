# Dream11 — Watch Parties & Match Buddy

A synchronized watch-party platform with **Match Buddy** — an AI-powered live fantasy cricket companion designed to solve the real problem fans face during a live sports match.

When friends watch cricket together, they are juggling the live video stream, their fantasy app, and WhatsApp, unable to easily tell how each ball affects their fantasy points versus their friends. **Match Buddy** lives directly inside the room's sidebar and chat, computing deterministic fantasy point deltas after every ball, projecting strategic "What-If" scenarios, maintaining an audit trail of team changes, and delivering AI-powered commentary and banter via Google Gemini (with seamless local template fallback).

---

## 🏆 Match Buddy (AI Fantasy Companion)

### The Fan Problem
During live matches, fans constantly calculate in their heads: *"Did that six push me ahead of Rahul? What if Bumrah takes a wicket next ball?"* Constantly switching between a broadcast and a fantasy scoreboard creates disjointed social viewing.

### How Match Buddy Solves It
1. **Real-time Ball-by-Ball Fantasy Tracker**: Driven by a server-side replay engine (Mumbai Blasters vs Delhi Titans, 36 legal balls across 6 dramatic overs), broadcasting live point changes and rank movements to everyone in the room.
2. **Deterministic Dream11 Scoring System**:
   - **Batting**: +1 per run, +4 boundary bonus (4s), +6 six bonus (6s), +8 at 30 runs, +16 at 50 runs (replaces 30), -2 duck for top/middle order.
   - **Bowling**: +25 per wicket (excluding run-outs), +4 maiden over.
   - **Multipliers**: Captain gets **2x**, Vice-Captain gets **1.5x**.
   - **Squad & Presets**: 16-player demo roster. Build custom 11s (max 7 per real team) or lock in with 1-click presets (*Balanced XI*, *Bowler Heavy*, *Finisher Heavy*).
3. **Pre-Match Team Audit Log**: Every lineup creation, captain change, or player swap is broadcast and logged with exact timestamps so room members know nobody altered their team after match lock.
4. **"What-If" Scenario Simulator**: Test hypothetical deliveries instantly (*"What if striker hits a six?"*, *"What if bowler takes a wicket?"*, *"What if dot?"*) with immediate point projections and leader impact before the ball is bowled.
5. **AI Companion & Commentary (Google Gemini)**:
   - Uses Google Gemini API (`gemini-1.5-flash`) for natural language over summaries, swing reactions, Q&A ("Who is winning?", "Who is the top performer?"), and catch-up recaps.
   - **Resilience Guarantee**: Strict 6-second timeout with 1 retry. If `GEMINI_API_KEY` is absent or unreachable, Match Buddy automatically falls back to the built-in authentic cricket commentary engine without disrupting the match.
6. **Solo Rivalry Bot**: Don't have a friend online right now? Click **"Add Bot (Rahul)"** to add a virtual friend with a Bowler Heavy squad for instant head-to-head competition.
7. **Simulated Demo Controls**: Adjust simulation speed (1x, 2x, 5x), skip over, pause, resume, or reset back to lobby anytime.

---

## ⚡ 60-Second Quick Demo Guide

1. **Launch Stack**:
   - Start backend: `cd backend && go run ./cmd/server` (or `make dev`)
   - Start frontend: `cd frontend && npm run dev`
2. **Open Watch Party Room**:
   - Navigate to `http://localhost:5173` and create or enter a room.
   - Click the **Trophy icon** in the top header or side panel tabs to open **Match Buddy**.
3. **Pick Your Fantasy Team**:
   - Click **"Edit Fantasy Team"** and select a 1-click preset like *Balanced XI* (or customize your 11, Captain, and Vice-Captain).
   - Click **"Save Lineup"**. Notice the audit entry appear in the Pre-Match Changes log!
4. **Enable Virtual Rival**:
   - Click **"Add Bot (Rahul)"** in the simulation controls. Notice Rahul joins the leaderboard with his bowler-heavy lineup.
5. **Start Live Simulation**:
   - Click **"Start Match"** (toggle to **2x** or **5x** for rapid replay).
   - Watch live scores update ball-by-ball with points flashing (+12, +25, etc.), the leaderboard rank flipping, and Match Buddy celebrating key swings!
6. **Test "What-If" & AI Q&A**:
   - Click a **"What if SIX?"** or **"What if WICKET?"** chip to see projected point deltas.
   - Type a question into Ask Buddy: *"Who is winning?"* or click **"Catch Me Up"** to receive an instant recap.

---

## Features (Inox Core Watch-Party Platform)

- **Synchronized playback.** Authoritative playback state per room with leader sync for VOD and live channels.
- **Real-time chat** persisted to Postgres and replayed on join.
- **Voice chat and screen sharing** through a Pion-backed selective forwarding unit (SFU).
- **Adaptive-bitrate media pipeline.** HLS transcoding (1080p/720p/480p) via asynq background worker.
- **Rooms, invitations and friendships** with per-member roles and permissions.
- **Live telemetry** streamed to the admin portal, plus Prometheus metrics.

## Architecture

```
  frontend/                    admin-portal/
  (watch-party client)         (operator dashboard)
        |                             |
        |  HTTPS REST                 |  HTTPS REST
        |  WSS room socket            |  WSS telemetry socket
        v                             v
  +--------------------------------------------------+
  |                    backend/                      |
  |                                                  |
  |   api/        stdlib net/http routing,           |
  |               CORS -> logging -> metrics         |
  |   ws/         one socket per room; chat,         |
  |               playback sync and SFU signalling   |
  |   sfu/        Pion SFU, one room per watch party |
  |   media/      upload, transcode queue, HLS       |
  |   live/       guarded upstream stream proxy      |
  |   auth/       Redis-backed sessions              |
  +--------------------------------------------------+
        |              |              |
        v              v              v
   PostgreSQL        Redis          MinIO
   (durable state)   (sessions,     (media objects)
                      queue,
                      event bus)
                         ^
                         |
                    cmd/worker
                (ffmpeg HLS transcoding)
```

`app.App.Run()` is the single composition root: it wires repositories, services,
the WebSocket hub and the SFU manager, then owns graceful shutdown. Each domain
package exposes a repository and a service; `domain/` holds shared types.

### Real-time protocol

Chat, playback control (`PLAY`, `PAUSE`, `SEEK`, `CHANGE_MEDIA`,
`SYNC_PLAYBACK`) and WebRTC signalling (`SFU_OFFER`, `SFU_ANSWER`,
`SFU_ICE_CANDIDATE`) all travel over a single WebSocket per room as typed JSON
events. Audio and video themselves never cross that socket; only SDP and ICE do,
with RTP routed by the SFU.

When Redis is configured, a pub/sub event bus mirrors events between processes
so the hub scales horizontally.

### Authentication

Sessions, not JWTs. `auth.SessionStore` persists sessions in Redis under the
`inox_session` cookie. Because the clients and the API are served from different
origins in production, the cookie is `SameSite=None; Secure` there and `Lax` in
development. `RequireAuth` also accepts `Authorization: Bearer`, `X-Session-ID`
or `?session_id=`, which is what lets the WebSocket handshake carry a session at
all, since browsers cannot set headers on it.

There is no role column on `users`. Administrative access is granted solely by
listing an address in `ADMIN_EMAILS`; with that unset, every admin route returns
403.

## Repository layout

```
backend/
  cmd/server        API and WebSocket server
  cmd/worker        asynq worker, ffmpeg transcoding
  cmd/migrate       SQL migration runner
  internal/         domain packages (see architecture above)
  migrations/       numbered, forward-only SQL migrations
frontend/           watch-party client
admin-portal/       operator dashboard
deploy/
  oracle/           single-VM production stack and CI deployment
  kubernetes/       Kustomize manifests
```

## Prerequisites

- Go 1.25 or newer
- Node.js 24 or newer
- Docker with the Compose plugin
- ffmpeg, if you intend to run the transcoding worker outside Docker

## Getting started

```bash
git clone https://github.com/SohamBalwant/Dream11.git
cd Dream11
cp .env.example .env

# Start Postgres, Redis, MinIO and pgAdmin
make up

# Apply the schema
cd backend && go run ./cmd/migrate --command=up && cd ..

# Install client dependencies
(cd frontend && npm ci)
(cd admin-portal && npm ci)

# Run the backend and both clients together
make dev
```

| Service | URL |
| --- | --- |
| Backend API | http://localhost:8080 |
| Watch-party client | http://localhost:5173 |
| Admin portal | http://localhost:3000 |
| MinIO console | http://localhost:9001 |
| pgAdmin | http://localhost:5050 |

`make dev` runs the three services under `concurrently`. Two alternatives exist
for the same set: `./dev.sh` uses a tmux session named `inox`, and `mprocs`
reads `mprocs.yaml` and additionally runs the transcoding worker.

To exercise media transcoding locally, start the worker alongside:

```bash
cd backend && go run ./cmd/worker
```

## Configuration

The backend reads configuration from the environment, falling back to a `.env`
file in the working directory or either of its two parents. Defaults suit local
development; production is stricter.

| Variable | Default | Notes |
| --- | --- | --- |
| `APP_ENV` | `development` | `production` and `staging` enable strict checks |
| `HTTP_PORT` | `8080` | |
| `LOG_LEVEL` | `debug` | |
| `DATABASE_URL` | local Postgres DSN | Required |
| `REDIS_URL` | `redis://localhost:6379/0` | Required |
| `SESSION_SECRET` | development placeholder | Must be set and at least 32 characters outside development |
| `SESSION_DURATION_HOURS` | `168` | |
| `CORS_ALLOWED_ORIGINS` | localhost origins | Exact match, comma separated, no wildcards |
| `ADMIN_EMAILS` | unset | The only thing granting admin access |
| `MINIO_ENDPOINT` | `localhost:9000` | Falls back to local disk if unreachable at boot |
| `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` | `minioadmin` | |
| `MINIO_BUCKET_NAME` | `inox-media` | Created on startup if absent |
| `STORAGE_DIR` | `./storage_data` | Local-disk fallback location |
| `MEDIA_STREAM_BASE_URL` | `http://localhost:8080/media/stream` | Baked into asset URLs |
| `LIVE_SOURCE_ALLOWED_HOSTS` | unset | Outbound fetch allowlist; empty refuses every live channel in production |
| `LIVE_BROWSER_PATH` | found on `PATH` | Chromium for the `browser` live resolver; `off` disables it |
| `LIVE_BROWSER_MAX_CONCURRENCY` | `2` | Channel pages resolved at once; the rest queue |
| `LIVE_BROWSER_TIMEOUT_SECONDS` | `30` | How long a page gets to request its manifest |
| `LIVE_BROWSER_NO_SANDBOX` | `false` | Only where Chromium cannot create its sandbox |
| `WEBRTC_ICE_SERVERS` | public STUN servers | Comma separated |
| `WEBRTC_PORT_MIN` / `WEBRTC_PORT_MAX` | `50000` / `50100` | UDP range the SFU binds media to |
| `WEBRTC_PUBLIC_IP` | unset | Required behind NAT; see operational notes |
| `METRICS_AUTH_TOKEN` | unset | Bearer token accepted by `/metrics` |

Client applications are configured at build time, since Vite inlines these
values:

```
# frontend/.env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/api/v1
VITE_MEDIA_STREAM_BASE_URL=http://localhost:8080/media/stream

# admin-portal/.env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_TELEMETRY_WS_URL=ws://localhost:8080/api/v1/admin/telemetry/ws
```

Neither client calls the API with a bare `fetch`. Each has one client module
that attaches the base URL and credentials, `frontend/src/api/client.ts` and
`admin-portal/src/lib/api.ts`. Route new calls through those; a relative
`fetch('/api/v1/...')` works only behind the Vite dev proxy and returns 404 once
deployed.

## Commands

### Backend

```bash
go run ./cmd/server                                  # API and WebSocket server
go run ./cmd/worker                                  # transcoding worker
go run ./cmd/migrate --command=up|down|status        # migrations
go build ./...
go vet ./...
go test ./...                                        # full suite
go test -race ./...                                  # as CI runs it
go test ./internal/auth/... -run TestName -v         # a single test
```

The migration runner is hand-written rather than golang-migrate. Run it with
`backend/` as the working directory, or pass `--db`. `--steps=N` limits how many
migrations are applied.

Tests use the standard library `testing` package with hand-written in-memory
fakes. There is no mocking framework and no assertion library.

### Clients

Both applications expose the same scripts.

```bash
npm run dev                                # Vite dev server
npm run build                              # tsc -b && vite build
npm run lint                               # ESLint
npx tsc --noEmit -p tsconfig.app.json      # typecheck only, as CI runs it
```

Neither client has a test runner configured, so `npm test` does nothing.

### Infrastructure

```bash
make up      # start Postgres, Redis, MinIO, pgAdmin
make down    # stop them
make dev     # infrastructure plus all three services
make clean   # stop containers and kill the tmux session
```

## Deployment

The backend runs on a single Oracle Cloud Ampere (arm64) instance behind Caddy,
which terminates TLS and serves the admin bundle. Both clients are also
deployed to Vercel.

Pushing to `main` deploys. The workflow in `.github/workflows/ci.yml` runs lint,
tests and typechecks, builds the backend image on a native arm64 runner,
publishes it to GHCR tagged with the commit SHA, then pulls that tag on the host
and restarts the stack behind a health check. Pull requests build and test but
never deploy.

The image is deliberately not built on the target: a single core makes that slow
and a failure would leave production half-built. Compose runs with `--no-build`,
so a missing image fails loudly instead of quietly compiling on the host.

`deploy/oracle/README.md` covers first-time provisioning, the networking model,
rollback and the manual fallback path.

## Operational notes

These are the non-obvious constraints worth knowing before changing related
code.

- **The SFU must know its public address.** A NAT'd cloud instance only ever
  sees its private IP, so without `WEBRTC_PUBLIC_IP` the SFU advertises ICE
  candidates no remote browser can route to and connections never leave
  `checking`. The configured UDP port range must also be open in the host
  firewall and in any cloud-level security list.

- **CORS is exact-match.** Origins are compared literally, with no wildcard
  support and no trailing slashes. Preview deployments have their own hostnames
  and must be listed explicitly. The WebSocket upgraders consult the same
  allowlist, so an unlisted origin fails the handshake as well as REST calls.

- **A clean ffmpeg exit does not mean a complete transcode.** ffmpeg reports
  success after a short read of its input exactly as it does after a real one,
  so the processor compares the encoded duration against the source length and
  refuses to publish anything shorter.

- **Storage degrades rather than fails.** If MinIO is unreachable at boot the
  backend falls back to local-filesystem storage automatically, and
  `ReconcileOrphanedAssets` recovers transcode jobs interrupted by a crash.

- **Demo mode must stay honest.** Admin panels fall back to built-in fixtures
  when the backend is unreachable. Those fixtures must never be reported as a
  live connection, mixed into chart history, or used to decide whether the
  backend is up.

## Contributing

Husky and lint-staged run on commit: `gofmt` and `goimports` over changed Go
files, Prettier over changed client files. Install the hooks once with
`npm install` at the repository root.

CI must pass before merge: `go vet`, `go test -race`, and lint plus typecheck
for both clients.

## License

Released under the [MIT License](LICENSE).
