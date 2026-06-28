package notifications

import (
	"context"

	"github.com/multica-ai/multica/server/internal/events"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// WireUp creates and starts the notification rule engine, registering all
// five rules (R1–R5) and wiring them into the event bus. Call this from
// cmd/server during startup, after the event bus and DB queries are ready.
//
// Usage in cmd/server/main.go or a server init function:
//
//	notifEngine := notifications.WireUp(bus, queries)
//	defer notifEngine.Shutdown()
func WireUp(bus *events.Bus, queries *db.Queries) *Engine {
	deps := RuleDeps{
		Queries: queries,
		Bus:     bus,
	}

	engine := NewEngine(bus, queries, deps)

	// Register rules in priority order (lowest number first = highest priority).
	// R5 (10) > R2 (20) > R4 (30) > R3 (40) > R1 (50)
	engine.RegisterRule(BuildR5(queries))
	engine.RegisterRule(BuildR2(queries))
	engine.RegisterRule(BuildR4(queries))
	engine.RegisterRule(BuildR3(queries))
	engine.RegisterRule(BuildR1(queries))

	engine.Start()

	return engine
}

// Shutdown gracefully stops the engine and its background workers.
func (e *Engine) Shutdown() {
	if e.persist != nil {
		e.persist.Shutdown()
	}
}

// DummyContext creates a background context for operations outside
// request scopes, matching the existing pattern in notification_listeners.go.
func DummyContext() context.Context {
	return context.Background()
}
