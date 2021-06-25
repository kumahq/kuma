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
	persistence *vips.Persistence
	resolver    resolver.DNSResolver
	newTicker   func() *time.Ticker
	cidr        string
}

// NewVIPsAllocator creates new object of VIPsAllocator. You can either
// call method CreateOrUpdateVIPConfig manually or start VIPsAllocator as a component.
// In the latter scenario it will call CreateOrUpdateVIPConfig every 'tickInterval'
// for all meshes in the store.
func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, cidr string, resolver resolver.DNSResolver) (*VIPsAllocator, error) {
	return &VIPsAllocator{
		rm:          rm,
		persistence: vips.NewPersistence(rm, configManager),
		cidr:        cidr,
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
			if err := d.CreateOrUpdateVIPConfigs(); err != nil {
				vipsAllocatorLog.Error(err, "errors during updating VIP configs")
			}
		case <-stop:
			vipsAllocatorLog.Info("stopping")
			return nil
		}
	}
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfigs() error {
	meshRes := core_mesh.MeshResourceList{}
	if err := d.rm.List(context.Background(), &meshRes); err != nil {
		return err
	}

	meshes := []string{}
	for _, mesh := range meshRes.Items {
		meshes = append(meshes, mesh.GetMeta().GetName())
	}

	return d.createOrUpdateVIPConfigs(meshes...)
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfig(mesh string) error {
	return d.createOrUpdateVIPConfigs(mesh)
}

func (d *VIPsAllocator) createOrUpdateVIPConfigs(meshes ...string) (errs error) {
	global, byMesh, err := d.persistence.Get()
	if err != nil {
		return err
	}

	ipam, err := d.newIPAM(global)
	if err != nil {
		return err
	}

	forEachMesh := func(mesh string, meshed vips.List) error {
		serviceSet, err := BuildServiceSet(d.rm, mesh)
		if err != nil {
			return err
		}

		changed, err := UpdateMeshedVIPs(global, meshed, ipam, serviceSet)
		if err != nil {
			// Error might occur only if we run out of VIPs. There is no point to pass it through,
			// we must notify user in logs and proceed
			vipsAllocatorLog.Error(err, "failed to allocate new VIPs")
		}
		if !changed {
			return nil
		}
		global.Append(meshed)

		return d.persistence.Set(mesh, meshed)
	}

	for _, mesh := range meshes {
		meshed, ok := byMesh[mesh]
		if !ok {
			meshed = vips.List{}
		}
		if err := forEachMesh(mesh, meshed); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	d.resolver.SetVIPs(global)

	return errs
}

func (d *VIPsAllocator) newIPAM(initialVIPs vips.List) (IPAM, error) {
	ipam, err := NewSimpleIPAM(d.cidr)
	if err != nil {
		return nil, err
	}

	for _, vip := range initialVIPs {
		if err := ipam.ReserveIP(vip); err != nil && !IsAddressAlreadyAllocated(err) {
			return nil, err
		}
	}

	return ipam, nil
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
	if err := rm.List(context.Background(), &dataplanes); err != nil {
		return nil, err
	}

	filteredDataplanes := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if d.GetMeta().GetMesh() == mesh || d.Spec.IsIngress() {
			_ = filteredDataplanes.AddItem(d)
		}
	}

	for _, dp := range filteredDataplanes.Items {
		// backwards compatibility
		if dp.Spec.IsIngress() {
			for _, service := range dp.Spec.GetNetworking().GetIngress().GetAvailableServices() {
				if service.Mesh != mesh {
					continue
				}
				serviceSet[service.Tags[mesh_proto.ServiceTag]] = true
			}
		} else {
			for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
				serviceSet[inbound.GetService()] = true
			}
		}
	}

	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := rm.List(context.Background(), &zoneIngresses); err != nil {
		return nil, err
	}

	for _, zi := range zoneIngresses.Items {
		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh != mesh {
				continue
			}
			serviceSet[service.Tags[mesh_proto.ServiceTag]] = true
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
		vipsAllocatorLog.Info("adding", "service", service, "ip", ip)
	}
	for service, ip := range meshed {
		if _, found := serviceSet[service]; !found {
			updated = true
			_ = ipam.FreeIP(ip)
			delete(meshed, service)
			vipsAllocatorLog.Info("deleting", "service", service, "ip", ip)
		}
	}
	return
}
