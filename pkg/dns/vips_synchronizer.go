package dns

import (
	"time"

	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

var (
	vipsSynchronizerLog = core.Log.WithName("dns-vips-synchronizer")
)

type vipsSynchronizer struct {
	resolver    resolver.DNSResolver
	persistence *vips.Persistence
	leadInfo    component.LeaderInfo
	newTicker   func() *time.Ticker
}

const (
	tickInterval = 500 * time.Millisecond
)

func NewVIPsSynchronizer(resolver resolver.DNSResolver, rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, leadInfo component.LeaderInfo) component.Component {
	return &vipsSynchronizer{
		resolver:    resolver,
		persistence: vips.NewPersistence(rm, configManager),
		leadInfo:    leadInfo,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}
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
	voByMesh, err := d.persistence.Get()
	if err != nil {
		return err
	}
	d.resolver.SetVIPs(vips.ToVIPMap(voByMesh))
	return nil
}
