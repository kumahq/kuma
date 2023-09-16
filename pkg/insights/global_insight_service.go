package insights

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_types "github.com/kumahq/kuma/api/openapi/types"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
)

type GlobalInsightService interface {
	GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsight, error)
}

type defaultGlobalInsightService struct {
	resManager resources_manager.ResourceManager
}

var _ GlobalInsightService = &defaultGlobalInsightService{}

func NewDefaultGlobalInsightService(resManager resources_manager.ResourceManager) GlobalInsightService {
	return &defaultGlobalInsightService{resManager: resManager}
}

// TODO do we want to add some metrics, like computing time?
func (gis *defaultGlobalInsightService) GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsight, error) {
	globalInsights := &api_types.GlobalInsight{CreatedAt: core.Now()}

	meshInsights := &mesh.MeshInsightResourceList{}
	err := gis.resManager.List(ctx, meshInsights)
	if err != nil {
		return nil, err
	}

	globalInsights.Meshes.Total = len(meshInsights.GetItems())

	gis.aggregateDataplanes(meshInsights, globalInsights)
	gis.aggregatePolicies(meshInsights, globalInsights)

	err = gis.aggregateServices(ctx, globalInsights)
	if err != nil {
		return nil, err
	}

	err = gis.aggregateZoneControlPlanes(ctx, globalInsights)
	if err != nil {
		return nil, err
	}

	err = gis.aggregateZoneIngresses(ctx, globalInsights)
	if err != nil {
		return nil, err
	}

	err = gis.aggregateZoneEgresses(ctx, globalInsights)
	if err != nil {
		return nil, err
	}

	return globalInsights, nil
}

func (gis *defaultGlobalInsightService) aggregateDataplanes(meshInsights *mesh.MeshInsightResourceList, globalInsight *api_types.GlobalInsight) {
	for _, meshInsight := range meshInsights.GetItems() {
		spec := meshInsight.GetSpec().(*mesh_proto.MeshInsight)
		dataplanesByType := spec.GetDataplanesByType()
		globalInsight.Dataplanes.Standard.Online += int(dataplanesByType.GetStandard().GetOnline())
		globalInsight.Dataplanes.Standard.Offline += int(dataplanesByType.GetStandard().GetOffline())
		globalInsight.Dataplanes.Standard.PartiallyDegraded += int(dataplanesByType.GetStandard().GetPartiallyDegraded())
		globalInsight.Dataplanes.Standard.Total += int(dataplanesByType.GetStandard().GetTotal())

		globalInsight.Dataplanes.GatewayBuiltin.Online += int(dataplanesByType.GetGatewayBuiltin().GetOnline())
		globalInsight.Dataplanes.GatewayBuiltin.Offline += int(dataplanesByType.GetGatewayBuiltin().GetOffline())
		globalInsight.Dataplanes.GatewayBuiltin.PartiallyDegraded += int(dataplanesByType.GetGatewayBuiltin().GetPartiallyDegraded())
		globalInsight.Dataplanes.GatewayBuiltin.Total += int(dataplanesByType.GetGatewayBuiltin().GetTotal())

		globalInsight.Dataplanes.GatewayDelegated.Online += int(dataplanesByType.GetGatewayDelegated().GetOnline())
		globalInsight.Dataplanes.GatewayDelegated.Offline += int(dataplanesByType.GetGatewayDelegated().GetOffline())
		globalInsight.Dataplanes.GatewayDelegated.PartiallyDegraded += int(dataplanesByType.GetGatewayDelegated().GetPartiallyDegraded())
		globalInsight.Dataplanes.GatewayDelegated.Total += int(dataplanesByType.GetGatewayDelegated().GetTotal())
	}
}

func (gis *defaultGlobalInsightService) aggregatePolicies(meshInsights *mesh.MeshInsightResourceList, globalInsight *api_types.GlobalInsight) {
	for _, meshInsight := range meshInsights.GetItems() {
		spec := meshInsight.GetSpec().(*mesh_proto.MeshInsight)
		for _, policy := range spec.GetPolicies() {
			globalInsight.Policies.Total += int(policy.GetTotal())
		}
	}
}

func (gis *defaultGlobalInsightService) aggregateServices(ctx context.Context, globalInsight *api_types.GlobalInsight) error {
	serviceInsights := &mesh.ServiceInsightResourceList{}
	err := gis.resManager.List(ctx, serviceInsights)
	if err != nil {
		return err
	}

	for _, serviceInsight := range serviceInsights.GetItems() {
		spec := serviceInsight.GetSpec().(*mesh_proto.ServiceInsight)
		for _, service := range spec.GetServices() {
			switch service.GetServiceType() {
			case mesh_proto.ServiceInsight_Service_internal:
				updateServiceStatus(service.GetStatus(), &globalInsight.Services.Internal)
			case mesh_proto.ServiceInsight_Service_external:
				globalInsight.Services.External.Total += 1
			case mesh_proto.ServiceInsight_Service_gateway_builtin:
				updateServiceStatus(service.GetStatus(), &globalInsight.Services.GatewayBuiltin)
			case mesh_proto.ServiceInsight_Service_gateway_delegated:
				updateServiceStatus(service.GetStatus(), &globalInsight.Services.GatewayDelegated)
			}
		}
	}

	return nil
}

func updateServiceStatus(serviceStatus mesh_proto.ServiceInsight_Service_Status, status *api_types.FullStatus) {
	status.Total += 1
	switch serviceStatus {
	case mesh_proto.ServiceInsight_Service_online:
		status.Online += 1
	case mesh_proto.ServiceInsight_Service_offline:
		status.Offline += 1
	case mesh_proto.ServiceInsight_Service_partially_degraded:
		status.PartiallyDegraded += 1
	default:
		return
	}
}

func (gis *defaultGlobalInsightService) aggregateZoneControlPlanes(ctx context.Context, globalInsight *api_types.GlobalInsight) error {
	zoneInsights := &system.ZoneInsightResourceList{}
	err := gis.resManager.List(ctx, zoneInsights)
	if err != nil {
		return err
	}

	for _, zoneInsight := range zoneInsights.GetItems() {
		spec := zoneInsight.GetSpec().(*system_proto.ZoneInsight)
		for _, subscription := range spec.GetSubscriptions() {
			globalInsight.Zones.ControlPlanes.Total += 1
			if subscription.GetDisconnectTime() == nil || subscription.GetDisconnectTime().AsTime().Before(subscription.GetConnectTime().AsTime()) {
				globalInsight.Zones.ControlPlanes.Online += 1
			}
		}
	}

	return nil
}

func (gis *defaultGlobalInsightService) aggregateZoneIngresses(ctx context.Context, globalInsight *api_types.GlobalInsight) error {
	zoneIngressInsights := &mesh.ZoneIngressInsightResourceList{}
	err := gis.resManager.List(ctx, zoneIngressInsights)
	if err != nil {
		return err
	}

	for _, zoneIngressInsight := range zoneIngressInsights.GetItems() {
		spec := zoneIngressInsight.GetSpec().(*mesh_proto.ZoneIngressInsight)
		for _, subscription := range spec.GetSubscriptions() {
			globalInsight.Zones.ZoneIngresses.Total += 1
			if subscription.GetDisconnectTime() == nil {
				globalInsight.Zones.ZoneIngresses.Online += 1
			}
		}
	}

	return nil
}

func (gis *defaultGlobalInsightService) aggregateZoneEgresses(ctx context.Context, globalInsight *api_types.GlobalInsight) error {
	zoneEgressInsights := &mesh.ZoneEgressInsightResourceList{}
	err := gis.resManager.List(ctx, zoneEgressInsights)
	if err != nil {
		return err
	}

	for _, zoneEgressInsight := range zoneEgressInsights.GetItems() {
		spec := zoneEgressInsight.GetSpec().(*mesh_proto.ZoneEgressInsight)
		for _, subscription := range spec.GetSubscriptions() {
			globalInsight.Zones.ZoneEgresses.Total += 1
			if subscription.GetDisconnectTime() == nil {
				globalInsight.Zones.ZoneEgresses.Online += 1
			}
		}
	}

	return nil
}
