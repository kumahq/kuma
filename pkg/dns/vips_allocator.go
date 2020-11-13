package dns

import (
	"context"
	"sort"
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
		persistence *MeshedPersistence
		resolver    DNSResolver
		newTicker   func() *time.Ticker
	}
)

func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, persistence *MeshedPersistence, ipam IPAM, resolver DNSResolver) (VIPsAllocator, error) {
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
			if err := d.createOrUpdateVIPConfigs(); err != nil {
				vipsAllocatorLog.Error(err, "unable to create or update VIP configs")
			}
		case <-stop:
			vipsAllocatorLog.Info("stopping")
			return nil
		}
	}
}

func (d *vipsAllocator) createOrUpdateVIPConfigs() error {
	meshes := core_mesh.MeshResourceList{}
	err := d.rm.List(context.Background(), &meshes)
	if err != nil {
		return err
	}

	for _, mesh := range meshes.Items {
		if err := CreateOrUpdateVIPConfig(d.persistence, d.rm, d.resolver, d.ipam, mesh.GetMeta().GetName()); err != nil {
			vipsAllocatorLog.Error(err, "unable to create or update VIP config", "mesh", mesh.GetMeta().GetName())
		}
	}
	return nil
}

func CreateOrUpdateVIPConfig(p *MeshedPersistence, rm manager.ReadOnlyResourceManager, r DNSResolver, ipam IPAM, mesh string) error {
	serviceSet, err := BuildServiceSet(rm, mesh)
	if err != nil {
		return err
	}

	global, err := p.Get()
	if err != nil {
		return err
	}

	meshed, err := p.GetByMesh(mesh)
	if err != nil {
		return err
	}

	updated, updError := UpdateMeshedVIPs(global, meshed, ipam, serviceSet)
	if !updated {
		return err
	}

	if err := p.Set(mesh, meshed); err != nil {
		return multierr.Append(updError, err)
	}

	r.SetVIPs(meshed)

	return updError
}

type ServiceSet map[string]bool

func (s ServiceSet) ToArray() (services []string) {
	for service := range s {
		services = append(services, service)
	}
	sort.Strings(services)
	return
}

func BuildServiceSet(rm manager.ReadOnlyResourceManager, mesh string) (ServiceSet, error) {
	serviceSet := make(map[string]bool)

	dataplanes := core_mesh.DataplaneResourceList{}
	if err := rm.List(context.Background(), &dataplanes, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	for _, dp := range dataplanes.Items {
		if dp.Spec.IsIngress() {
			for _, service := range dp.Spec.Networking.Ingress.AvailableServices {
				if service.Mesh != mesh {
					continue
				}
				serviceSet[service.Tags[mesh_proto.ServiceTag]] = true
			}
		} else {
			for _, inbound := range dp.Spec.Networking.Inbound {
				serviceSet[inbound.GetService()] = true
			}
		}
	}

	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := rm.List(context.Background(), &externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	for _, es := range externalServices.Items {
		serviceSet[es.Spec.GetService()] = true
	}

	return serviceSet, nil
}

func UpdateMeshedVIPs(global, meshed VIPList, ipam IPAM, serviceSet ServiceSet) (updated bool, errs error) {
	for _, service := range serviceSet.ToArray() {
		_, found := meshed[service]
		if found {
			continue
		}
		ip, found := global[service]
		if found {
			meshed[service] = ip
			updated = true
			continue
		}
		ip, err := ipam.AllocateIP()
		if err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "unable to allocate an ip for service %s", service))
			continue
		}
		meshed[service] = ip
		updated = true
	}
	for service, ip := range meshed {
		if _, found := serviceSet[service]; !found {
			updated = true
			_ = ipam.FreeIP(ip)
			delete(meshed, service)
		}
	}
	return
}
