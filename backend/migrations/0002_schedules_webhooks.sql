CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY,
    waker_id UUID NOT NULL REFERENCES wakers(id) ON DELETE CASCADE,
    cron_expr TEXT NOT NULL,
    input TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY,
    waker_id UUID NOT NULL REFERENCES wakers(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_schedules_waker_id ON schedules(waker_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_waker_id ON webhooks(waker_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_token ON webhooks(token);
