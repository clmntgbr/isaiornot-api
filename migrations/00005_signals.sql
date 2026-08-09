-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_id UUID NOT NULL REFERENCES medias (id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    score INTEGER NOT NULL DEFAULT 0,
    confidence TEXT NOT NULL DEFAULT 'unknown',
    details JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT signals_confidence_check CHECK (confidence IN ('high', 'medium', 'low', 'unknown'))
);

CREATE INDEX IF NOT EXISTS idx_signals_media_id ON signals (media_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_signals_media_id_name ON signals (media_id, name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS signals;
-- +goose StatementEnd
