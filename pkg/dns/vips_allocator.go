package dns

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var vipsAllocatorLog = core.Log.WithName("dns-vips-allocator")

type VIPsAllocator struct {
	rm          manager.ReadOnlyResourceManager
	ipam        IPAM
	persistence *vips.Persistence
	resolver    resolver.DNSResolver
	newTicker   func() *time.Ticker
}

// NewVIPsAllocator creates new object of VIPsAllocator. You can either
// call method CreateOrUpdateVIPConfig manually or start VIPsAllocator as a component.
// In the latter scenario it will call CreateOrUpdateVIPConfig every 'tickInterval'
// for all meshes in the store.
func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, cidr string, resolver resolver.DNSResolver) (*VIPsAllocator, error) {
	ipam, err := NewSimpleIPAM(cidr)
	if err != nil {
		return nil, err
	}
	return &VIPsAllocator{
		rm:          rm,
		persistence: vips.NewPersistence(rm, configManager),
		ipam:        ipam,
		resolver:    resolver,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (d *VIPsAllocator) NeedLeaderElection() bool {
	return true
}

func (d *VIPsAllocator) Start(stop <-chan struct{}) error {
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

func (d *VIPsAllocator) createOrUpdateVIPConfigs() error {
	meshes := core_mesh.MeshResourceList{}
	err := d.rm.List(context.Background(), &meshes)
	if err != nil {
		return err
	}

	for _, mesh := range meshes.Items {
		if err := d.CreateOrUpdateVIPConfig(mesh.GetMeta().GetName()); err != nil {
			vipsAllocatorLog.Error(err, "unable to create or update VIP config", "mesh", mesh.GetMeta().GetName())
		}
	}
	return nil
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfig(mesh string) error {
	serviceSet, err := BuildServiceSet(d.rm, mesh)
	if err != nil {
		return err
	}

	global, err := d.persistence.Get()
	if err != nil {
		return err
	}

	meshed, err := d.persistence.GetByMesh(mesh)
	if err != nil {
		return err
	}

	updated, updError := UpdateMeshedVIPs(global, meshed, d.ipam, serviceSet)
	if !updated {
		return err
	}

	if err := d.persistence.Set(mesh, meshed); err != nil {
		return multierr.Append(updError, err)
	}

	d.resolver.SetVIPs(meshed)

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

func UpdateMeshedVIPs(global, meshed vips.List, ipam IPAM, serviceSet ServiceSet) (updated bool, errs error) {
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
