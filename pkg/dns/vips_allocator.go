package dns

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/metadata"
	dns_metrics "github.com/kumahq/kuma/pkg/dns/metrics"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var Log = core.Log.WithName("dns-vips-allocator")

type VIPsAllocator struct {
	rm                manager.ReadOnlyResourceManager
	persistence       *vips.Persistence
	cidr              string
	serviceVipEnabled bool
	dnsSuffix         string
	zone              string
	metrics           *dns_metrics.Metrics
}

// NewVIPsAllocator creates new object of VIPsAllocator. You can either
// call method CreateOrUpdateVIPConfig manually or start VIPsAllocator as a component.
// In the latter scenario it will call CreateOrUpdateVIPConfig every 'tickInterval'
// for all meshes in the store.
func NewVIPsAllocator(rm manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, config dns_server.Config, experimentalConfig config.ExperimentalConfig, zone string, metrics core_metrics.Metrics) (*VIPsAllocator, error) {
	dnsMetrics, err := dns_metrics.NewMetrics(metrics)
	if err != nil {
		return nil, err
	}

	return &VIPsAllocator{
		rm:                rm,
		persistence:       vips.NewPersistence(rm, configManager, experimentalConfig.UseTagFirstVirtualOutboundModel),
		serviceVipEnabled: config.ServiceVipEnabled,
		cidr:              config.CIDR,
		dnsSuffix:         config.Domain,
		zone:              zone,
		metrics:           dnsMetrics,
	}, nil
}

func (d *VIPsAllocator) CreateOrUpdateVIPConfigs(ctx context.Context) error {
	start := core.Now()

	if err := d.createOrUpdateVIPConfigs(ctx); err != nil {
		d.metrics.VipGenerationsErrors.Inc()
		return err
	}

	d.metrics.VipGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
	return nil
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

func (d *VIPsAllocator) createOrUpdateVIPConfigs(ctx context.Context) error {
	meshesRes := core_mesh.MeshResourceList{}
	err := d.rm.List(ctx, &meshesRes)
	if err != nil {
		return err
	}

	meshesNames := []string{}
	for _, meshRes := range meshesRes.Items {
		meshesNames = append(meshesNames, meshRes.GetMeta().GetName())
	}

	oldViewByMesh, globalViewByMesh, err := d.fetchViewByMesh(ctx, meshesNames)
	if err != nil {
		return err
	}

	zoneIngresses, err := d.fetchZoneIngresses(ctx)
	if err != nil {
		return err
	}

	dataplanesByMesh, err := d.fetchDataplanesByMesh(ctx)
	if err != nil {
		return err
	}

	virtualOutboundsByMesh, err := d.fetchVirtualOutboundsByMesh(ctx)
	if err != nil {
		return err
	}

	externalServicesByMesh, err := d.fetchExternalServicesByMesh(ctx)
	if err != nil {
		return err
	}

	meshGatewaysByMesh, err := d.fetchMeshGatewaysByMesh(ctx)
	if err != nil {
		return err
	}

	var errs error
	for _, mesh := range meshesNames {
		newView, err := d.buildVirtualOutboundMeshView(mesh, virtualOutboundsByMesh[mesh], dataplanesByMesh[mesh], zoneIngresses, externalServicesByMesh[mesh], meshGatewaysByMesh)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		oldView, globalView := oldViewByMesh[mesh], globalViewByMesh[mesh]
		err = d.createOrUpdateMeshVIPConfig(ctx, mesh, oldView, newView, globalView)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func (d *VIPsAllocator) fetchViewByMesh(ctx context.Context, meshes []string) (map[string]*vips.VirtualOutboundMeshView, map[string]*vips.GlobalView, error) {
	meshViewByMesh, err := d.persistence.Get(ctx, meshes)
	if err != nil {
		return nil, nil, err
	}

	globalViewByMesh := map[string]*vips.GlobalView{}

	for mesh, meshView := range meshViewByMesh {
		globalView, err := vips.NewGlobalView(d.cidr)
		if err != nil {
			return nil, nil, err
		}

		err = d.updateView(globalView, meshView)
		if err != nil {
			return nil, nil, err
		}

		globalViewByMesh[mesh] = globalView
	}

	return meshViewByMesh, globalViewByMesh, nil
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

	err = d.updateView(gv, meshView)
	if err != nil {
		return nil, nil, err
	}

	return meshView, gv, nil
}

func (d *VIPsAllocator) updateView(gv *vips.GlobalView, meshView *vips.VirtualOutboundMeshView) error {
	for _, hostEntry := range meshView.HostnameEntries() {
		if hostEntry.Type == vips.Host && net.ParseIP(hostEntry.Name) != nil {
			continue
		}
		vo := meshView.Get(hostEntry)
		if err := gv.Reserve(hostEntry, vo.Address); err != nil {
			return err
		}
	}
	return nil
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
			origin := vips.OriginGateway(mesh, gateway.GetMeta().GetName(), hostname)
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

func (d *VIPsAllocator) fetchZoneIngresses(ctx context.Context) ([]*core_mesh.ZoneIngressResource, error) {
	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := d.rm.List(ctx, &zoneIngresses); err != nil {
		return nil, err
	}

	return zoneIngresses.Items, nil
}

func (d *VIPsAllocator) fetchDataplanesByMesh(ctx context.Context) (map[string][]*core_mesh.DataplaneResource, error) {
	dataplanes := core_mesh.DataplaneResourceList{}
	if err := d.rm.List(ctx, &dataplanes); err != nil {
		return nil, err
	}

	out := map[string][]*core_mesh.DataplaneResource{}
	for _, dataplane := range dataplanes.Items {
		out[dataplane.Meta.GetMesh()] = append(out[dataplane.Meta.GetMesh()], dataplane)
	}

	return out, nil
}

func (d *VIPsAllocator) fetchVirtualOutboundsByMesh(ctx context.Context) (map[string][]*core_mesh.VirtualOutboundResource, error) {
	virtualOutbounds := core_mesh.VirtualOutboundResourceList{}
	if err := d.rm.List(ctx, &virtualOutbounds); err != nil {
		return nil, err
	}

	out := map[string][]*core_mesh.VirtualOutboundResource{}
	for _, virtualOutbound := range virtualOutbounds.Items {
		mesh := virtualOutbound.Meta.GetMesh()
		out[mesh] = append(out[mesh], virtualOutbound)
	}

	return out, nil
}

func (d *VIPsAllocator) fetchExternalServicesByMesh(ctx context.Context) (map[string][]*core_mesh.ExternalServiceResource, error) {
	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := d.rm.List(ctx, &externalServices); err != nil {
		return nil, err
	}

	out := map[string][]*core_mesh.ExternalServiceResource{}
	for _, externalService := range externalServices.Items {
		mesh := externalService.Meta.GetMesh()
		out[mesh] = append(out[mesh], externalService)
	}

	return out, nil
}

func (d *VIPsAllocator) fetchMeshGatewaysByMesh(ctx context.Context) (map[string][]*core_mesh.MeshGatewayResource, error) {
	meshGateways := core_mesh.MeshGatewayResourceList{}
	if err := d.rm.List(ctx, &meshGateways); err != nil {
		return nil, err
	}

	out := map[string][]*core_mesh.MeshGatewayResource{}
	for _, meshGateway := range meshGateways.Items {
		mesh := meshGateway.Meta.GetMesh()
		out[mesh] = append(out[mesh], meshGateway)
	}

	return out, nil
}

func (d *VIPsAllocator) buildVirtualOutboundMeshView(
	mesh string,
	virtualOutbounds []*core_mesh.VirtualOutboundResource,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	meshGatewaysByMesh map[string][]*core_mesh.MeshGatewayResource,
) (*vips.VirtualOutboundMeshView, error) {
	outboundSet := vips.NewEmptyVirtualOutboundView()

	var errs error
	for _, dp := range dataplanes {
		for _, inbound := range dp.Spec.GetNetworking().GetInbound() {
			if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
				continue
			}
			if inbound.Port == mesh_proto.TCPPortReserved {
				continue
			}
			if d.serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, inbound.GetService(), 0))
			}
			for _, vob := range Match(virtualOutbounds, inbound.Tags) {
				addFromVirtualOutbound(outboundSet, vob, inbound.Tags, dp.Descriptor().Name, dp.Meta.GetName())
			}
		}
	}

	for _, zi := range zoneIngresses {
		for _, service := range zi.Spec.GetAvailableServices() {
			if !zi.IsRemoteIngress(d.zone) {
				continue
			}
			if service.Mesh == mesh && d.serviceVipEnabled {
				errs = multierr.Append(errs, addDefault(outboundSet, service.GetTags()[mesh_proto.ServiceTag], 0))
			}
			for _, vob := range Match(virtualOutbounds, service.Tags) {
				addFromVirtualOutbound(outboundSet, vob, service.Tags, zi.Descriptor().Name, zi.Meta.GetName())
			}
		}
	}

	// TODO(lukidzi): after switching to use the resource store as in the code for tests
	// we should switch to `ListOrdered` https://github.com/kumahq/kuma/issues/7356
	sort.SliceStable(externalServices, func(i, j int) bool {
		return (externalServices[i].GetMeta().GetName() < externalServices[j].GetMeta().GetName())
	})

	for _, es := range externalServices {
		tags := map[string]string{mesh_proto.ServiceTag: es.Spec.GetService()}
		if d.serviceVipEnabled {
			errs = multierr.Append(errs, addDefault(outboundSet, es.Spec.GetService(), es.Spec.GetPortUInt32()))
		}
		if !es.Spec.Networking.DisableHostDNSEntry {
			addError := outboundSet.Add(vips.NewHostEntry(es.Spec.GetHost()), vips.OutboundEntry{
				Port:   es.Spec.GetPortUInt32(),
				TagSet: tags,
				Origin: vips.OriginHost(es.GetMeta().GetName()),
			})
			if addError != nil {
				errs = multierr.Append(errs, errors.Wrapf(addError, "cannot add outbound for external service '%s'", es.GetMeta().GetName()))
			}
		}
		for _, vob := range Match(virtualOutbounds, tags) {
			addFromVirtualOutbound(outboundSet, vob, tags, es.Descriptor().Name, es.Meta.GetName())
		}
	}

	for mesh, gateways := range meshGatewaysByMesh {
		for _, gateway := range gateways {
			addFromMeshGateway(outboundSet, d.dnsSuffix, mesh, gateway)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return outboundSet, nil
}

func (d *VIPsAllocator) BuildVirtualOutboundMeshView(ctx context.Context, mesh string) (*vips.VirtualOutboundMeshView, error) {
	virtualOutbounds := core_mesh.VirtualOutboundResourceList{}
	if err := d.rm.List(ctx, &virtualOutbounds, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	dataplanes := core_mesh.DataplaneResourceList{}
	if err := d.rm.List(ctx, &dataplanes, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	zoneIngresses := core_mesh.ZoneIngressResourceList{}
	if err := d.rm.List(ctx, &zoneIngresses); err != nil {
		return nil, err
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

	meshGatewaysByMesh, err := d.fetchMeshGatewaysByMesh(ctx)
	if err != nil {
		return nil, err
	}

	return d.buildVirtualOutboundMeshView(mesh, virtualOutbounds.Items, dataplanes.Items, zoneIngresses.Items, externalServices.Items, meshGatewaysByMesh)
}

func AllocateVIPs(global *vips.GlobalView, voView *vips.VirtualOutboundMeshView) error {
	var errs error
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
		Origin: string(metadata.OriginService),
		Port:   port,
	})
}
