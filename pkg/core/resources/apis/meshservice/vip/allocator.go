package vip

import (
	"context"
	"net"
	"time"

	"github.com/Nordix/simple-ipam/pkg/ipam"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var _ vip.VIPAllocator = &MeshServiceAllocator{}

type MeshServiceAllocator struct {
	logger     logr.Logger
	cidr       string
	resManager manager.ResourceManager
	interval   time.Duration
	kumaIpam   *ipam.IPAM
	metric     prometheus.Summary
}

func NewMeshServiceAllocator(
	logger logr.Logger,
	cidr string,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
) (*MeshServiceAllocator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_ms_vip_allocator",
		Help:       "Summary of Inter CP Heartbeat component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &MeshServiceAllocator{
		logger:     logger,
		cidr:       cidr,
		resManager: resManager,
		interval:   interval,
		metric:     metric,
	}, nil
}

func (a *MeshServiceAllocator) InitIPAM(ctx context.Context) error {
	newIPAM, err := ipam.New(a.cidr)
	if err != nil {
		return errors.Wrapf(err, "could not allocate IPAM of CIDR %s", a.cidr)
	}
	services := &meshservice_api.MeshServiceResourceList{}
	if err := a.resManager.List(ctx, services); err != nil {
		return errors.Wrap(err, "could not list mesh services for initialization of ipam")
	}
	for _, service := range services.Items {
		for _, vip := range service.Status.VIPs {
			_ = newIPAM.Reserve(net.ParseIP(vip.IP)) // ignore error for outside of range
		}
	}
	a.kumaIpam = newIPAM
	return nil
}

func (a *MeshServiceAllocator) AllocateVIPs(ctx context.Context) error {
	services := &meshservice_api.MeshServiceResourceList{}
	if err := a.resManager.List(ctx, services); err != nil {
		return errors.Wrap(err, "could not list mesh services for ip allocation")
	}

	for _, svc := range services.Items {
		if len(svc.Status.VIPs) == 0 {
			log := a.logger.WithValues("service", svc.Meta.GetName(), "mesh", svc.Meta.GetMesh())
			ip, err := a.kumaIpam.Allocate()
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

func (a *MeshServiceAllocator) ReportDuration(start time.Time) {
	a.metric.Observe(float64(time.Since(start).Milliseconds()))
}
