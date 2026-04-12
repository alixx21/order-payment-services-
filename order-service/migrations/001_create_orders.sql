CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36)  PRIMARY KEY,
    customer_id     VARCHAR(36)  NOT NULL,
    item_name       VARCHAR(255) NOT NULL,
    amount          BIGINT       NOT NULL CHECK (amount > 0),
    status          VARCHAR(20)  NOT NULL DEFAULT 'Pending',
    created_at      TIMESTAMPTZ  NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_orders_idempotency_key ON orders (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
