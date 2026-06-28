-- 128_notification_bus: Cross-squad collaboration notification and
-- coordination infrastructure (Feature #4596).
--
-- Adds two tables:
--   1. dispatch_contracts — tracks cross-squad dispatch lifecycle
--   2. notification_records — persistent log of notification deliveries

-- ── dispatch_contracts ──────────────────────────────────────────────────────
--
-- A dispatch contract is created alongside a cross-squad sub-issue and binds
-- the two issues together with a callback contract. When the dispatched
-- sub-issue reaches a terminal status, the notification bus evaluates the
-- contract and executes the prescribed callback (comment on target, mention
-- assignee, etc.).
CREATE TABLE dispatch_contracts (
    contract_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    creator_agent_id  UUID NOT NULL,
    from_issue_id     UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    to_issue_id       UUID UNIQUE REFERENCES issues(id) ON DELETE SET NULL,
    from_squad_id     UUID,
    to_squad_id       UUID,
    trigger_events    JSONB NOT NULL DEFAULT '["child_terminal"]',
    target_issue      UUID,
    notify_assignee   BOOLEAN NOT NULL DEFAULT true,
    notify_creator    BOOLEAN NOT NULL DEFAULT false,
    include_summary   BOOLEAN NOT NULL DEFAULT true,
    status            TEXT NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'triggered', 'fulfilled', 'cancelled')),
    fulfilled_at      TIMESTAMPTZ
);

CREATE INDEX dispatch_contracts_from_issue_idx ON dispatch_contracts (from_issue_id);
CREATE INDEX dispatch_contracts_to_issue_idx ON dispatch_contracts (to_issue_id)
    WHERE to_issue_id IS NOT NULL;
CREATE INDEX dispatch_contracts_status_idx ON dispatch_contracts (status)
    WHERE status = 'pending';
CREATE INDEX dispatch_contracts_from_squad_idx ON dispatch_contracts (from_squad_id)
    WHERE from_squad_id IS NOT NULL;

-- ── notification_records ────────────────────────────────────────────────────
--
-- Persistent audit log of every notification delivery. Serves three purposes:
--   1. Idempotency guard: skip duplicate notifications for the same target /
--      source / type combination within a brief window.
--   2. Pull-mode query: agents can query "what notifications are pending for me"
--      via the CLI (`multica notification pending`).
--   3. Diagnostics: operators can trace notification delivery chains across
--      issues for debugging deadlocks.
CREATE TABLE notification_records (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    target_type      TEXT NOT NULL CHECK (target_type IN ('agent', 'squad', 'member')),
    target_id        UUID NOT NULL,
    source_issue_id  UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    type             TEXT NOT NULL CHECK (type IN (
                         'review_completed', 'child_terminal', 'unblocked',
                         'stage_completed', 'deadlock_detected',
                         'contract_fulfilled', 'contract_timeout',
                         'stale_review', 'orphan_notification',
                         'cross_squad_silence', 'circuit_broken'
                     )),
    message          TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'acknowledged')),
    acknowledged_at  TIMESTAMPTZ
);

CREATE INDEX notification_records_target_idx
    ON notification_records (target_type, target_id, status);
CREATE INDEX notification_records_source_idx
    ON notification_records (source_issue_id);
CREATE INDEX notification_records_type_idx
    ON notification_records (type, created_at);
