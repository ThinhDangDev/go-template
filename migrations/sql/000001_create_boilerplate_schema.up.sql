CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(64) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL DEFAULT '',
    v0 VARCHAR(100) NOT NULL DEFAULT '',
    v1 VARCHAR(100) NOT NULL DEFAULT '',
    v2 VARCHAR(100) NOT NULL DEFAULT '',
    v3 VARCHAR(100) NOT NULL DEFAULT '',
    v4 VARCHAR(100) NOT NULL DEFAULT '',
    v5 VARCHAR(100) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_casbin_rule_unique
    ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

INSERT INTO casbin_rule (ptype, v0, v1, v2)
VALUES
    ('p', 'admin', '/api/v1/*', '(GET|POST|PUT|PATCH|DELETE)'),
    ('p', 'operator', '/api/v1/auth/me', 'GET'),
    ('p', 'operator', '/api/v1/operator/*', 'GET'),
    ('p', 'viewer', '/api/v1/auth/me', 'GET'),
    ('p', 'viewer', '/api/v1/viewer/*', 'GET')
ON CONFLICT DO NOTHING;
