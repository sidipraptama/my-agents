# PRD: my-agents

**Status:** Draft
**Owner:** span26
**Last updated:** 2026-07-17

## 1. Summary

`my-agents` is a self-hosted, always-on AI agent platform — a personal/small-team version of Alibaba's QoderWake. Each Agent has a persistent identity (name, role, persona) and a job. Unlike a chat assistant, an Agent keeps state across sessions, can be triggered by schedule/event/API — not just by a human typing a prompt — and is model-agnostic: you plug in your own model via API key (BYOM), starting with the Claude API.

The two things this project deliberately does that QoderWake doesn't give you:
1. **Full control over the model provider** — no vendor lock to Qwen/Alibaba's model roster.
2. **Full control over the persona system** — persona is a first-class, editable object, not a fixed role template.

## 2. Problem / Motivation

- Existing "always-on agent" platforms (QoderWake, etc.) bundle the orchestration layer with a fixed set of hosted models — you don't get to choose the model provider, and data goes through their cloud.
- For personal/DevOps use (this user's context: managing infra, scripts, small SaaS tooling), a lighter self-hosted version that talks directly to the Claude API (or any OpenAI-compatible endpoint) is more useful than a fully managed platform.
- No existing self-hosted tool cleanly combines: persistent agent identity + persona, BYOM, and a task/schedule model in one small stack.

## 3. Goals

- Run one or more persistent agents ("Agents"), each with a name, role, and persona (tone, constraints, domain knowledge).
- Bring your own model: configure any model provider via API key, first-class support for Anthropic (Claude) API, designed so other providers (OpenAI-compatible, local Ollama, etc.) can be added later without rearchitecting.
- Assign work to an Agent via: direct chat, scheduled cron job, incoming webhook/event, or REST API call.
- Give each Agent persistent memory (facts, past task summaries, preferences) that survives across sessions and that a human can view/edit/delete.
- Provide a simple web console: manage Agents, chat with an Agent, see a task board of what's running/done/failed.
- Support a basic approval gate: an Agent can be configured to pause and wait for human approval before "red line" actions (e.g. deploying, deleting, sending external messages).
- Fully self-hostable via Docker Compose, single binary + Postgres, no mandatory external dependency besides the model provider's API.

## 4. Non-goals (v1)

- Multi-tenant SaaS, billing, org-level RBAC/SSO — this is single-user/small-team, not enterprise.
- Building/hosting our own model — BYOM only, no local model inference engine bundled.
- Full IDE / codebase-wide "RepoWiki" style code understanding — out of scope for v1; an Agent can still be given repo access as a tool if needed, but deep codebase indexing is a later phase, not v1.
- Mobile app. Web console only.

## 5. Users

- Primary: the author, self-hosting for personal DevOps/automation use (e.g., an Agent that triages alerts, drafts PRs, summarizes logs).
- Secondary: small teams who want an internal, self-hosted alternative to hosted agent platforms without giving up model choice.

## 6. Core Concepts

| Term | Meaning |
|---|---|
| **Agent** | A persistent agent instance: identity + persona + role + assigned model + memory + tools. |
| **Persona** | Editable config attached to an Agent: system prompt fragment, tone/style rules, domain constraints, do/don't list. |
| **Model Binding** | The model provider + credentials + model name an Agent uses. Swappable per-Agent. |
| **Task** | A unit of work assigned to an Agent — has a status (queued/running/waiting-approval/done/failed), a trigger source, and a transcript. |
| **Trigger** | What starts a task: manual chat message, cron schedule, inbound webhook, API call. |
| **Memory** | Structured, inspectable long-term store per Agent: facts, past task summaries, preferences. Human can view/edit/forget entries. |
| **Red Line Rule** | A per-Agent policy: actions matching a rule require human approval before execution. |

## 7. Features (v1 scope)

### 7.1 Agent Management
- Create/edit/delete an Agent: name, avatar/emoji, role (free text, no fixed template requirement), persona text, assigned model binding, enabled tools.
- Enable/disable an Agent (pause/resume its always-on loop).

### 7.2 Persona System
- Persona is a structured object, not just a blob of text:
  - `identity`: name, role title, one-line bio.
  - `tone`: style/voice instructions.
  - `constraints`: hard do/don't rules (e.g. "never push to main", "always ask before deleting").
  - `knowledge`: static context always included (e.g. "our stack is Go + Next.js + Postgres").
- Persona compiles into a system prompt prefix at runtime; editing persona takes effect on the next task, no redeploy needed.

### 7.3 Bring Your Own Model (BYOM)
- Model provider abstraction with a common interface: `send(messages, tools, systemPrompt) -> response`.
- v1 providers: **Anthropic (Claude API)** as the primary/reference implementation, plus a generic **OpenAI-compatible** adapter (covers OpenAI, and self-hosted OpenAI-compatible endpoints like Ollama/vLLM) — validates the abstraction isn't Claude-specific.
- API keys stored encrypted at rest (see §9), configurable per-Agent (different Agents can use different models/providers).
- Model config includes: provider, model name, max tokens, temperature, and an optional "thinking/reasoning effort" pass-through if the provider supports it.

### 7.4 Tasks & Triggers
- Manual: chat with an Agent in the console; message becomes a task.
- Scheduled: cron-style recurring task definition per Agent (e.g. "every morning, summarize overnight alerts").
- Event/webhook: register an inbound webhook URL per Agent; payload becomes task input.
- API: REST endpoint to create a task for a given Agent programmatically.
- Task lifecycle: `queued -> running -> (waiting_approval) -> done | failed`.
- Each task stores full transcript (messages, tool calls, tool results) for audit/debugging.

### 7.5 Memory
- Per-Agent memory store: key facts/preferences + rolling summaries of past tasks.
- Memory is retrieved and injected into context at the start of a task (simple relevance/recency selection in v1 — no vector DB required initially; can be added later if needed).
- Console UI to view, edit, or delete individual memory entries.

### 7.6 Approval / Red Line Rules
- Per-Agent list of rules (simple pattern/keyword or tool-name match, e.g. "any call to `deploy_tool`" or "any message containing 'DELETE'").
- When a task's next action matches a rule, task status becomes `waiting_approval`; human approves/rejects from the console before execution continues.

### 7.7 Tools
- Pluggable tool interface (function-calling schema) an Agent can be granted access to. v1 ships a small built-in set (HTTP request, shell command in a sandboxed/whitelisted way, file read/write in a scoped workspace) — extensible later.

### 7.8 Web Console (Next.js)
- Agent list/management screens.
- Chat view per Agent (conversation + live task status).
- Task board: list and kanban/swimlane views by status.
- Memory viewer/editor per Agent.
- Model binding + API key management screen.
- Approval queue view (tasks waiting on a human decision).

## 8. Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Backend / orchestrator | **Go** | Concurrency primitives fit an always-on task loop per Agent well (goroutines + channels); single static binary simplifies self-hosting. |
| Frontend / console | **Next.js** (TypeScript) | Console needs routing, forms, and a chat UI — Next.js gives SSR/CSR flexibility and a large component ecosystem. |
| API between them | REST + Server-Sent Events (SSE) for streaming chat/task output | Simpler to self-host than WebSockets at v1; upgrade path to WS later if needed. |
| Database | **Postgres** | Tasks, Agents, memory, transcripts — all relational with some JSONB fields (persona, tool schemas). |
| Task queue / scheduler | In-process Go scheduler (cron-style) + a durable `tasks` table as the queue (poll-based) | Avoids adding Redis/RabbitMQ as a hard dependency for v1; revisit if throughput demands it. |
| Model providers | Anthropic Claude API (primary), OpenAI-compatible adapter | Directly satisfies BYOM goal; OpenAI-compatible covers most self-hosted/local model servers too. |
| Secrets | Encrypted columns (e.g. AES-GCM with a key from env/KMS) for API keys | Self-hosted — no managed secrets service assumed. |
| Deployment | Docker Compose (Go binary + Next.js + Postgres) | One-command self-host, matches "self-hosted" goal. |

## 9. Data Model (initial sketch)

- `agents`: id, name, role, persona (jsonb), model_binding_id, enabled, created_at.
- `model_bindings`: id, provider, model_name, api_key (encrypted), config (jsonb).
- `tasks`: id, agent_id, trigger_type, status, input, transcript (jsonb), created_at, updated_at.
- `memory_entries`: id, agent_id, type (fact/summary/preference), content, created_at, updated_at.
- `red_line_rules`: id, agent_id, match_type, pattern, created_at.
- `schedules`: id, agent_id, cron_expr, task_template (jsonb), enabled.
- `webhooks`: id, agent_id, token, created_at.

## 10. Success Criteria (v1)

- Can create an Agent, attach a persona, bind it to a Claude API key, and have it complete a manually-triggered chat task end-to-end.
- Can define a cron-scheduled task and see it execute unattended.
- Can define a red-line rule and see a task correctly pause for approval, then resume after approval.
- Memory entries persist across separate tasks and are visibly used (e.g. Agent recalls a stated preference in a later, unrelated session).
- Entire stack runs via `docker compose up` with no external SaaS dependency other than the model provider API.

## 11. Open Questions

- Single-user auth only for v1, or basic multi-user from the start? (Affects `agents` ownership model.)
- Memory retrieval strategy: is naive recency-based injection good enough for v1, or is a vector store needed sooner than expected?
- Tool sandboxing: how strict should shell/file tool access be by default (allowlist vs. full access with red-line gating)?
- Streaming: is SSE sufficient for the chat UI, or does task-board "live update" need something bidirectional (WebSocket) sooner?

## 12. Phasing

- **Phase 1 (MVP):** Agent CRUD, persona, Claude API binding, manual chat tasks, task transcript storage, basic console (no scheduling/webhooks/approval yet).
- **Phase 2:** Scheduling, webhooks, API-triggered tasks, task board with statuses.
- **Phase 3:** Memory system (view/edit/delete), red-line approval flow.
- **Phase 4:** Additional model providers (OpenAI-compatible), tool plugin system, multi-user basics.
