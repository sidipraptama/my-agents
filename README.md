# my-agents

Self-hosted, always-on AI agents with a persona, a role, and your own model (BYOM). See [PRD.md](PRD.md) for the full product spec, feature status, and known gaps.

## What it is

Instead of one generic chat window, you create named **Agents** — each with a role, a persona (bio, tone, constraints, knowledge), and its own model binding. An Agent can be triggered four ways:

- **Chat** — talk to it directly in the console.
- **API** — `POST` a message to it from a script.
- **Schedule** — a cron expression that fires automatically.
- **Webhook** — a public URL an external service can `POST` to.

Every trigger produces a **Task** with a full message transcript, visible on the per-Agent chat history and the global **Task Board**.

## Features

- **Agents**: create/edit/delete, enable/pause, 5 built-in role templates (Frontend/Backend/DevOps/QA/Database Engineer) to start from.
- **Persona**: structured (bio, tone, constraints, knowledge), compiled into the system prompt on every call.
- **BYOM (Bring Your Own Model)**: Anthropic (Claude) direct, or **LiteLLM** — a generic adapter for any OpenAI-compatible endpoint (a real LiteLLM proxy, self-hosted vLLM, Ollama's compat mode, etc.), configured with just a base URL + token. Model bindings are independent objects — create once, attach to any Agent, edit or rotate the key later, delete when done.
- **Chat UX**: your message appears the instant you hit send (not stuck on "sending" for the whole model round-trip), a proper typing indicator while the model thinks, retry on failure, and Markdown rendering (headings, lists, code blocks, links) for replies.
- **Scheduling**: cron-based recurring tasks per Agent, synced live — no restart needed after creating/editing/deleting a schedule.
- **Webhooks**: token-authenticated public URLs per Agent; accepts `{"message": "..."}` JSON or raw text.
- **API-triggered tasks**: a plain REST endpoint for scripting, separate from the chat endpoint but sharing the same execution path.
- **Task Board**: every task across every Agent, list or kanban view, grouped by status.
- **Dark/light mode**, minimal Apple-inspired design.

Not built yet (see [PRD.md §7.5–7.7, §13](PRD.md)): persistent memory, approval/red-line gating, tool use (function calling), response streaming, multi-user auth.

## Architecture

```
┌─────────────┐        REST (JSON)        ┌─────────────┐        ┌──────────┐
│  Next.js    │ ───────────────────────▶  │  Go backend │ ─────▶ │ Postgres │
│  (frontend) │ ◀───────────────────────  │  (chi)      │ ◀───── │          │
└─────────────┘                           └─────────────┘        └──────────┘
                                                  │
                                                  ▼
                                       Anthropic API / LiteLLM
                                       proxy / any OpenAI-
                                       compatible endpoint
```

- **Backend** (`backend/`): Go, `chi` router, layered `api` → `service` → `store`. In-process cron scheduler (`internal/scheduler`, `robfig/cron/v3`). Model provider abstraction (`internal/llm`) — Anthropic and LiteLLM implementations share one interface. API keys encrypted at rest (AES-256-GCM, `internal/crypto`).
- **Frontend** (`frontend/`): Next.js 16 App Router, TypeScript, Tailwind v4, `next-themes` for dark/light, `react-markdown` for chat formatting.
- **Database**: Postgres 16. Migrations are flat `.sql` files in `backend/migrations/`, tracked in a `schema_migrations` table so each runs exactly once (needed since one migration is a table rename, not idempotent by nature).

### ⚠️ `INTERNAL_API_URL` vs `NEXT_PUBLIC_API_URL`

Next.js Server Components render **inside the frontend container**, so they need the backend's Docker-internal address. The browser runs on your host machine and needs the publicly reachable address. Two different env vars serve these two call sites (see `frontend/lib/api.ts`):

| Var | Used by | Value in Docker Compose |
|---|---|---|
| `INTERNAL_API_URL` | Server Components (runtime env, not baked into the build) | `http://backend:8080` |
| `NEXT_PUBLIC_API_URL` | Browser code (baked in at build time) | `http://localhost:8080` |

If you ever see an Agent's detail page 404 while the list page works fine, this is almost certainly why.

## Run with Docker Compose

```bash
cp .env.example .env
# edit .env: set ENCRYPTION_KEY (see below)
docker-compose up --build
```

> **Note:** this project uses the standalone `docker-compose` binary (hyphenated). If your machine instead has the `docker compose` CLI plugin, that works too — check with `docker compose version`.

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

First-time setup in the UI: go to **Model Bindings**, add a Claude API key (or a LiteLLM base URL + token), then create an **Agent** and attach the binding from its detail page.

### Environment variables

| Var | Required | Default | Notes |
|---|---|---|---|
| `ENCRYPTION_KEY` | Yes | — (backend refuses to start without it) | Must be exactly 32 characters — used as an AES-256 key to encrypt model API keys at rest. Generate with `openssl rand -base64 24`. |
| `ALLOWED_ORIGIN` | No | `http://localhost:3000` | CORS origin the backend accepts browser requests from. |
| `NEXT_PUBLIC_API_URL` | No | `http://localhost:8080` | Public backend URL, baked into the frontend build for browser-side calls. |
| `DATABASE_URL` | No (Docker sets it) | `postgres://myagents:myagents@localhost:5432/myagents?sslmode=disable` | Only relevant when running the backend outside Docker. |
| `PORT` | No | `8080` | Backend listen port. |

Regenerating `ENCRYPTION_KEY` after model bindings already exist makes their stored keys undecryptable — you'd need to re-add them.

## Run locally without Docker

**Backend** (needs Postgres running locally):

```bash
cd backend
export DATABASE_URL="postgres://myagents:myagents@localhost:5432/myagents?sslmode=disable"
export ENCRYPTION_KEY="$(openssl rand -base64 24)"
go run ./cmd/server
```

**Frontend**:

```bash
cd frontend
npm install
npm run dev
```

## API reference

All endpoints are under `/api`. No authentication beyond `ENCRYPTION_KEY`-at-rest for secrets and per-webhook tokens — this is a self-hosted, trusted-network tool, not a multi-tenant service.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/agents` | List Agents |
| `POST` | `/agents` | Create an Agent |
| `GET` | `/agents/{id}` | Get an Agent |
| `PATCH` | `/agents/{id}` | Update an Agent (merges against existing — omitted fields are preserved) |
| `DELETE` | `/agents/{id}` | Delete an Agent |
| `POST` | `/agents/{id}/chat` | Send a chat message (trigger type `manual`) |
| `GET` | `/agents/{id}/tasks` | List an Agent's tasks |
| `POST` | `/agents/{id}/tasks` | Run a task programmatically (trigger type `api`) |
| `GET` | `/agents/{id}/schedules` | List an Agent's schedules |
| `POST` | `/agents/{id}/schedules` | Create a schedule (`cron_expr`, `input`) |
| `PATCH` | `/schedules/{id}` | Update a schedule |
| `DELETE` | `/schedules/{id}` | Delete a schedule |
| `GET` | `/agents/{id}/webhooks` | List an Agent's webhooks |
| `POST` | `/agents/{id}/webhooks` | Create a webhook (returns its token) |
| `DELETE` | `/webhooks/{id}` | Delete a webhook |
| `POST` | `/hooks/{token}` | **Public.** Trigger a task via webhook (trigger type `webhook`) |
| `GET` | `/tasks` | Task board — every task across every Agent (capped at 200, most recent first) |
| `GET` | `/bindings` | List model bindings (API keys never included) |
| `POST` | `/bindings` | Create a model binding |
| `PATCH` | `/bindings/{id}` | Update a binding (omit `api_key` to keep the existing one) |
| `DELETE` | `/bindings/{id}` | Delete a binding (Agents using it fall back to unbound) |

**Trigger a task from a script:**

```bash
curl -X POST http://localhost:8080/api/agents/<agent-id>/tasks \
  -H "Content-Type: application/json" \
  -d '{"message": "Summarize last night'\''s error logs"}'
```

**Trigger via webhook** (create one first from the Agent's detail page, then):

```bash
curl -X POST http://localhost:8080/api/hooks/<token> \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy finished"}'
```

## Project structure

```
backend/
  cmd/server/         entrypoint
  internal/api/       HTTP handlers + routing
  internal/service/   shared task-execution logic (used by chat, API, schedules, webhooks)
  internal/store/     Postgres queries
  internal/llm/       model provider abstraction (Anthropic, LiteLLM)
  internal/scheduler/ in-process cron scheduler
  internal/persona/   persona → system prompt compiler
  internal/crypto/    AES-256-GCM for secrets at rest
  migrations/         flat .sql files, applied in order, tracked in schema_migrations
frontend/
  app/                Next.js App Router pages (agents, tasks, bindings)
  components/         UI components, incl. components/ui/ design-system primitives
  lib/                API client, types, templates
docker-compose.yml    postgres + backend + frontend
PRD.md                full product spec, feature status, phasing
```

## Status

Phase 1 and Phase 2 of the PRD are shipped: Agent CRUD with personas and templates, BYOM (Claude + LiteLLM), manual/API/scheduled/webhook triggers, and the task board. Memory, approval gating, tool use, and response streaming are not yet built — see [PRD.md §11, §13](PRD.md) for the honest list of gaps and open questions.
