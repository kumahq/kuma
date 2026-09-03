package globalinsight

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

type GlobalInsightService interface {
	GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsightBase, error)
}

type defaultGlobalInsightService struct {
	resourceStore             core_store.ResourceStore
	globalResourceDescriptors []core_model.ResourceTypeDescriptor
}

var _ GlobalInsightService = &defaultGlobalInsightService{}

func NewDefaultGlobalInsightService(resourceStore core_store.ResourceStore) GlobalInsightService {
	return &defaultGlobalInsightService{
		resourceStore:             resourceStore,
		globalResourceDescriptors: globalResourceDescriptors(),
	}
}

func (gis *defaultGlobalInsightService) GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsightBase, error) {
	globalInsights := &api_types.GlobalInsightBase{CreatedAt: core.Now()}

	meshInsights := &mesh.MeshInsightResourceList{}
	if err := gis.resourceStore.List(ctx, meshInsights); err != nil {
		return nil, err
	}

	globalInsights.Meshes.Total = len(meshInsights.GetItems())

	gis.aggregateDataplanes(meshInsights, globalInsights)
	gis.aggregatePolicies(meshInsights, globalInsights)
	if err := gis.aggregateResources(ctx, meshInsights, globalInsights); err != nil {
		return nil, err
	}

	if err := gis.aggregateServices(ctx, globalInsights); err != nil {
		return nil, err
	}

	if err := gis.aggregateZoneControlPlanes(ctx, globalInsights); err != nil {
		return nil, err
	}

	// zones.zoneIngresses and zones.zoneEgresses stay at zero: they were fed by
	// ZoneIngressInsight/ZoneEgressInsight, and zone proxies are now Dataplanes
	// with listeners, which MeshInsight does not break out yet.
	return globalInsights, nil
}

func (gis *defaultGlobalInsightService) aggregateDataplanes(
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsightBase,
) {
	for _, meshInsight := range meshInsights.GetItems() {
		dataplanesByType := meshInsight.GetSpec().(*mesh_proto.MeshInsight).GetDataplanesByType()

		standard := dataplanesByType.GetStandard()
		globalInsight.Dataplanes.Standard.Online += int(standard.GetOnline())
		globalInsight.Dataplanes.Standard.Offline += int(standard.GetOffline())
		globalInsight.Dataplanes.Standard.PartiallyDegraded += int(standard.GetPartiallyDegraded())
		globalInsight.Dataplanes.Standard.Total += int(standard.GetTotal())

		gatewayDelegated := dataplanesByType.GetGatewayDelegated()
		globalInsight.Dataplanes.GatewayDelegated.Online += int(gatewayDelegated.GetOnline())
		globalInsight.Dataplanes.GatewayDelegated.Offline += int(gatewayDelegated.GetOffline())
		globalInsight.Dataplanes.GatewayDelegated.PartiallyDegraded += int(gatewayDelegated.GetPartiallyDegraded())
		globalInsight.Dataplanes.GatewayDelegated.Total += int(gatewayDelegated.GetTotal())
	}
}

func (gis *defaultGlobalInsightService) aggregatePolicies(
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsightBase,
) {
	for _, meshInsight := range meshInsights.GetItems() {
		resources := meshInsight.GetSpec().(*mesh_proto.MeshInsight).GetResources()

		for resType, stat := range resources {
			desc, err := registry.Global().DescriptorFor(core_model.ResourceType(resType))
			if err != nil || !desc.IsPolicy {
				continue
			}
			globalInsight.Policies.Total += int(stat.GetTotal())
		}
	}
}

func (gis *defaultGlobalInsightService) aggregateResources(
	ctx context.Context,
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsightBase,
) error {
	globalInsight.Resources = map[string]api_types.ResourceStats{}
	for _, meshInsight := range meshInsights.GetItems() {
		for resName, resStat := range meshInsight.GetSpec().(*mesh_proto.MeshInsight).Resources {
			addResourceTotal(globalInsight.Resources, resName, int(resStat.Total))
		}
	}

	for _, descriptor := range gis.globalResourceDescriptors {
		list := descriptor.NewList()
		if err := gis.resourceStore.List(ctx, list); err != nil {
			return err
		}

		count := len(list.GetItems())
		if count == 0 {
			continue
		}

		globalInsight.Resources[string(descriptor.Name)] = api_types.ResourceStats{
			Total: count,
		}
	}

	return nil
}

func globalResourceDescriptors() []core_model.ResourceTypeDescriptor {
	return registry.Global().ObjectDescriptors(
		core_model.HasScope(core_model.ScopeGlobal),
		core_model.HasWsEnabled(),
		core_model.TypeFilterFn(func(descriptor core_model.ResourceTypeDescriptor) bool {
			return !descriptor.AdminOnly
		}),
		core_model.Not(core_model.IsInsight()),
		core_model.Not(core_model.Named(
			mesh.MeshType,
			system.ZoneType,
			system.ConfigType,
		)),
	)
}

func addResourceTotal(resources map[string]api_types.ResourceStats, resourceName string, total int) {
	stats := resources[resourceName]
	stats.Total += total
	resources[resourceName] = stats
}

// aggregateServices counts internal services from MeshService, external services
// from MeshExternalService and gateway services from gateway Dataplanes.
func (gis *defaultGlobalInsightService) aggregateServices(
	ctx context.Context,
	globalInsight *api_types.GlobalInsightBase,
) error {
	meshServices := &meshservice_api.MeshServiceResourceList{}
	if err := gis.resourceStore.List(ctx, meshServices); err != nil {
		return err
	}
	for _, meshService := range meshServices.Items {
		proxies := meshService.Status.DataplaneProxies
		// A proxy is only serving traffic when it is both connected to the control plane
		// and has all inbounds selected by this service ready. Connected and Healthy are
		// independent counters, so their minimum is the tightest bound on that overlap.
		updateServiceStatus(min(proxies.Connected, proxies.Healthy), proxies.Total, &globalInsight.Services.Internal)
	}

	externalServices := &meshexternalservice_api.MeshExternalServiceResourceList{}
	if err := gis.resourceStore.List(ctx, externalServices); err != nil {
		return err
	}
	globalInsight.Services.External.Total = len(externalServices.GetItems())

	return gis.aggregateGatewayServices(ctx, globalInsight)
}

type gatewayServiceStat struct {
	online int
	total  int
}

// aggregateGatewayServices counts delegated gateway services from gateway
// Dataplanes grouped by the service they belong to. Gateway Dataplanes are
// never turned into a MeshService, so they are the only source for these stats.
func (gis *defaultGlobalInsightService) aggregateGatewayServices(
	ctx context.Context,
	globalInsight *api_types.GlobalInsightBase,
) error {
	dataplanes := &mesh.DataplaneResourceList{}
	if err := gis.resourceStore.List(ctx, dataplanes); err != nil {
		return err
	}

	hasGateway := false
	for _, dp := range dataplanes.Items {
		if dp.Spec.GetNetworking().GetGateway() != nil {
			hasGateway = true
			break
		}
	}
	if !hasGateway {
		return nil
	}

	dataplaneInsights := &mesh.DataplaneInsightResourceList{}
	if err := gis.resourceStore.List(ctx, dataplaneInsights); err != nil {
		return err
	}

	delegated := map[string]*gatewayServiceStat{}

	for _, overview := range mesh.NewDataplaneOverviews(*dataplanes, *dataplaneInsights).Items {
		if overview.Spec.GetDataplane().GetNetworking().GetGateway() == nil {
			continue
		}

		key := gatewayServiceKey(overview)
		stat, ok := delegated[key]
		if !ok {
			stat = &gatewayServiceStat{}
			delegated[key] = stat
		}
		stat.total++
		if status, _ := overview.Status(); status == mesh.Online {
			stat.online++
		}
	}

	for _, stat := range delegated {
		updateServiceStatus(stat.online, stat.total, &globalInsight.Services.GatewayDelegated)
	}

	return nil
}

// gatewayServiceKey identifies the mesh-scoped service a gateway Dataplane belongs
// to. Gateways carry no kuma.io/service tag in tag-free mode, so fall back to the
// workload and finally to the Dataplane itself, which keeps every gateway counted.
// Each tier is prefixed so that gateways grouped by different tiers never share a key.
func gatewayServiceKey(overview *mesh.DataplaneOverviewResource) string {
	key := "tag:" + overview.Spec.GetDataplane().GetNetworking().GetGateway().GetTags()[mesh_proto.ServiceTag]
	if key == "tag:" {
		key = "workload:" + overview.GetMeta().GetLabels()[metadata.KumaWorkload]
	}
	if key == "workload:" {
		key = "dataplane:" + overview.GetMeta().GetName()
	}
	return overview.GetMeta().GetMesh() + "/" + key
}

func updateServiceStatus(online, total int, status *api_types.FullStatus) {
	status.Total += 1
	switch {
	case total == 0 || online == 0:
		status.Offline += 1
	case online == total:
		status.Online += 1
	default:
		status.PartiallyDegraded += 1
	}
}

func (gis *defaultGlobalInsightService) aggregateZoneControlPlanes(
	ctx context.Context,
	globalInsight *api_types.GlobalInsightBase,
) error {
	zoneInsights := &system.ZoneInsightResourceList{}
	if err := gis.resourceStore.List(ctx, zoneInsights); err != nil {
		return err
	}

	for _, zoneInsight := range zoneInsights.GetItems() {
		globalInsight.Zones.ControlPlanes.Total += 1
		if zoneInsight.GetSpec().(*system_proto.ZoneInsight).IsOnline() {
			globalInsight.Zones.ControlPlanes.Online += 1
		}
	}

	return nil
}
