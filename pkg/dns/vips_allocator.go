package dns

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var vipsAllocatorLog = core.Log.WithName("dns-vips-allocator")

type (
	// VIPsAllocator takes service tags in dataplanes and allocate VIP for every unique service. It is only run by CP leader.
	VIPsAllocator interface {
		Start(<-chan struct{}) error
		NeedLeaderElection() bool
	}

	vipsAllocator struct {
		rm          manager.ReadOnlyResourceManager
		ipam        IPAM
		persistence *DNSPersistence
		resolver    DNSResolver
		newTicker   func() *time.Ticker
	}
)

func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, persistence *DNSPersistence, ipam IPAM, resolver DNSResolver) (VIPsAllocator, error) {
	return &vipsAllocator{
		rm:          rm,
		persistence: persistence,
		ipam:        ipam,
		resolver:    resolver,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (d *vipsAllocator) NeedLeaderElection() bool {
	return true
}

func (d *vipsAllocator) Start(stop <-chan struct{}) error {
	ticker := d.newTicker()
	defer ticker.Stop()

	vipsAllocatorLog.Info("starting the DNS VIPs allocator")
	for {
		select {
		case <-ticker.C:
			if err := d.synchronize(); err != nil {
				vipsAllocatorLog.Error(err, "unable to synchronise")
			}
		case <-stop:
			vipsAllocatorLog.Info("stopping")
			return nil
		}
	}
}

func (d *vipsAllocator) synchronize() error {
	meshes := core_mesh.MeshResourceList{}
	err := d.rm.List(context.Background(), &meshes)
	if err != nil {
		return err
	}

	serviceMap := make(map[string]bool)
	for _, mesh := range meshes.Items {
		dataplanes := core_mesh.DataplaneResourceList{}

		err := d.rm.List(context.Background(), &dataplanes, store.ListByMesh(mesh.Meta.GetName()))
		if err != nil {
			return err
		}

		// TODO: Do we need to reflect somehow the fact this service belongs to a particular `mesh`
		for _, dp := range dataplanes.Items {
			if dp.Spec.IsIngress() {
				for _, service := range dp.Spec.Networking.Ingress.AvailableServices {
					serviceMap[service.Tags[mesh_proto.ServiceTag]] = true
				}
			} else {
				for _, inbound := range dp.Spec.Networking.Inbound {
					serviceMap[inbound.GetService()] = true
				}
			}
		}

		externalServices := core_mesh.ExternalServiceResourceList{}
		err = d.rm.List(context.Background(), &externalServices, store.ListByMesh(mesh.Meta.GetName()))
		if err != nil {
			return err
		}

		for _, es := range externalServices.Items {
			service := es.Spec.GetService()
			if _, exists := serviceMap[service]; exists {
				vipsAllocatorLog.V(0).Info("Overlapping Extrnal Service name", "service", service)
			}
			serviceMap[service] = true
		}
	}

	return d.allocateVIPs(serviceMap)
}

func (d *vipsAllocator) allocateVIPs(services map[string]bool) (errs error) {
	viplist, err := d.persistence.Get()
	if err != nil {
		return err
	}
	change := false

	// ensure all services have entries in the domain
	for service := range services {
		_, found := viplist[service]
		if !found {
			ip, err := d.ipam.AllocateIP()
			if err != nil {
				errs = multierr.Append(errs, errors.Wrapf(err, "unable to allocate an ip for service %s", service))
			} else {
				viplist[service] = ip
				change = true
				vipsAllocatorLog.Info("Adding", "service", service, "ip", ip)
			}
		}
	}

	// ensure all entries in the domain are present in the service list, and delete them otherwise
	for service := range viplist {
		_, found := services[service]
		if !found {
			ip := viplist[service]
			change = true
			vipsAllocatorLog.Info("Removing", "service", service, "ip", ip)
			_ = d.ipam.FreeIP(ip)
			delete(viplist, service)
		}
	}

	if change {
		if err := d.persistence.Set(viplist); err != nil {
			return err
		}
		d.resolver.SetVIPs(viplist)
	}
	return
}
