package dns

import (
	"context"
	"net"

	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

var Log = core.Log.WithName("dns-vips-allocator")

type VIPsAllocator struct {
	rm                manager.ReadOnlyResourceManager
	persistence       *vips.Persistence
	cidr              string
	serviceVipEnabled bool
}

// NewVIPsAllocator creates new object of VIPsAllocator. You can either
// call method CreateOrUpdateVIPConfig manually or start VIPsAllocator as a component.
// In the latter scenario it will call CreateOrUpdateVIPConfig every 'tickInterval'
// for all meshes in the store.
func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, serviceVipEnabled bool, cidr string) (*VIPsAllocator, error) {
	return &VIPsAllocator{
		rm:                rm,
		persistence:       vips.NewPersistence(rm, configManager),
		serviceVipEnabled: serviceVipEnabled,
		cidr:              cidr,
	}, nil
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfigs(ctx context.Context) error {
	meshRes := core_mesh.MeshResourceList{}
	if err := d.rm.List(ctx, &meshRes); err != nil {
		return err
	}

	var errs error
	for _, mesh := range meshRes.Items {
		if err := d.createOrUpdateVIPConfigs(ctx, mesh.GetMeta().GetName()); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfig(ctx context.Context, mesh string, viewModificator func(*vips.VirtualOutboundMeshView) error) error {
	oldView, globalView, err := d.fetchView(ctx, mesh)
	if err != nil {
		return err
	}

	newView, err := BuildVirtualOutboundMeshView(ctx, d.rm, d.serviceVipEnabled, mesh)
	if err != nil {
		return err
	}
	if err := viewModificator(newView); err != nil {
		return err
	}

	if err := d.createOrUpdateMeshVIPConfig(ctx, mesh, oldView, newView, globalView); err != nil {
		return err
	}
	return nil
}

func (d *VIPsAllocator) createOrUpdateVIPConfigs(ctx context.Context, mesh string) (err error) {
	oldView, globalView, err := d.fetchView(ctx, mesh)
	if err != nil {
		return err
	}

	newView, err := BuildVirtualOutboundMeshView(ctx, d.rm, d.serviceVipEnabled, mesh)
	if err != nil {
		return err
	}
	return d.createOrUpdateMeshVIPConfig(ctx, mesh, oldView, newView, globalView)
}

func (d *VIPsAllocator) fetchView(ctx context.Context, mesh string) (*vips.VirtualOutboundMeshView, *vips.GlobalView, error) {
	meshView, err := d.persistence.GetByMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}

	gv, err := vips.NewGlobalView(d.cidr)
	if err != nil {
		return nil, nil, err
	}
	for _, hostEntry := range meshView.HostnameEntries() {
		if hostEntry.Type == vips.Host && net.ParseIP(hostEntry.Name) != nil {
			continue
		}
		vo := meshView.Get(hostEntry)
		if err := gv.Reserve(hostEntry, vo.Address); err != nil {
			return nil, nil, err
		}
	}
	return meshView, gv, nil
}

func (d *VIPsAllocator) createOrUpdateMeshVIPConfig(
	ctx context.Context,
	mesh string,
	oldView *vips.VirtualOutboundMeshView,
	newView *vips.VirtualOutboundMeshView,
	globalView *vips.GlobalView,
) error {
	if err := AllocateVIPs(globalView, newView); err != nil {
		// Error might occur only if we run out of VIPs. There is no point to pass it through,
		// we must notify user in logs and proceed
		Log.Error(err, "failed to allocate new VIPs", "mesh", mesh)
	}
	changes, out := oldView.Update(newView)
	if len(changes) == 0 {
		return nil
	}
	Log.Info("mesh vip changes", "mesh", mesh, "changes", changes)
	return d.persistence.Set(ctx, mesh, out)
}

func BuildVirtualOutboundMeshView(ctx context.Context, rm manager.ReadOnlyResourceManager, serviceVipEnabled bool, mesh string) (*vips.VirtualOutboundMeshView, error) {
	outboundSet := vips.NewEmptyVirtualOutboundView()

	virtualOutbounds := core_mesh.VirtualOutboundResourceList{}
	if err := rm.List(ctx, &virtualOutbounds, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	dataplanes := core_mesh.DataplaneResourceList{}
	if err := rm.List(ctx, &dataplanes, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	var errs error
	for _, dp := range dataplanes.Items {
		for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
			if serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, inbound.GetService(), 0))
			}
			for _, vob := range Match(virtualOutbounds.Items, inbound.Tags) {
				addFromVirtualOutbound(outboundSet, vob, inbound.Tags, dp.Descriptor().Name, dp.Meta.GetName())
			}
		}
	}

	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := rm.List(ctx, &zoneIngresses); err != nil {
		return nil, err
	}

	for _, zi := range zoneIngresses.Items {
		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh == mesh && serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, service.GetTags()[mesh_proto.ServiceTag], 0))
			}
			for _, vob := range Match(virtualOutbounds.Items, service.Tags) {
				addFromVirtualOutbound(outboundSet, vob, service.Tags, zi.Descriptor().Name, zi.Meta.GetName())
			}
		}
	}

	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := rm.List(ctx, &externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	for _, es := range externalServices.Items {
		tags := map[string]string{mesh_proto.ServiceTag: es.Spec.GetService()}
		if serviceVipEnabled {
			errs = multierr.Append(errs, addDefault(outboundSet, es.Spec.GetService(), es.Spec.GetPortUInt32()))
		}
		errs = multierr.Append(errs, outboundSet.Add(vips.NewHostEntry(es.Spec.GetHost()), vips.OutboundEntry{
			Port:   es.Spec.GetPortUInt32(),
			TagSet: tags,
			Origin: vips.OriginHost,
		}))
		for _, vob := range Match(virtualOutbounds.Items, tags) {
			addFromVirtualOutbound(outboundSet, vob, tags, es.Descriptor().Name, es.Meta.GetName())
		}
	}

	if errs != nil {
		return nil, errs
	}
	return outboundSet, nil
}

func AllocateVIPs(global *vips.GlobalView, voView *vips.VirtualOutboundMeshView) (errs error) {
	// Assign ips for all services
	for _, hostnameEntry := range voView.HostnameEntries() {
		vo := voView.Get(hostnameEntry)
		if vo.Address == "" {
			ip, err := global.Allocate(hostnameEntry)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				vo.Address = ip
			}
		}
	}
	return errs
}

func addFromVirtualOutbound(outboundSet *vips.VirtualOutboundMeshView, vob *core_mesh.VirtualOutboundResource, tags map[string]string, resourceType model.ResourceType, resourceName string) {
	host, err := vob.EvalHost(tags)
	l := Log.WithValues("mesh", vob.Meta.GetMesh(), "virtualOutboundName", vob.Meta.GetName(), "type", resourceType, "name", resourceName, "tags", tags)
	if err != nil {
		l.Info("Failed evaluating host template", "reason", err.Error())
		return
	}

	port, err := vob.EvalPort(tags)
	if err != nil {
		l.Info("Failed evaluating port template", "reason", err.Error())
		return
	}

	err = outboundSet.Add(vips.NewFqdnEntry(host), vips.OutboundEntry{
		Port:   port,
		TagSet: vob.FilterTags(tags),
		Origin: vips.OriginVirtualOutbound(vob.Meta.GetName()),
	})
	if err != nil {
		l.Info("Failed adding generated outbound", "reason", err.Error())
	}
}

func addDefault(outboundSet *vips.VirtualOutboundMeshView, service string, port uint32) error {
	return outboundSet.Add(vips.NewServiceEntry(service), vips.OutboundEntry{
		TagSet: map[string]string{mesh_proto.ServiceTag: service},
		Origin: vips.OriginService,
		Port:   port,
	})
}
