-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    slug VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    billing_interval VARCHAR(255) NOT NULL DEFAULT 'month',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(255) NOT NULL DEFAULT 'EUR',
    quota_id UUID NOT NULL REFERENCES quotas (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plan_quota_id ON plans (quota_id);
CREATE INDEX IF NOT EXISTS idx_plan_slug ON plans (slug);
CREATE INDEX IF NOT EXISTS idx_plan_stripe_price_id ON plans (stripe_price_id);
CREATE INDEX IF NOT EXISTS idx_plan_is_active ON plans (is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_plan_is_active;
DROP INDEX IF EXISTS idx_plan_stripe_price_id;
DROP INDEX IF EXISTS idx_plan_slug;
DROP INDEX IF EXISTS idx_plan_quota_id;
DROP TABLE IF EXISTS plans;
-- +goose StatementEnd
