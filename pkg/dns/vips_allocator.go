package dns

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
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
	dnsSuffix         string
	zone              string
}

// NewVIPsAllocator creates new object of VIPsAllocator. You can either
// call method CreateOrUpdateVIPConfig manually or start VIPsAllocator as a component.
// In the latter scenario it will call CreateOrUpdateVIPConfig every 'tickInterval'
// for all meshes in the store.
func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, config dns_server.Config, zone string) (*VIPsAllocator, error) {
	return &VIPsAllocator{
		rm:                rm,
		persistence:       vips.NewPersistence(rm, configManager),
		serviceVipEnabled: config.ServiceVipEnabled,
		cidr:              config.CIDR,
		dnsSuffix:         config.Domain,
		zone:              zone,
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

	newView, err := d.BuildVirtualOutboundMeshView(ctx, mesh)
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

	newView, err := d.BuildVirtualOutboundMeshView(ctx, mesh)
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
	Log.Info("mesh VIPs changed", "mesh", mesh, "changes", changes)
	return d.persistence.Set(ctx, mesh, out)
}

func generatedHostname(meta model.ResourceMeta, suffix string) string {
	return fmt.Sprintf("internal.%s.%s.%s", meta.GetName(), meta.GetMesh(), suffix)
}

func addFromMeshGateway(outboundSet *vips.VirtualOutboundMeshView, dnsSuffix, mesh string, gateway *core_mesh.MeshGatewayResource) {
	for i, listener := range gateway.Spec.Conf.Listeners {
		// We only setup outbounds for cross mesh listeners and only ones with a
		// concrete hostname.
		if !listener.CrossMesh {
			continue
		}

		hostname := listener.Hostname

		if hostname == "" || strings.Contains(hostname, "*") {
			hostname = generatedHostname(gateway.GetMeta(), dnsSuffix)
		}

		// We only allow one selector with a crossMesh listener
		for _, selector := range gateway.Spec.Selectors {
			tags := mesh_proto.Merge(
				gateway.Spec.GetTags(),
				listener.GetTags(),
				map[string]string{
					mesh_proto.MeshTag: mesh,
				},
				selector.GetMatch(),
			)
			origin := fmt.Sprintf("mesh-gateway:%s:%s:%s", mesh, gateway.GetMeta().GetName(), hostname)

			entry := vips.OutboundEntry{
				Port:   listener.Port,
				TagSet: tags,
				Origin: origin,
			}
			if err := outboundSet.Add(vips.NewFqdnEntry(hostname), entry); err != nil {
				Log.WithValues("mesh", mesh, "gateway", gateway.GetMeta().GetName(), "listener", i).
					Info("failed to add MeshGateway-generated outbound", "reason", err.Error())
			}
		}
	}
}

func (d *VIPsAllocator) BuildVirtualOutboundMeshView(ctx context.Context, mesh string) (*vips.VirtualOutboundMeshView, error) {
	outboundSet := vips.NewEmptyVirtualOutboundView()

	virtualOutbounds := core_mesh.VirtualOutboundResourceList{}
	if err := d.rm.List(ctx, &virtualOutbounds, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	dataplanes := core_mesh.DataplaneResourceList{}
	if err := d.rm.List(ctx, &dataplanes, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	var errs error
	for _, dp := range dataplanes.Items {
		for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
			if d.serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, inbound.GetService(), 0))
			}
			for _, vob := range Match(virtualOutbounds.Items, inbound.Tags) {
				addFromVirtualOutbound(outboundSet, vob, inbound.Tags, dp.Descriptor().Name, dp.Meta.GetName())
			}
		}
	}

	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := d.rm.List(ctx, &zoneIngresses); err != nil {
		return nil, err
	}

	for _, zi := range zoneIngresses.Items {
		for _, service := range zi.Spec.GetAvailableServices() {
			if !zi.IsRemoteIngress(d.zone) {
				continue
			}
			if service.Mesh == mesh && d.serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, service.GetTags()[mesh_proto.ServiceTag], 0))
			}
			for _, vob := range Match(virtualOutbounds.Items, service.Tags) {
				addFromVirtualOutbound(outboundSet, vob, service.Tags, zi.Descriptor().Name, zi.Meta.GetName())
			}
		}
	}

	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := d.rm.List(ctx, &externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	// TODO(lukidzi): after switching to use the resource store as in the code for tests
	// we should switch to `ListOrdered` https://github.com/kumahq/kuma/issues/7356
	sort.SliceStable(externalServices.Items, func(i, j int) bool {
		return (externalServices.Items[i].GetMeta().GetName() < externalServices.Items[j].GetMeta().GetName())
	})

	for _, es := range externalServices.Items {
		tags := map[string]string{mesh_proto.ServiceTag: es.Spec.GetService()}
		if d.serviceVipEnabled {
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

	meshList := core_mesh.MeshResourceList{}
	if err := d.rm.List(ctx, &meshList); err != nil {
		return nil, err
	}

	for _, mesh := range meshList.Items {
		meshName := mesh.GetMeta().GetName()
		gateways := core_mesh.MeshGatewayResourceList{}
		if err := d.rm.List(ctx, &gateways, store.ListByMesh(meshName)); err != nil {
			return nil, err
		}

		for _, gateway := range gateways.Items {
			addFromMeshGateway(outboundSet, d.dnsSuffix, meshName, gateway)
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
