CREATE TABLE IF NOT EXISTS e2e_identity_keys (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    identity_key  BYTEA NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT e2e_identity_key_len CHECK (octet_length(identity_key) = 32)
);

CREATE TABLE IF NOT EXISTS e2e_signed_prekeys (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    key_id     INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    signature  BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT e2e_signed_prekey_len CHECK (octet_length(public_key) = 32)
);

CREATE TABLE IF NOT EXISTS e2e_one_time_prekeys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_id     INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, key_id),
    CONSTRAINT e2e_one_time_prekey_len CHECK (octet_length(public_key) = 32)
);

CREATE INDEX IF NOT EXISTS e2e_one_time_prekeys_user_unused_idx
    ON e2e_one_time_prekeys (user_id)
    WHERE used_at IS NULL;

CREATE TABLE IF NOT EXISTS e2e_messages (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id            TEXT NOT NULL,
    ratchet_session_id TEXT NOT NULL,
    sender_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ciphertext         BYTEA NOT NULL,
    ratchet_header     BYTEA NOT NULL,
    sent_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS e2e_messages_recipient_sent_at_idx
    ON e2e_messages (recipient_id, sent_at, id);
