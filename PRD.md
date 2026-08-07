# PRD: my-agents

**Status:** Phase 1 + Phase 2 shipped. Phase 3 not started.
**Owner:** span26
**Last updated:** 2026-08-07

## 1. Summary

`my-agents` is a self-hosted, always-on AI agent platform — a personal/small-team version of Alibaba's QoderWake. Each Agent has a persistent identity (name, role, persona) and a job. Unlike a chat assistant, an Agent keeps state across sessions and can be triggered by schedule, webhook, or API call — not just by a human typing a prompt — and is model-agnostic: you plug in your own model via API key (BYOM).

The two things this project deliberately does that QoderWake doesn't give you:
1. **Full control over the model provider** — no vendor lock to Qwen/Alibaba's model roster. Bring your own Claude API key, or point at any OpenAI-compatible endpoint (LiteLLM proxy, self-hosted vLLM/Ollama, etc.) via URL + token.
2. **Full control over the persona system** — persona is a first-class, editable object, not a fixed role template. Templates exist only as a starting point at creation time.

## 2. Problem / Motivation

- Existing "always-on agent" platforms (QoderWake, etc.) bundle the orchestration layer with a fixed set of hosted models — you don't get to choose the model provider, and data goes through their cloud.
- For personal/DevOps use, a lighter self-hosted version that talks directly to the Claude API (or any OpenAI-compatible endpoint) is more useful than a fully managed platform.
- No existing self-hosted tool cleanly combines: persistent agent identity + persona, BYOM, and a task/schedule model in one small stack.

## 3. Goals

- Run one or more persistent Agents, each with a name, role, and persona (tone, constraints, domain knowledge). **✅ Shipped.**
- Bring your own model: configure any model provider via API key. **✅ Shipped** — Anthropic (Claude) direct, and a generic OpenAI-compatible adapter (branded "LiteLLM" in the UI, works with any server speaking the `/chat/completions` contract).
- Assign work to an Agent via: direct chat, scheduled cron job, incoming webhook, or REST API call. **✅ Shipped**, all four trigger types.
- Give each Agent persistent memory (facts, past task summaries, preferences) that survives across sessions. **❌ Not built** — deferred to Phase 3 (see §11).
- Provide a simple web console: manage Agents, chat with an Agent, see a task board of what's running/done/failed. **✅ Shipped.**
- Support a basic approval gate: an Agent can be configured to pause and wait for human approval before "red line" actions. **❌ Not built** — deferred to Phase 3.
- Fully self-hostable via Docker Compose, single binary + Postgres, no mandatory external dependency besides the model provider's API. **✅ Shipped.**

## 4. Non-goals (v1)

- Multi-tenant SaaS, billing, org-level RBAC/SSO — this is single-user/small-team, not enterprise.
- Building/hosting our own model — BYOM only, no local model inference engine bundled.
- Full IDE / codebase-wide "RepoWiki" style code understanding — out of scope; an Agent can still be given repo access as a tool if needed later, but deep codebase indexing is not planned.
- Mobile app. Web console only.
- Response streaming (SSE/WebSocket) — chat and API-triggered tasks are synchronous request/response; the caller waits for the full model reply. See §8 and §11.

## 5. Users

- Primary: the author, self-hosting for personal DevOps/automation use (e.g., an Agent that triages alerts, drafts PRs, summarizes logs).
- Secondary: small teams who want an internal, self-hosted alternative to hosted agent platforms without giving up model choice.

## 6. Core Concepts

| Term | Meaning |
|---|---|
| **Agent** | A persistent agent instance: identity + persona + role + assigned model binding. Can be enabled or paused. |
| **Persona** | Editable config attached to an Agent: bio, tone, constraints list, knowledge list. Compiled into a system prompt at runtime. |
| **Model Binding** | The model provider + credentials + model name (+ optional config like a base URL) an Agent uses. Created independently, then attached to one or more Agents; editable and deletable. |
| **Task** | A unit of work run against an Agent — has a status (`running` → `done`/`failed`), a trigger type, an input, and a full message transcript. |
| **Trigger** | What starts a task: manual chat message, cron schedule, inbound webhook, or direct API call. |
| **Schedule** | A cron expression + a fixed input message, attached to an Agent. Fires automatically while enabled. |
| **Webhook** | A random unguessable token attached to an Agent. POSTing to its public URL runs a task with the request body as input. |
| **Memory** *(not built)* | Planned: structured, inspectable long-term store per Agent. Phase 3. |
| **Red Line Rule** *(not built)* | Planned: a per-Agent policy requiring human approval before matching actions. Phase 3. |

## 7. Features

### 7.1 Agent Management — ✅ Shipped
- Create/edit/delete an Agent: name, role (free text), persona, assigned model binding, enabled/paused state.
- Creation flow offers 5 built-in templates (Frontend/Backend/DevOps/QA/Database Engineer) that pre-fill name, role, and persona — fully editable before submit, and not pre-seeded as real Agents.
- Enable/disable an Agent. A paused Agent rejects every trigger type (manual, scheduled, webhook, API) with a clear error — enforced once, centrally, in the task execution service.

### 7.2 Persona System — ✅ Shipped
- Persona is a structured object: `bio`, `tone`, `constraints[]`, `knowledge[]`.
- Compiles into a system prompt prefix at runtime (`internal/persona`); editing persona takes effect on the next task, no redeploy needed.

### 7.3 Bring Your Own Model (BYOM) — ✅ Shipped
- Provider abstraction (`internal/llm.Provider`) with a single `Send(ctx, Request) (string, error)` method.
- **Anthropic (Claude)**: direct integration against `api.anthropic.com/v1/messages`.
- **LiteLLM**: generic adapter against the OpenAI `/chat/completions` contract — works with an actual LiteLLM proxy, and equally with any other OpenAI-compatible server (self-hosted vLLM, Ollama's compat mode, OpenAI itself). Requires `provider: "litellm"`, a `model_name`, a `config.base_url`, and a token.
- API keys/tokens encrypted at rest with AES-256-GCM (`internal/crypto`), keyed by a 32-byte `ENCRYPTION_KEY` env var. Never returned in any API response after creation.
- Model bindings are independent objects: create once, attach to any number of Agents, edit (including rotating the key without re-entering it if unchanged) or delete at any time from the **Model Bindings** page.

### 7.4 Tasks & Triggers — ✅ Shipped
- **Manual**: chat with an Agent in the console (`POST /api/agents/{id}/chat`).
- **API**: any external caller can POST a message directly (`POST /api/agents/{id}/tasks`) — same execution path as chat, different trigger label.
- **Scheduled**: cron expression + fixed input message per Agent, run by an in-process scheduler (`robfig/cron/v3`). Schedules are loaded from the DB at boot and kept in sync with the running cron instance on every create/update/delete — no restart needed to pick up changes.
- **Webhook**: each Agent can have one or more webhook URLs (`POST /api/hooks/{token}`), authenticated only by the token in the URL. Accepts `{"message": "..."}` / `{"input": "..."}` JSON, or falls back to raw request body text.
- Task lifecycle as actually implemented: a task is created directly in `running` state (no separate `queued` state today — execution starts synchronously in the same request) and ends in `done` or `failed`. The `queued` status value exists in the schema for forward-compatibility with a real queue, but nothing produces it yet.
- Every task stores its full message transcript (user input + assistant reply) for audit/debugging, and prior tasks for the same Agent are replayed as conversation history on the next call.

### 7.5 Memory — ❌ Not built (Phase 3)
- `memory_entries` table exists in the schema but nothing reads or writes it yet.

### 7.6 Approval / Red Line Rules — ❌ Not built (Phase 3)
- No `red_line_rules` table exists yet. Every trigger runs to completion unattended (subject only to the Agent being enabled and having a model binding).

### 7.7 Tools — ❌ Not built
- No function-calling/tool-use plugin system yet. An Agent's only capability today is generating a text reply from its persona + conversation history.

### 7.8 Web Console (Next.js) — ✅ Shipped
- **Agents** (`/agents`): list, create (with template picker), delete.
- **Agent detail** (`/agents/[id]`): persona display, model binding picker, schedules manager, webhooks manager (with copy-to-clipboard URLs), a ready-to-copy `curl` snippet for the API-trigger endpoint, chat window, delete.
- **Chat**: optimistic message send (appears instantly, not stuck on "sending" for the whole model round-trip), a distinct failed/retry state, an animated typing indicator while the model is generating, and Markdown rendering (headings, lists, code blocks, links) for assistant replies.
- **Model Bindings** (`/bindings`): list, create, inline edit (provider/model/base URL/rotate token), delete.
- **Task Board** (`/tasks`): every task across every Agent, toggle between a flat list view and a kanban/swimlane view grouped by status.
- Dark/light mode (system-aware, manually toggleable), Apple-inspired minimal design system (`components/ui/*`).
- No memory viewer, no approval queue — not built yet (§7.5, §7.6).

## 8. Tech Stack (as implemented)

| Layer | Choice | Notes |
|---|---|---|
| Backend | **Go**, `chi` router | Single static binary; layered `api` → `service` → `store` packages. |
| Frontend | **Next.js 16** (App Router, TypeScript), Tailwind v4 | `next-themes` for dark/light; `react-markdown` + `remark-gfm` for chat formatting. |
| API between them | **REST, synchronous** — no streaming | A chat/task call blocks until the model finishes generating. Simpler to implement and self-host than SSE/WebSockets; revisit if response latency becomes annoying (see §11). |
| Database | **Postgres 16**, `pgx` driver | JSONB for persona and model-binding config; plain columns elsewhere. |
| Migrations | Flat `.sql` files in `backend/migrations/`, tracked in a `schema_migrations` table so each runs exactly once | Needed once a non-idempotent migration (a table/column rename) was introduced — earlier versions re-ran every file on every boot, which only worked by luck because those files happened to be idempotent. |
| Scheduler | In-process `robfig/cron/v3` instance inside the backend binary | Loaded from the `schedules` table at boot; synced live on schedule CRUD. No Redis/queue dependency. |
| Model providers | Anthropic Claude API; generic OpenAI-compatible adapter (LiteLLM) | Both implement the same `Provider` interface — adding a third provider means one new file. |
| Secrets | AES-256-GCM encrypted columns, key from `ENCRYPTION_KEY` env var | Self-hosted — no managed secrets service assumed. |
| Deployment | Docker Compose: `postgres` + `backend` + `frontend` services | See §12 gotcha on `INTERNAL_API_URL` vs `NEXT_PUBLIC_API_URL`. |

## 9. Data Model (as implemented)

- `agents`: id, name, role, persona (jsonb), model_binding_id (nullable FK), enabled, created_at, updated_at.
- `model_bindings`: id, provider, model_name, api_key_encrypted (bytea), config (jsonb), created_at.
- `tasks`: id, agent_id (FK, cascade delete), trigger_type, status, input, transcript (jsonb array), created_at, updated_at.
- `memory_entries`: id, agent_id (FK), entry_type, content, created_at, updated_at. *(Table exists, unused — Phase 3.)*
- `schedules`: id, agent_id (FK, cascade delete), cron_expr, input, enabled, created_at, updated_at.
- `webhooks`: id, agent_id (FK, cascade delete), token (unique), created_at.
- `schema_migrations`: filename (PK), applied_at — migration bookkeeping, not a domain table.

Not yet built: `red_line_rules` (Phase 3).

Historical note: the core entity was originally named `Waker` (borrowed from QoderWake's own terminology) and renamed to `Agent` for consistency with the project's own name. Migration `0003_rename_wakers_to_agents.sql` performed the rename on a live database with real data; migrations `0001`/`0002` still say `waker`/`wakers` internally as an accurate historical record and are not rewritten.

## 10. Success Criteria

- ✅ Can create an Agent, attach a persona, bind it to a Claude or LiteLLM binding, and have it complete a manually-triggered chat task end-to-end.
- ✅ Can define a cron schedule and see it execute unattended, producing a `scheduled`-trigger task with the correct output — verified live against a real LiteLLM proxy (7 consecutive runs, all completed).
- ✅ Can create a webhook and POST to its public URL to produce a `webhook`-trigger task.
- ✅ Can POST directly to the API-task endpoint to produce an `api`-trigger task, with no console interaction.
- ✅ Task Board shows all of the above, correctly grouped by status, linking back to the originating Agent.
- ✅ Entire stack runs via `docker-compose up` with no external SaaS dependency other than the model provider's API.
- ❌ Red-line rule pausing a task for approval — not built (Phase 3), so not yet verifiable.
- ❌ Memory recalled across unrelated sessions — not built (Phase 3), so not yet verifiable.

## 11. Open Questions / Known Gaps

- **No streaming**: chat responses arrive all at once when the model finishes, not token-by-token. Fine for short replies; noticeably slow-feeling for long ones. Candidate fix: SSE from `/chat`, one incremental change to the existing synchronous handler rather than a rearchitecture.
- **No delete-task endpoint/UI**: task rows accumulate forever (the task board query caps at 200 rows, but nothing prunes the table). Low priority until it's actually a problem.
- **Memory retrieval strategy** (once Phase 3 starts): is naive recency-based injection good enough, or is a vector store needed sooner than expected?
- **Tool sandboxing** (once tool-use is built): how strict should shell/file tool access be by default?
- Single-user auth only — no login at all today. Fine for a personal self-hosted instance on a private network; would need addressing before exposing this beyond localhost/a trusted LAN.

## 12. Operational Gotcha Worth Documenting

Next.js Server Components run **inside the frontend container**, so they need the backend's Docker-internal address (`http://backend:8080`, via a plain `INTERNAL_API_URL` runtime env var) — while the browser runs on the host and needs the publicly reachable address (`http://localhost:8080`, via `NEXT_PUBLIC_API_URL`, baked in at build time). Using one URL for both breaks one of the two call sites. This bit us once already (every Agent detail page 404'd via SSR while the list page looked fine) — see `frontend/lib/api.ts`.

## 13. Phasing

- **Phase 1 (MVP) — ✅ Done.** Agent CRUD, structured persona, Claude API binding, manual chat, task transcript storage, basic console.
- **Phase 2 — ✅ Done.** Scheduling (cron), webhooks, API-triggered tasks, task board (list + kanban). Plus, done opportunistically alongside: LiteLLM/BYOM second provider, model-binding edit/delete, Agent creation templates, chat UX (optimistic send, Markdown rendering, thinking indicator), the Waker→Agent rename, and a migration-tracking fix.
- **Phase 3 — Not started.** Memory system (view/edit/delete), red-line approval flow.
- **Phase 4 — Not started.** Tool plugin system (function calling), multi-user basics, response streaming.
