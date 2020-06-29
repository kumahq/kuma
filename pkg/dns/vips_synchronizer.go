package dns

import (
	"time"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/runtime/component"
)

var (
	vipsSynchronizerLog = core.Log.WithName("dns-vips-synchronizer")
)

type (
	// VIPsSynchronizer takes allocated VIPs by VIPsAllocator and updates DNSResolver.
	VIPsSynchronizer interface {
		Start(<-chan struct{}) error
		NeedLeaderElection() bool
	}

	vipsSynchronizer struct {
		rm          manager.ReadOnlyResourceManager
		resolver    DNSResolver
		persistence *DNSPersistence
		leadInfo    component.LeaderInfo
		newTicker   func() *time.Ticker
	}
)

const (
	tickInterval = 500 * time.Millisecond
)

func NewVIPsSynchronizer(rm manager.ReadOnlyResourceManager, resolver DNSResolver, persistence *DNSPersistence, leadInfo component.LeaderInfo) (VIPsSynchronizer, error) {
	return &vipsSynchronizer{
		rm:          rm,
		resolver:    resolver,
		persistence: persistence,
		leadInfo:    leadInfo,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (d *vipsSynchronizer) NeedLeaderElection() bool {
	return false
}

func (d *vipsSynchronizer) Start(stop <-chan struct{}) error {
	ticker := d.newTicker()
	defer ticker.Stop()

	vipsSynchronizerLog.Info("starting the DNS VIPs Synchronizer")
	for {
		select {
		case <-ticker.C:
			if err := d.synchronize(); err != nil {
				vipsSynchronizerLog.Error(err, "unable to synchronise")
			}
		case <-stop:
			vipsSynchronizerLog.Info("stopping")
			return nil
		}
	}
}

func (d *vipsSynchronizer) synchronize() error {
	if d.leadInfo.IsLeader() {
		return nil // when CP is leader we skip this because VIP allocator updates DNSResolver
	}
	vipList, err := d.persistence.Get()
	if err != nil {
		return err
	}
	d.resolver.SetVIPs(vipList)
	return nil
}
