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
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var _ vip.VIPAllocator = &MeshExternalServiceAllocator{}

type MeshExternalServiceAllocator struct {
	logger     logr.Logger
	cidr       string
	resManager manager.ResourceManager
	interval   time.Duration
	kumaIpam   *ipam.IPAM
	metric     prometheus.Summary
}

func NewMeshExternalServiceAllocator(
	logger logr.Logger,
	cidr string,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
) (*MeshExternalServiceAllocator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_mes_vip_allocator",
		Help:       "Summary of Inter CP Heartbeat component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &MeshExternalServiceAllocator{
		logger:     logger,
		cidr:       cidr,
		resManager: resManager,
		interval:   interval,
		metric:     metric,
	}, nil
}

func (a *MeshExternalServiceAllocator) InitIPAM(ctx context.Context) error {
	newIPAM, err := ipam.New(a.cidr)
	if err != nil {
		return errors.Wrapf(err, "could not allocate IPAM of CIDR %s", a.cidr)
	}

	es := &meshexternalservice_api.MeshExternalServiceResourceList{}
	if err := a.resManager.List(ctx, es); err != nil {
		return errors.Wrap(err, "could not list mesh external services for initialization of ipam")
	}
	for _, service := range es.Items {
		_ = newIPAM.Reserve(net.ParseIP(service.Status.VIP.IP)) // ignore error for outside of range
	}
	a.kumaIpam = newIPAM
	return nil
}

func (a *MeshExternalServiceAllocator) AllocateVIPs(ctx context.Context) error {
	services := &meshexternalservice_api.MeshExternalServiceResourceList{}
	if err := a.resManager.List(ctx, services); err != nil {
		return errors.Wrap(err, "could not list mesh services for ip allocation")
	}

	for _, svc := range services.Items {
		if svc.Status.VIP.IP == "" {
			log := a.logger.WithValues("external service", svc.Meta.GetName(), "mesh", svc.Meta.GetMesh())
			ip, err := a.kumaIpam.Allocate()
			if err != nil {
				return errors.Wrapf(err, "could not allocate the address for external service %s", svc.Meta.GetName())
			}
			log.Info("allocating IP for a external service", "ip", ip.String())
			svc.Status.VIP = meshexternalservice_api.VIP{IP: ip.String()}

			if err := a.resManager.Update(ctx, svc); err != nil {
				msg := "could not update external service with allocated Kuma VIP. Will try to update in the next allocation window"
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

func (a *MeshExternalServiceAllocator) ReportDuration(start time.Time) {
	a.metric.Observe(float64(time.Since(start).Milliseconds()))
}
