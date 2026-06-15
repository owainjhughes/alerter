CREATE TABLE IF NOT EXISTS monitors (
    id               uuid PRIMARY KEY,
    name             text NOT NULL,
    type             text NOT NULL,
    target           text NOT NULL,
    interval_seconds integer NOT NULL,
    timeout_seconds  integer NOT NULL,
    expected_status  integer NOT NULL DEFAULT 0,
    enabled          boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL
);
