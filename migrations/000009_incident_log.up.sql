CREATE TABLE incident_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    type        VARCHAR(30) NOT NULL,
    message     TEXT NOT NULL,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_incident_log_incident ON incident_log(incident_id, created_at);
