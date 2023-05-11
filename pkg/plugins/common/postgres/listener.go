package postgres

type Listener interface {
	Notify() chan *Notification
	Error() <-chan error
	Close() error
}

// Notification represents a single notification from the database.
type Notification struct {
	// Payload, or the empty string if unspecified.
	Payload string
}

const channelName = "resource_events"
