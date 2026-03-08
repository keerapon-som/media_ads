CREATE TABLE IF NOT EXISTS object_library (
    object_id    TEXT PRIMARY KEY,
    key          TEXT NOT NULL,
    filename     TEXT NOT NULL,
    extension    TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL CHECK (size_bytes >= 0),
    content_type TEXT NOT NULL,
    probe_data   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);