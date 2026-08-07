CREATE TABLE IF NOT EXISTS model_bindings (
    id UUID PRIMARY KEY,
    provider TEXT NOT NULL,
    model_name TEXT NOT NULL,
    api_key_encrypted BYTEA NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wakers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT '',
    persona JSONB NOT NULL DEFAULT '{}',
    model_binding_id UUID REFERENCES model_bindings(id) ON DELETE SET NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    waker_id UUID NOT NULL REFERENCES wakers(id) ON DELETE CASCADE,
    trigger_type TEXT NOT NULL DEFAULT 'manual',
    status TEXT NOT NULL DEFAULT 'queued',
    input TEXT NOT NULL DEFAULT '',
    transcript JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS memory_entries (
    id UUID PRIMARY KEY,
    waker_id UUID NOT NULL REFERENCES wakers(id) ON DELETE CASCADE,
    entry_type TEXT NOT NULL DEFAULT 'fact',
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_waker_id ON tasks(waker_id);
CREATE INDEX IF NOT EXISTS idx_memory_entries_waker_id ON memory_entries(waker_id);
