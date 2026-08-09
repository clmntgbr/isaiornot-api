-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS medias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scans (id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    filename TEXT NOT NULL DEFAULT '',
    thumbnail TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    status TEXT NOT NULL,
    statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT medias_key_key UNIQUE (key),
    CONSTRAINT medias_status_check CHECK (status IN ('processing', 'uploaded', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_media_scan_id ON medias (scan_id);
CREATE INDEX IF NOT EXISTS idx_media_status ON medias (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS medias;
-- +goose StatementEnd
