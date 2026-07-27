CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO quotas (
    id, max_images_per_month, max_videos_per_month,
    max_file_size_image, max_file_size_video, full_pipeline, history_retention,
    created_at, updated_at
)
VALUES
('e75e7940-2683-4ea1-a5c7-6d1724892369', 10, 0, 5242880, 0, false, 604800000000000, now(), now()),
('e437e948-9ac7-422a-8a86-79ca43ca9d5d', 100, 0, 26214400, 0, true, 2592000000000000, now(), now()),
('aca09dc8-1c60-4670-9e2a-2eac0a8dc268', 100, 0, 26214400, 0, true, 2592000000000000, now(), now()),
('dcad50c5-7a2d-4de6-958f-e9d7827fab4d', 1000, 50, 209715200, 524288000, true, 7776000000000000, now(), now()),
('33c61c6e-d2c5-4915-9b5a-3fa8287047b8', 1000, 50, 209715200, 524288000, true, 7776000000000000, now(), now()),
('b8f650d8-e625-4604-ac34-033de9f5246a', 10000, 500, 1073741824, 1073741824, true, 31536000000000000, now(), now()),
('c90170c4-3824-4190-b81b-207db611442b', 10000, 500, 1073741824, 1073741824, true, 31536000000000000, now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO plans (
    id,
    slug,
    name,
    description,
    billing_interval,
    price,
    currency,
    stripe_price_id,
    is_active,
    quota_id,
    created_at,
    updated_at
)
VALUES
(gen_random_uuid(),'free','Free','Découverte du produit, scans limitées','monthly',0,'eur','price_1TwdhH5vcjninxNhcRf8qkUH',true,'e75e7940-2683-4ea1-a5c7-6d1724892369',now(),now()),
(gen_random_uuid(),'starter','Starter','Pour un usage régulier, image uniquement','monthly',1200,'eur','price_1Twdhe5vcjninxNhbAVuBcEG',true,'e437e948-9ac7-422a-8a86-79ca43ca9d5d',now(),now()),
(gen_random_uuid(),'starter','Starter','Pour un usage régulier, image uniquement','annually',11500,'eur','price_1TwdiA5vcjninxNhk4J7n7Tt',true,'aca09dc8-1c60-4670-9e2a-2eac0a8dc268',now(),now()),
(gen_random_uuid(),'pro','Pro','Pipeline complet, images et vidéos','monthly',3900,'eur','price_1TwdjN5vcjninxNh3vLqLC6W',true,'dcad50c5-7a2d-4de6-958f-e9d7827fab4d',now(),now()),
(gen_random_uuid(),'pro','Pro','Pipeline complet, images et vidéos','annually',37500,'eur','price_1Twdjd5vcjninxNhS2kRx2WH',true,'33c61c6e-d2c5-4915-9b5a-3fa8287047b8',now(),now()),
(gen_random_uuid(),'business','Business','Volume, support dédié, détail par sous-signal','monthly',14900,'eur','price_1Twdjz5vcjninxNhLbKezUvo',true,'b8f650d8-e625-4604-ac34-033de9f5246a',now(),now()),
(gen_random_uuid(),'business','Business','Volume, support dédié, détail par sous-signal','annually',143000,'eur','price_1TwdkH5vcjninxNhCY6mZyvV',true,'c90170c4-3824-4190-b81b-207db611442b',now(),now())
ON CONFLICT (id) DO NOTHING;
