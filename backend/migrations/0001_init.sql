CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name TEXT NOT NULL,
    image TEXT NOT NULL,
    namespace TEXT NOT NULL,
    replicas INTEGER NOT NULL CHECK (replicas > 0),
    cpu_limit TEXT,
    memory_limit TEXT,
    privileged BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS policy_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    control_no TEXT NOT NULL,
    message TEXT NOT NULL,
    severity TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deployments_created_at
    ON deployments (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_policy_violations_deployment_id
    ON policy_violations (deployment_id);
