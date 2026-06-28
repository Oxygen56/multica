package protocol

// Cross-squad notification system event types (Feature #4596).
// These events extend the existing WebSocket / event bus vocabulary
// with contract and notification-delivery signals.
const (
	// Notification events — fired by the rule engine when a notification is
	// delivered across issues.
	EventNotificationDelivered = "notification:delivered"

	// Dispatch contract events — lifecycle of a formal dispatch_contract row.
	EventContractCreated   = "contract:created"
	EventContractTriggered = "contract:triggered"
	EventContractFulfilled = "contract:fulfilled"
	EventContractCancelled = "contract:cancelled"

	// Anomaly detection events — fired when a deadlock, timeout, or silence
	// is detected.
	EventAnomalyDetected = "anomaly:detected"

	// Child-terminal synthesized event — fired by the rule engine after a child
	// issue reaches terminal status, distinct from the generic status_changed
	// so cross-squad subscribers can filter precisely.
	EventChildTerminal = "child:terminal"

	// Stage-closed synthesized event — fired when every child in the lowest
	// unfinished stage of a parent becomes terminal.
	EventStageClosed = "stage:closed"
)
