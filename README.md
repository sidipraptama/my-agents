# my-agents

Self-hosted, always-on AI agents ("Agents") with a persona, a role, and your own model (BYOM). See [PRD.md](PRD.md) for the full spec.

## Run with Docker Compose

```bash
cp .env.example .env
# edit .env: set ENCRYPTION_KEY to a random 32-character string
docker-compose up --build
```

> Note: use `docker-compose` (hyphenated). If your machine has the newer `docker compose` CLI plugin instead, that works too — check with `docker compose version`.

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

Then in the UI: go to **Model Bindings**, add your Claude API key, then create a **Agent** and attach it.

## Run locally without Docker

**Backend** (needs Postgres running locally):

```bash
cd backend
export DATABASE_URL="postgres://myagents:myagents@localhost:5432/myagents?sslmode=disable"
export ENCRYPTION_KEY="change-this-to-a-random-32-char-key"
go run ./cmd/server
```

**Frontend**:

```bash
cd frontend
npm install
npm run dev
```

## Status

Phase 1 (MVP) per the PRD: Agent CRUD, structured persona, Claude API model binding, manual chat with persisted task transcripts. Scheduling, webhooks, memory, and approval gating are not yet implemented (Phases 2-3).
