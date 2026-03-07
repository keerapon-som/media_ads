CREATE TABLE IF NOT EXISTS media_archives (
    object_id    TEXT PRIMARY KEY,
    key          TEXT NOT NULL,
    filename     TEXT NOT NULL,
    extension    TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL CHECK (size_bytes >= 0),
    content_type TEXT NOT NULL,
    probe_data   JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_media_archives_key ON media_archives (key);