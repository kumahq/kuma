package poller

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"

	"github.com/kumahq/kuma/pkg/core"
)

var (
	zonesStatusLog = core.Log.WithName("zones-status")
)

type (
	ZoneStatusPoller interface {
		Start(<-chan struct{}) error
		NeedLeaderElection() bool
		Zones() Zones
	}

	Zone struct {
		Name    string `json:"name"`
		Address string `json:"url"`
		Active  bool   `json:"active"`
	}

	Zones []Zone

	ZonesStatusPoller struct {
		sync.RWMutex
		zones             Zones
		roResourceManager manager.ReadOnlyResourceManager
		newTicker         func() *time.Ticker
	}
)

const (
	tickInterval = 15 * time.Second
	dialTimeout  = 100 * time.Millisecond
)

func NewZonesStatusPoller(roResourceManager manager.ReadOnlyResourceManager) (ZoneStatusPoller, error) {
	poller := &ZonesStatusPoller{
		zones:             Zones{},
		roResourceManager: roResourceManager,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}

	return poller, nil
}

func (p *ZonesStatusPoller) Start(stop <-chan struct{}) error {
	ticker := p.newTicker()
	defer ticker.Stop()

	// update the status before running the API
	p.pollZones()

	zonesStatusLog.Info("starting the Zones polling")
	for {
		select {
		case <-ticker.C:
			p.syncZones()
			p.pollZones()
		case <-stop:
			zonesStatusLog.Info("Stopping down API Server")
			return nil
		}
	}
}

func (p *ZonesStatusPoller) NeedLeaderElection() bool {
	return false
}

func (p *ZonesStatusPoller) syncZones() {
	p.Lock()
	defer p.Unlock()

	p.zones = Zones{}

	zones := &system.ZoneResourceList{}
	err := p.roResourceManager.List(context.Background(), zones)
	if err != nil {
		zonesStatusLog.Error(err, "Unable to list Zone resources")
	}

	for _, zone := range zones.Items {
		ingressAddress := zone.Spec.GetIngress().GetPublicAddress()
		_, _, err := net.SplitHostPort(ingressAddress)
		if err != nil {
			zonesStatusLog.Info(fmt.Sprintf("failed to parse the ingress address %s", ingressAddress))
			continue
		}
		p.zones = append(p.zones, Zone{
			Name:    zone.Meta.GetName(),
			Address: ingressAddress,
			Active:  false,
		})
	}
}

func (p *ZonesStatusPoller) pollZones() {
	p.Lock()
	defer p.Unlock()

	for i, zone := range p.zones {
		conn, err := net.DialTimeout("tcp", zone.Address, dialTimeout)
		if err != nil {
			if zone.Active {
				zonesStatusLog.Info(fmt.Sprintf("%s at %s did not respond", zone.Name, zone.Address))
				p.zones[i].Active = false
			}
			continue
		}
		defer conn.Close()

		if !p.zones[i].Active {
			zonesStatusLog.Info(fmt.Sprintf("%s responded", zone.Address))
			p.zones[i].Active = true
		}
	}
}

func (p *ZonesStatusPoller) Zones() Zones {
	p.RLock()
	defer p.RUnlock()
	newZones := Zones{}
	return append(newZones, p.zones...)
}
