package poller

import (
	"context"
	"fmt"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/core"
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
		Name   string `json:"name"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
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
	p.Unlock()

	p.zones = Zones{}

	zones := &system.ZoneResourceList{}
	p.roResourceManager.List(context.Background(), zones)
	for _, zone := range zones.Items {
		// ignore the Ingress for now
		p.zones = append(p.zones, Zone{
			Name:   zone.Meta.GetName(),
			URL:    zone.Spec.GetIngress().GetPublicAddress(),
			Active: false,
		})
	}
}

func (p *ZonesStatusPoller) pollZones() {
	p.Lock()
	defer p.Unlock()

	for i, zone := range p.zones {
		u, err := url.Parse(zone.URL)
		if err != nil {
			zonesStatusLog.Info(fmt.Sprintf("failed to parse URL %s", zone.URL))
			continue
		}
		conn, err := net.DialTimeout("tcp", u.Host, dialTimeout)
		if err != nil {
			if zone.Active {
				zonesStatusLog.Info(fmt.Sprintf("%s at %s did not respond", zone.Name, zone.URL))
				p.zones[i].Active = false
			}
			continue
		}
		defer conn.Close()

		if !p.zones[i].Active {
			zonesStatusLog.Info(fmt.Sprintf("%s responded", zone.URL))
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
