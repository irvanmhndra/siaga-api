CREATE TABLE escalation_policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(150) NOT NULL,
    repeat_count    INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_escalation_policies_org ON escalation_policies(organization_id);

CREATE TABLE escalation_rules (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    escalation_policy_id UUID NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,
    position             INT NOT NULL,
    delay_minutes        INT NOT NULL DEFAULT 30,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_escalation_rules_policy ON escalation_rules(escalation_policy_id, position);

CREATE TABLE escalation_targets (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    escalation_rule_id UUID NOT NULL REFERENCES escalation_rules(id) ON DELETE CASCADE,
    target_type        VARCHAR(20) NOT NULL,
    target_id          UUID NOT NULL,
    position           INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_escalation_targets_rule ON escalation_targets(escalation_rule_id);
