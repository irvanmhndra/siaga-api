CREATE TABLE schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(150) NOT NULL,
    timezone        VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_schedules_org ON schedules(organization_id);

CREATE TABLE schedule_layers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id          UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    name                 VARCHAR(150) NOT NULL DEFAULT 'Layer 1',
    rotation_type        VARCHAR(20) NOT NULL DEFAULT 'weekly',
    rotation_length_days INT NOT NULL DEFAULT 7,
    handoff_time         TIME NOT NULL DEFAULT '09:00',
    start_at             TIMESTAMPTZ NOT NULL,
    user_ids             UUID[] NOT NULL DEFAULT '{}',
    position             INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_schedule_layers_schedule ON schedule_layers(schedule_id);

CREATE TABLE schedule_overrides (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_schedule_overrides ON schedule_overrides(schedule_id, start_at, end_at);
