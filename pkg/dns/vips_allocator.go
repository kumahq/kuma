package dns

import (
	"context"
	"time"

	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/dns/vips"
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
	byMesh, err := d.persistence.Get()
	if err != nil {
		return err
	}

	gv, err := vips.NewGlobalView(d.cidr)
	if err != nil {
		return err
	}
	for _, mesh := range meshes {
		if _, ok := byMesh[mesh]; !ok {
			byMesh[mesh] = vips.NewVirtualOutboundView(map[vips.Entry]vips.VirtualOutbound{})
		}
		for _, key := range byMesh[mesh].Keys() {
			vo := byMesh[mesh].Get(key)
			err := gv.Reserve(key, vo.Address)
			if err != nil {
				return err
			}
		}
	}

	forEachMesh := func(mesh string, meshed *vips.VirtualOutboundView) error {
		newVirtualOutboundView, err := BuildServiceSet(d.rm, mesh)
		if err != nil {
			return err
		}

		err = AllocateVIPs(gv, newVirtualOutboundView)
		if err != nil {
			// Error might occur only if we run out of VIPs. There is no point to pass it through,
			// we must notify user in logs and proceed
			vipsAllocatorLog.Error(err, "failed to allocate new VIPs", "mesh", mesh)
		}
		changes, out := meshed.Update(newVirtualOutboundView)
		if len(changes) == 0 {
			return nil
		}
		vipsAllocatorLog.Info("mesh vip changes", "mesh", mesh, "changes", changes)
		return d.persistence.Set(mesh, out)
	}

	for _, mesh := range meshes {
		if err := forEachMesh(mesh, byMesh[mesh]); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	d.resolver.SetVIPs(gv.VipList())

	return errs
}

func BuildServiceSet(rm manager.ReadOnlyResourceManager, mesh string) (*vips.VirtualOutboundView, error) {
	outboundSet := vips.NewVirtualOutboundView(map[vips.Entry]vips.VirtualOutbound{})

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
	var err error
	for _, dp := range filteredDataplanes.Items {
		// backwards compatibility
		if dp.Spec.IsIngress() {
			for _, service := range dp.Spec.GetNetworking().GetIngress().GetAvailableServices() {
				if service.Mesh == mesh {
					err = multierr.Append(err, addDefault(outboundSet, service.GetTags()[mesh_proto.ServiceTag]))
				}
			}
		} else {
			for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
				err = multierr.Append(err, addDefault(outboundSet, inbound.GetService()))
			}
		}
	}

	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := rm.List(context.Background(), &zoneIngresses); err != nil {
		return nil, err
	}

	for _, zi := range zoneIngresses.Items {
		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh == mesh {
				err = multierr.Append(err, addDefault(outboundSet, service.GetTags()[mesh_proto.ServiceTag]))
			}
		}
	}

	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := rm.List(context.Background(), &externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	for _, es := range externalServices.Items {
		err = multierr.Append(err, addDefault(outboundSet, es.Spec.GetService()))
		err = multierr.Append(err, outboundSet.Add(vips.NewHostEntry(es.Spec.GetHost()), vips.MeshOutbound{
			Port:   es.Spec.GetPortUInt32(),
			TagSet: map[string]string{mesh_proto.ServiceTag: es.Spec.GetService()},
			Origin: "host",
		}))
	}

	if err != nil {
		return nil, err
	}
	return outboundSet, nil
}

func AllocateVIPs(global *vips.GlobalView, voView *vips.VirtualOutboundView) (errs error) {
	// Assign ips for all services
	for _, key := range voView.Keys() {
		vo := voView.Get(key)
		if vo.Address == "" {
			ip, err := global.Allocate(key)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				vo.Address = ip
			}
		}
	}
	return errs
}

func addDefault(outboundSet *vips.VirtualOutboundView, service string) error {
	return outboundSet.Add(vips.NewServiceEntry(service), vips.MeshOutbound{TagSet: map[string]string{mesh_proto.ServiceTag: service}, Origin: "default"})
}
