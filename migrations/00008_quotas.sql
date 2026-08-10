-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    max_images_per_month INTEGER NOT NULL,
    max_videos_per_month INTEGER NOT NULL,
    max_file_size_image BIGINT NOT NULL,
    max_file_size_video BIGINT NOT NULL,
    full_pipeline BOOLEAN NOT NULL DEFAULT FALSE,
    history_retention BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS quotas;
-- +goose StatementEnd
