CREATE TABLE IF NOT EXISTS media (
    media_id    TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    description TEXT,
    object_id    TEXT NOT NULL,
    content_type TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_object_id ON media (object_id);

CREATE TABLE IF NOT EXISTS media_publish_register (
    publisher_id   SERIAL PRIMARY KEY,
    media_id       TEXT NOT NULL,
    publish_from   TIMESTAMPTZ NOT NULL,
    publish_to     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (media_id) REFERENCES media(media_id) ON DELETE CASCADE,
    CHECK (publish_to IS NULL OR publish_to > publish_from)
);

CREATE INDEX IF NOT EXISTS idx_media_publish_register_media_id ON media_publish_register (media_id);


CREATE TABLE IF NOT EXISTS media_owner (
    media_id    TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (media_id) REFERENCES media(media_id) ON DELETE CASCADE
);
