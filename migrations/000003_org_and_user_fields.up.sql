CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE users ADD COLUMN phone        VARCHAR(30);
ALTER TABLE users ADD COLUMN timezone     VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta';
ALTER TABLE users ADD COLUMN notify_email BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN notify_sms   BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_users_org ON users(organization_id);
