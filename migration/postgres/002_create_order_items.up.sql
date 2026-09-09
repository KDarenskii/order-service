-- TODO: Создайте таблицу order_items и индекс:
-- 1. id — BIGSERIAL, NOT NULL, UNIQUE.
-- 2. guid — UUID, NOT NULL, PRIMARY KEY.
-- 3. order_guid — UUID, NOT NULL, FK на orders(guid) с ON DELETE CASCADE.
-- 4. product_guid — UUID, NOT NULL (ссылка на товар из catalog-service).
-- 5. quantity — INTEGER, NOT NULL.
-- 6. unit_price — BIGINT, NOT NULL (цена за единицу на момент покупки).
-- 7. created_at — TIMESTAMPTZ, NOT NULL, DEFAULT NOW().
-- 8. updated_at — TIMESTAMPTZ, NOT NULL, DEFAULT NOW().
--
-- После CREATE TABLE создайте индекс:
-- idx_order_items_order_guid на колонку order_guid.

CREATE TABLE order_items (
     id           BIGSERIAL NOT NULL UNIQUE,
     guid         UUID NOT NULL PRIMARY KEY,
     order_guid   UUID NOT NULL REFERENCES orders(guid) ON DELETE CASCADE,
     product_guid UUID NOT NULL,
     quantity     INTEGER NOT NULL,
     unit_price   BIGINT NOT NULL,
     created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_guid ON order_items (order_guid);