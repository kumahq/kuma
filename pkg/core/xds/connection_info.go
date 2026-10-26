package xds

import (
	"sync"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// ConnectionInfo carries auth-time facts identity issuance honors later: the
// stream authenticates once, but the workload identity is reissued on changes.
type ConnectionInfo struct {
	// TokenWorkload is the workload the token is bound to, empty if unbound.
	TokenWorkload string
}

// ConnectionInfoTracker holds the ConnectionInfo of each connected dataplane.
// Only token-authenticated dataplanes get an entry; absence means don't enforce.
type ConnectionInfoTracker interface {
	Store(key core_model.ResourceKey, info ConnectionInfo)
	Get(key core_model.ResourceKey) (ConnectionInfo, bool)
	Delete(key core_model.ResourceKey)
}

type connectionInfoTracker struct {
	sync.RWMutex
	infos map[core_model.ResourceKey]ConnectionInfo
}

func NewConnectionInfoTracker() ConnectionInfoTracker {
	return &connectionInfoTracker{
		infos: map[core_model.ResourceKey]ConnectionInfo{},
	}
}

func (t *connectionInfoTracker) Store(key core_model.ResourceKey, info ConnectionInfo) {
	t.Lock()
	defer t.Unlock()
	t.infos[key] = info
}

func (t *connectionInfoTracker) Get(key core_model.ResourceKey) (ConnectionInfo, bool) {
	t.RLock()
	defer t.RUnlock()
	info, ok := t.infos[key]
	return info, ok
}

func (t *connectionInfoTracker) Delete(key core_model.ResourceKey) {
	t.Lock()
	defer t.Unlock()
	delete(t.infos, key)
}
