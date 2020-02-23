package broker

// private type for Context keys
type contextKey int

const (
	clientIDKey contextKey = iota

	// TODO: Polling time for subscribe (edgenode and edgeserver)
	maxPollTime = 100
)
