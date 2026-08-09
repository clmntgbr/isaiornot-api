-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    noise DOUBLE PRECISION NOT NULL,
    compression DOUBLE PRECISION NOT NULL,
    frequency DOUBLE PRECISION NOT NULL,
    histogram DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE medias
    ADD COLUMN IF NOT EXISTS insight_id UUID REFERENCES insights (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_media_insight_id ON medias (insight_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_insight_id;
ALTER TABLE medias DROP COLUMN IF EXISTS insight_id;
DROP TABLE IF EXISTS insights;
-- +goose StatementEnd
