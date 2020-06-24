package component

// LeaderCallbacks defines callbacks for events from LeaderElector
// It is guaranteed that each methods will be executed from the same goroutine, so only one method can be run at once.
type LeaderCallbacks struct {
	OnStartedLeading func()
	OnStoppedLeading func()
}

type LeaderElector interface {
	AddCallbacks(LeaderCallbacks)
	// IsLeader should be used for diagnostic reasons (metrics/API info), because there may not be any leader elector for a short period of time.
	// Use Callbacks to write logic to execute when Leader is elected.
	IsLeader() bool

	// Start blocks until the channel is closed or an error occurs.
	Start(stop <-chan struct{})
}
