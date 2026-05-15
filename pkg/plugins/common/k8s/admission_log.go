package k8s

import (
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/v2/pkg/core"
)

var log = core.Log.WithName("webhooks")

const (
	rejectionSampleWindow = 10 * time.Second
	maxRejectionReasonLen = 512
)

type rejectionLimiterEntry struct {
	last       time.Time
	suppressed int
}

type rejectionLimiter struct {
	window  time.Duration
	mu      sync.Mutex
	entries map[string]rejectionLimiterEntry
}

func (l *rejectionLimiter) shouldLog(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e := l.entries[key]
	if !e.last.IsZero() && now.Sub(e.last) < l.window {
		e.suppressed++
		l.entries[key] = e
		return false, 0
	}
	n := e.suppressed
	l.entries[key] = rejectionLimiterEntry{last: now}
	return true, n
}

var defaultRejectionLimiter = &rejectionLimiter{
	window:  rejectionSampleWindow,
	entries: map[string]rejectionLimiterEntry{},
}

// LogWebhookRejection logs one event per rejection with a per-resource sample
// window so a misconfigured controller in a tight reconcile loop cannot flood
// the log. Reason is truncated to bound line size.
func LogWebhookRejection(req admission.Request, resp admission.Response) {
	reason := ""
	if resp.Result != nil {
		reason = resp.Result.Message
	}
	if len(reason) > maxRejectionReasonLen {
		reason = reason[:maxRejectionReasonLen] + "...(truncated)"
	}
	key := req.Kind.Kind + "/" + req.Namespace + "/" + req.Name
	ok, suppressed := defaultRejectionLimiter.shouldLog(key)
	if !ok {
		return
	}
	kv := []any{
		"kind", req.Kind.Kind,
		"name", req.Name,
		"namespace", req.Namespace,
		"operation", req.Operation,
		"reason", reason,
	}
	if suppressed > 0 {
		kv = append(kv, "suppressed", suppressed)
	}
	log.Info("webhook rejected resource", kv...)
}
