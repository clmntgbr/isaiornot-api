-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
    message TEXT NOT NULL DEFAULT '',
    final_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    confidence TEXT NOT NULL DEFAULT '',
    verdict TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scans_user_id ON scans (user_id);
CREATE INDEX IF NOT EXISTS idx_scans_status ON scans (status);
CREATE INDEX IF NOT EXISTS idx_scans_created_at ON scans (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scans;
-- +goose StatementEnd
