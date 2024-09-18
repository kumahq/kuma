package vip

import (
	"context"
	"net"
	"time"

	"github.com/Nordix/simple-ipam/pkg/ipam"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_time "github.com/kumahq/kuma/pkg/util/time"
)

type ResourceHoldingVIPs interface {
	model.Resource
	VIPs() []string
	AllocateVIP(vip string)
}

// Allocator manages IPs for resources holding vips like MeshService/MeshExternalServices/MeshMultiZoneService.
// Each time allocator starts it initiates the IPAM based on existing vips
// We don't free addresses explicitly, but we always allocate next free IP to avoid a problem when we
// 1) Remove MeshService A with IP X
// 2) Add new MeshService B that gets IP X
// 3) Clients that were sending the traffic to A now sends the traffic to B for brief amount of time
// IPAM is kept in memory to avoid state management, so technically this problem can still happen when leader changes
// However, leader should not change before TTL of a DNS that serves this VIP.
//
// It's technically possible to allocate all addresses by creating and removing services in the loop.
// However, CIDR has range of 16M addresses, after that the component will just restart.
type Allocator struct {
	logger            logr.Logger
	interval          time.Duration
	cidrToDescriptors map[string]model.ResourceTypeDescriptor
	resManager        manager.ResourceManager
	metric            prometheus.Summary
}

var _ component.Component = &Allocator{}

func NewAllocator(
	logger logr.Logger,
	interval time.Duration,
	cidrToDescriptors map[string]model.ResourceTypeDescriptor,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
) (*Allocator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_vip_allocator",
		Help:       "Summary of VIP allocation duration",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Allocator{
		logger:            logger,
		interval:          interval,
		cidrToDescriptors: cidrToDescriptors,
		metric:            metric,
		resManager:        resManager,
	}, nil
}

func (a *Allocator) Start(stop <-chan struct{}) error {
	// sleep to mitigate update conflicts with other components
	util_time.SleepUpTo(a.interval)
	a.logger.Info("starting")
	ticker := time.NewTicker(a.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	ipams := map[string]*ipam.IPAM{}

	for cidr, typeDescriptor := range a.cidrToDescriptors {
		newIPAM, err := ipam.New(cidr)
		if err != nil {
			return errors.Wrapf(err, "could not allocate IPAM of CIDR %s", cidr)
		}

		resources, err := a.listResourceHoldingVIPs(ctx, typeDescriptor)
		if err != nil {
			return errors.Wrapf(err, "could not list resources for IPAM initialization for %s", typeDescriptor.Name)
		}

		for _, res := range resources {
			for _, vip := range res.VIPs() {
				_ = newIPAM.Reserve(net.ParseIP(vip)) // ignore error for outside of range
			}
		}
		ipams[cidr] = newIPAM
	}

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			for cidr, typeDesc := range a.cidrToDescriptors {
				if err := a.allocateVIPs(ctx, typeDesc, ipams[cidr]); err != nil {
					a.logger.Error(err, "could not allocate vips", "type", typeDesc.Name)
				}
			}
			a.metric.Observe(float64(time.Since(start).Milliseconds()))

		case <-stop:
			a.logger.Info("stopping")
			return nil
		}
	}
}

func (a *Allocator) listResourceHoldingVIPs(ctx context.Context, typeDesc model.ResourceTypeDescriptor) ([]ResourceHoldingVIPs, error) {
	list := typeDesc.NewList()
	if err := a.resManager.List(ctx, list); err != nil {
		return nil, err
	}
	result := make([]ResourceHoldingVIPs, len(list.GetItems()))
	for i, res := range list.GetItems() {
		result[i] = res.(ResourceHoldingVIPs)
	}
	return result, nil
}

func (a *Allocator) allocateVIPs(ctx context.Context, typeDesc model.ResourceTypeDescriptor, kumaIpam *ipam.IPAM) error {
	resources, err := a.listResourceHoldingVIPs(ctx, typeDesc)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		if len(resource.VIPs()) == 0 {
			log := a.logger.WithValues(
				"name", resource.GetMeta().GetName(),
				"mesh", resource.GetMeta().GetMesh(),
				"type", resource.Descriptor().Name,
			)
			ip, err := kumaIpam.Allocate()
			if err != nil {
				return errors.Wrapf(err, "could not allocate vip for %s %s", typeDesc.Name, resource.GetMeta().GetName())
			}
			log.Info("allocating IP", "ip", ip.String())
			resource.AllocateVIP(ip.String())

			if err := a.resManager.Update(ctx, resource); err != nil {
				msg := "could not update the resource with allocated Kuma VIP. Will try to update in the next allocation window"
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

func (a *Allocator) NeedLeaderElection() bool {
	return true
}
