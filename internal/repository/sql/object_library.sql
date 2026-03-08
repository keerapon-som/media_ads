CREATE TABLE IF NOT EXISTS object_library (
    object_id    TEXT PRIMARY KEY,
    key          TEXT NOT NULL,
    filename     TEXT NOT NULL,
    extension    TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL CHECK (size_bytes >= 0),
    content_type TEXT NOT NULL,
    probe_data   JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_object_library_key ON object_library (key);


CREATE TABLE IF NOT EXISTS upload_slot (
    upload_id   TEXT PRIMARY KEY,
    status       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);