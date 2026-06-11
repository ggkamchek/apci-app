CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Known chat user IDs (external to this service).
CREATE TABLE IF NOT EXISTS account_nodes (
    user_id       UUID PRIMARY KEY,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_devices (
    user_id       UUID NOT NULL REFERENCES account_nodes(user_id) ON DELETE CASCADE,
    device_hash   TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, device_hash)
);

CREATE INDEX IF NOT EXISTS user_devices_device_hash_idx ON user_devices(device_hash);

CREATE TABLE IF NOT EXISTS user_ips (
    user_id       UUID NOT NULL REFERENCES account_nodes(user_id) ON DELETE CASCADE,
    ip_hash       TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, ip_hash)
);

CREATE INDEX IF NOT EXISTS user_ips_ip_hash_idx ON user_ips(ip_hash);

-- Undirected link: account_a < account_b (lexicographic UUID order).
CREATE TABLE IF NOT EXISTS account_links (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_a    UUID NOT NULL,
    account_b    UUID NOT NULL,
    link_type    TEXT NOT NULL,
    weight       INT NOT NULL,
    detected_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT account_links_ordered CHECK (account_a::text < account_b::text),
    UNIQUE (account_a, account_b, link_type)
);

CREATE INDEX IF NOT EXISTS account_links_account_a_idx ON account_links(account_a);
CREATE INDEX IF NOT EXISTS account_links_account_b_idx ON account_links(account_b);

CREATE TABLE IF NOT EXISTS ingested_events (
    event_id    UUID PRIMARY KEY,
    event_type  TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
