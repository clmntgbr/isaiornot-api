-- +goose Up
-- +goose StatementBegin
ALTER TABLE scans
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'image';

ALTER TABLE scans
    ADD CONSTRAINT scans_type_check CHECK (type IN ('image', 'video'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE scans DROP CONSTRAINT IF EXISTS scans_type_check;
ALTER TABLE scans DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
