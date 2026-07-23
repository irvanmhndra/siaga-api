CREATE TABLE services (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name                 VARCHAR(150) NOT NULL,
    description          TEXT,
    escalation_policy_id UUID REFERENCES escalation_policies(id) ON DELETE SET NULL,
    integration_key      VARCHAR(64) NOT NULL UNIQUE,
    status               VARCHAR(20) NOT NULL DEFAULT 'active',
    auto_resolve_minutes INT,
    ack_timeout_minutes  INT,
    slack_webhook_url    TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_services_org ON services(organization_id);
