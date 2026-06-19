CREATE TABLE IF NOT EXISTS streams (
    id TEXT PRIMARY KEY,
    stream_key TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (
        status IN ('created', 'running', 'stopped', 'failed')
    ),
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    stopped_at TIMESTAMPTZ
);