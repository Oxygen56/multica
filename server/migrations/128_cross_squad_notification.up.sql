-- 128: Cross-Squad Collaboration Notification System
-- Implements Feature #4596 — dispatch contracts, notification records,
-- and anomaly detection infrastructure.

-- ============================================================================
-- A. dispatch_contracts — formal contracts between two issues
-- ============================================================================
CREATE TABLE dispatch_contracts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    from_issue_id   UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    to_issue_id     UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    trigger_events  TEXT[] NOT NULL DEFAULT '{}',   -- event types that trigger fulfillment
    target_issue    UUID REFERENCES issues (id),     -- optional: issue the contract serves
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'triggered', 'fulfilled', 'cancelled')),
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    fulfilled_at    TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ
);

CREATE INDEX dispatch_contracts_from_idx ON dispatch_contracts (from_issue_id, status);
CREATE INDEX dispatch_contracts_to_idx ON dispatch_contracts (to_issue_id, status);
CREATE INDEX dispatch_contracts_workspace_idx ON dispatch_contracts (workspace_id);
CREATE INDEX dispatch_contracts_pending_target_idx
    ON dispatch_contracts (target_issue, status) WHERE status = 'pending';

-- ============================================================================
-- B. notification_records — cross-issue notification delivery tracking
-- ============================================================================
CREATE TABLE notification_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    notification_type TEXT NOT NULL,                 -- e.g. 'child_terminal', 'review_completed', 'contract_fulfilled'
    source_issue_id UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    target_issue_id UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    rule_id         TEXT,                            -- e.g. 'R1', 'R2', 'R5' — which rule triggered this
    status          TEXT NOT NULL DEFAULT 'delivered'
                    CHECK (status IN ('delivered', 'acknowledged', 'suppressed', 'failed')),
    details         JSONB NOT NULL DEFAULT '{}',
    acknowledged_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notification_records_target_idx
    ON notification_records (target_issue_id, created_at DESC);
CREATE INDEX notification_records_source_idx
    ON notification_records (source_issue_id, created_at DESC);
CREATE INDEX notification_records_workspace_type_idx
    ON notification_records (workspace_id, notification_type);

-- ============================================================================
-- C. event_log — structured event persistence for reliability (Module E)
-- ============================================================================
CREATE TABLE event_log (
    id              BIGSERIAL PRIMARY KEY,
    workspace_id    UUID NOT NULL,
    event_type      TEXT NOT NULL,
    event_data      JSONB NOT NULL DEFAULT '{}',
    sequence_num    BIGINT NOT NULL,                 -- per-workspace monotonic sequence
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, sequence_num)
);

CREATE INDEX event_log_workspace_seq_idx ON event_log (workspace_id, sequence_num DESC);
CREATE INDEX event_log_type_created_idx ON event_log (event_type, created_at DESC);

-- ============================================================================
-- D. notify_subscribers_v2: enhanced with rule-driven routing metadata
-- ============================================================================
-- Add metadata columns to inbox_items for dual-layer status semantics (Module C).
-- interaction_posture, review_target etc. are stored on the issue as metadata KV.
-- The inbox_items.actor_type is extended to support 'system_rule' for rule-driven
-- notifications so the UI can distinguish user actions from automated rule fires.
-- No schema change needed — actor_type is TEXT; we just document the new value.
COMMENT ON COLUMN inbox_items.actor_type IS
    'actor type: member, agent, system, or system_rule (for rule-driven notifications)';
