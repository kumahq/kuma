package vip

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
)

type VIPAllocator interface {
	ReportDuration(start time.Time)
	InitIPAM(ctx context.Context) error
	AllocateVIPs(context.Context) error
}

// Allocator manages IPs for MeshExternalServices.
// Each time allocator starts it initiates the IPAM based on existing MeshExternalServices
// We don't free addresses explicitly, but we always allocate next free IP to avoid a problem when we
// 1) Remove MeshExternalService A with IP X
// 2) Add new MeshExternalService B that gets IP X
// 3) Clients that were sending the traffic to A now sends the traffic to B for brief amount of time
// IPAM is kept in memory to avoid state management, so technically this problem can still happen when leader changes
// However, leader should not change before TTL of a DNS that serves this VIP.
//
// It's technically possible to allocate all addresses by creating and removing services in the loop.
// However, CIDR has range of 16M addresses, after that the component will just restart.
type Allocator struct {
	logger     logr.Logger
	interval   time.Duration
	allocators []VIPAllocator
}

var _ component.Component = &Allocator{}

func NewAllocator(
	logger logr.Logger,
	interval time.Duration,
	allocators []VIPAllocator,
) (*Allocator, error) {
	return &Allocator{
		logger:     logger,
		interval:   interval,
		allocators: allocators,
	}, nil
}

func (a *Allocator) Start(stop <-chan struct{}) error {
	a.logger.Info("starting")
	ticker := time.NewTicker(a.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for _, allocator := range a.allocators {
		if err := allocator.InitIPAM(ctx); err != nil {
			return err
		}
	}

	for {
		select {
		case <-ticker.C:
			for _, allocator := range a.allocators {
				start := time.Now()
				if err := allocator.AllocateVIPs(ctx); err != nil {
					return err
				}
				allocator.ReportDuration(start)
			}

		case <-stop:
			a.logger.Info("stopping")
			return nil
		}
	}
}

func (a *Allocator) NeedLeaderElection() bool {
	return true
}
