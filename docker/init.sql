-- Create databases
CREATE DATABASE orders_db;
CREATE DATABASE payments_db;
CREATE DATABASE notifications_db;

\connect orders_db;

CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36)  PRIMARY KEY,
    customer_id     VARCHAR(36)  NOT NULL,
    item_name       VARCHAR(255) NOT NULL,
    amount          BIGINT       NOT NULL CHECK (amount > 0),
    status          VARCHAR(20)  NOT NULL DEFAULT 'Pending',
    created_at      TIMESTAMPTZ  NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_orders_idempotency_key
    ON orders (idempotency_key) WHERE idempotency_key IS NOT NULL;

\connect payments_db;

CREATE TABLE IF NOT EXISTS payments (
    id             VARCHAR(36)  PRIMARY KEY,
    order_id       VARCHAR(36)  NOT NULL UNIQUE,
    transaction_id VARCHAR(36)  NOT NULL,
    amount         BIGINT       NOT NULL CHECK (amount > 0),
    status         VARCHAR(20)  NOT NULL,
    customer_email VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);

\connect notifications_db;

CREATE TABLE IF NOT EXISTS processed_events (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
