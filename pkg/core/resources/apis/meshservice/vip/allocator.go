package vip

import (
	"context"
	"net"
	"time"

	"github.com/Nordix/simple-ipam/pkg/ipam"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

// Allocator manages IPs for MeshServices.
// Each time allocator starts it initiates the IPAM based on existing MeshServices
// We don't free addresses explicitly, but we always allocate next free IP to avoid a problem when we
// 1) Remove Service A with IP X
// 2) Add new Service B that gets IP X
// 3) Clients that were sending the traffic to A now sends the traffic to B for brief amount of time
// IPAM is kept in memory to avoid state management, so technically this problem can still happen when leader changes
// However, leader should not change before TTL of a DNS that serves this VIP.
//
// It's technically possible to allocate all addresses by creating and removing services in the loop.
// However, CIDR has range of 16M addresses, after that the component will just restart.
type Allocator struct {
	logger     logr.Logger
	cidr       string
	interval   time.Duration
	metric     prometheus.Summary
	resManager manager.ResourceManager
}

var _ component.Component = &Allocator{}

func NewAllocator(
	logger logr.Logger,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	cidr string,
	interval time.Duration,
) (*Allocator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_ms_vip_allocator",
		Help:       "Summary of Inter CP Heartbeat component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}

	return &Allocator{
		logger:     logger,
		resManager: resManager,
		cidr:       cidr,
		interval:   interval,
		metric:     metric,
	}, nil
}

func (a *Allocator) Start(stop <-chan struct{}) error {
	a.logger.Info("starting")
	ticker := time.NewTicker(a.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	kumaIpam, err := a.initIpam(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := a.allocateVips(ctx, kumaIpam); err != nil {
				return err
			}
			a.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			a.logger.Info("stopping")
			return nil
		}
	}
}

func (a *Allocator) NeedLeaderElection() bool {
	return true
}

func (a *Allocator) initIpam(ctx context.Context) (*ipam.IPAM, error) {
	newIPAM, err := ipam.New(a.cidr)
	if err != nil {
		return nil, errors.Wrapf(err, "could not allocate IPAM of CIDR %s", a.cidr)
	}

	services := &meshservice_api.MeshServiceResourceList{}
	if err := a.resManager.List(ctx, services); err != nil {
		return nil, errors.Wrap(err, "could not list mesh services for initialization of ipam")
	}
	for _, service := range services.Items {
		for _, vip := range service.Status.VIPs {
			_ = newIPAM.Reserve(net.ParseIP(vip.IP)) // ignore error for outside of range
		}
	}

	return newIPAM, nil
}

func (a *Allocator) allocateVips(ctx context.Context, kumaIpam *ipam.IPAM) error {
	services := &meshservice_api.MeshServiceResourceList{}
	if err := a.resManager.List(ctx, services); err != nil {
		return errors.Wrap(err, "could not list mesh services for ip allocation")
	}

	for _, svc := range services.Items {
		if len(svc.Status.VIPs) == 0 {
			log := a.logger.WithValues("service", svc.Meta.GetName(), "mesh", svc.Meta.GetMesh())
			ip, err := kumaIpam.Allocate()
			if err != nil {
				return errors.Wrapf(err, "could not allocate the address for svc %s", svc.Meta.GetName())
			}
			log.Info("allocating IP for a service", "ip", ip.String())
			svc.Status.VIPs = []meshservice_api.VIP{
				{
					IP: ip.String(),
				},
			}

			if err := a.resManager.Update(ctx, svc); err != nil {
				msg := "could not update service with allocated Kuma VIP. Will try to update in the next allocation window"
				if errors.Is(err, &store.ResourceConflictError{}) {
					log.Info(msg, "cause", "conflict", "interval", a.interval)
				} else {
					log.Error(err, msg, "interval", a.interval)
				}
			}
		}
	}
	return nil
}
