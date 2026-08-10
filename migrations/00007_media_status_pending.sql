-- +goose Up
-- +goose StatementBegin
ALTER TABLE medias DROP CONSTRAINT IF EXISTS medias_status_check;
ALTER TABLE medias
    ADD CONSTRAINT medias_status_check
    CHECK (status IN ('pending', 'uploaded', 'processing', 'completed', 'failed'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE medias SET status = 'processing' WHERE status = 'pending';
ALTER TABLE medias DROP CONSTRAINT IF EXISTS medias_status_check;
ALTER TABLE medias
    ADD CONSTRAINT medias_status_check
    CHECK (status IN ('processing', 'uploaded', 'completed', 'failed'));
-- +goose StatementEnd
