package notifications

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// EventPersister provides structured event persistence for reliability (Module E).
// It records critical events to the event_log table with per-workspace monotonic
// sequence numbers. Sequence gaps are detectable by consumers: if the consumer's
// last-seen sequence is N-2 but the current event is N, a gap exists at N-1.
//
// Events are written asynchronously to avoid blocking the main event processing
// path. A bounded channel buffers writes; when full, events are dropped (logged
// at warn) rather than blocking the bus.
type EventPersister struct {
	queries  *db.Queries
	ch       chan persistEntry
	seqCache sync.Map // workspace_id → *int64
	done     chan struct{}
}

type persistEntry struct {
	WorkspaceID string
	EventType   string
	Payload     any
}

// NewEventPersister creates a new event persister. It starts a background
// goroutine that writes events to the event_log table.
func NewEventPersister(queries *db.Queries) *EventPersister {
	ep := &EventPersister{
		queries: queries,
		ch:      make(chan persistEntry, 1024),
		done:    make(chan struct{}),
	}
	go ep.worker()
	return ep
}

// Record enqueues an event for persistence. Non-blocking: if the buffer is
// full, the event is dropped and logged at warn level.
func (ep *EventPersister) Record(ctx context.Context, workspaceID, eventType string, payload any) {
	entry := persistEntry{
		WorkspaceID: workspaceID,
		EventType:   eventType,
		Payload:     payload,
	}
	select {
	case ep.ch <- entry:
	default:
		slog.Warn("event persister buffer full, dropping event",
			"event_type", eventType,
			"workspace_id", workspaceID)
	}
}

// Shutdown gracefully stops the background worker.
func (ep *EventPersister) Shutdown() {
	close(ep.done)
}

// worker reads from the channel and writes to the database.
func (ep *EventPersister) worker() {
	for {
		select {
		case <-ep.done:
			// Drain remaining entries before exiting.
			for {
				select {
				case entry := <-ep.ch:
					ep.write(context.Background(), entry)
				default:
					return
				}
			}
		case entry := <-ep.ch:
			ep.write(context.Background(), entry)
		}
	}
}

// write persists a single event to the database with a monotonic sequence number.
func (ep *EventPersister) write(ctx context.Context, entry persistEntry) {
	seq := ep.nextSeq(entry.WorkspaceID)
	payloadBytes, err := json.Marshal(entry.Payload)
	if err != nil {
		slog.Warn("event persister: failed to marshal payload",
			"event_type", entry.EventType, "error", err)
		return
	}

	if err := ep.queries.InsertEventLog(ctx, db.InsertEventLogParams{
		WorkspaceID: util.MustParseUUID(entry.WorkspaceID),
		EventType:   entry.EventType,
		EventData:   payloadBytes,
		SequenceNum: seq,
	}); err != nil {
		slog.Warn("event persister: failed to insert event log",
			"event_type", entry.EventType,
			"workspace_id", entry.WorkspaceID,
			"sequence", seq,
			"error", err)
		return
	}
}

// nextSeq returns the next monotonic sequence number for a workspace.
// Sequence numbers start at 1 and are tracked in-memory. On restart,
// the sequence resets — consumers detect gaps via the low-water mark.
func (ep *EventPersister) nextSeq(workspaceID string) int64 {
	raw, _ := ep.seqCache.LoadOrStore(workspaceID, new(int64))
	ptr := raw.(*int64)
	return atomic.AddInt64(ptr, 1)
}

// LastSeq returns the last sequence number written for a workspace.
// Returns 0 if no events have been written.
func (ep *EventPersister) LastSeq(workspaceID string) int64 {
	raw, ok := ep.seqCache.Load(workspaceID)
	if !ok {
		return 0
	}
	return atomic.LoadInt64(raw.(*int64))
}

// ============================================================================
// Sequence gap detection
// ============================================================================

// SequenceGap represents a detected gap in the event sequence for a workspace.
type SequenceGap struct {
	WorkspaceID string
	Expected    int64 // the sequence number we expected
	Got         int64 // the sequence number we actually received
}

// DetectGap checks whether a received sequence number is contiguous with the
// last seen sequence for the workspace. Returns nil if contiguous, or a
// SequenceGap describing the gap.
func DetectGap(workspaceID string, lastSeen, current int64) *SequenceGap {
	if current == lastSeen+1 || lastSeen == 0 {
		return nil
	}
	return &SequenceGap{
		WorkspaceID: workspaceID,
		Expected:    lastSeen + 1,
		Got:         current,
	}
}

// SnapshotFilter defines which issues to include in the startup snapshot.
// Used by the bus startup to seed state for key issues (blocked, in_review, etc.).
type SnapshotFilter struct {
	Statuses       []string // e.g. ["blocked", "in_review"]
	MetadataKey    string   // e.g. "dispatch_contract_id"
	MetadataHasKey bool     // true → key must exist; false → key must NOT exist
}
