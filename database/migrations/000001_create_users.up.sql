CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT NOT NULL,
    username        CITEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    display_name    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_length CHECK (char_length(email::text) BETWEEN 3 AND 254),
    CONSTRAINT users_username_length CHECK (char_length(username::text) BETWEEN 3 AND 32)
);

CREATE INDEX users_created_at_idx ON users (created_at);
