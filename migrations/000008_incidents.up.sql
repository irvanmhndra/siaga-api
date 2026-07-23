CREATE TABLE incidents (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    service_id           UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    title                TEXT NOT NULL,
    status               VARCHAR(20) NOT NULL DEFAULT 'triggered',
    urgency              VARCHAR(10) NOT NULL DEFAULT 'high',
    dedup_key            VARCHAR(255) NOT NULL,
    details              JSONB,
    assigned_user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    escalation_policy_id UUID REFERENCES escalation_policies(id) ON DELETE SET NULL,
    escalation_level     INT NOT NULL DEFAULT 0,
    escalation_repeats   INT NOT NULL DEFAULT 0,
    next_escalation_at   TIMESTAMPTZ,
    triggered_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at      TIMESTAMPTZ,
    resolved_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Dedup: at most one OPEN incident per (service, dedup_key).
CREATE UNIQUE INDEX uniq_open_incident ON incidents(service_id, dedup_key) WHERE status <> 'resolved';
CREATE INDEX idx_incidents_org_status ON incidents(organization_id, status);
CREATE INDEX idx_incidents_next_escalation ON incidents(next_escalation_at) WHERE status = 'triggered';
