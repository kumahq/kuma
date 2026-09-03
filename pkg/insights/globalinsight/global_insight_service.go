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

	return globalInsights, nil
}

func (gis *defaultGlobalInsightService) aggregateDataplanes(
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsightBase,
) {
	for _, meshInsight := range meshInsights.GetItems() {
		dataplanes := meshInsight.GetSpec().(*mesh_proto.MeshInsight).GetDataplanes()

		globalInsight.Dataplanes.Online += int(dataplanes.GetOnline())
		globalInsight.Dataplanes.Offline += int(dataplanes.GetOffline())
		globalInsight.Dataplanes.PartiallyDegraded += int(dataplanes.GetPartiallyDegraded())
		globalInsight.Dataplanes.Total += int(dataplanes.GetTotal())
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

// aggregateServices counts internal services from MeshService and external
// services from MeshExternalService.
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

	return nil
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
