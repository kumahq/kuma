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

func (gis *defaultGlobalInsightService) GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsight, error) {
	globalInsights := &api_types.GlobalInsight{CreatedAt: core.Now()}

	meshInsights := &mesh.MeshInsightResourceList{}
	if err := gis.resManager.List(ctx, meshInsights); err != nil {
		return nil, err
	}

	globalInsights.Meshes.Total = len(meshInsights.GetItems())

	gis.aggregateDataplanes(meshInsights, globalInsights)
	gis.aggregatePolicies(meshInsights, globalInsights)

	if err := gis.aggregateServices(ctx, globalInsights); err != nil {
		return nil, err
	}

	if err := gis.aggregateZoneControlPlanes(ctx, globalInsights); err != nil {
		return nil, err
	}

	if err := gis.aggregateZoneIngresses(ctx, globalInsights); err != nil {
		return nil, err
	}

	if err := gis.aggregateZoneEgresses(ctx, globalInsights); err != nil {
		return nil, err
	}

	return globalInsights, nil
}

func (gis *defaultGlobalInsightService) aggregateDataplanes(
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsight,
) {
	for _, meshInsight := range meshInsights.GetItems() {
		dataplanesByType := meshInsight.GetSpec().(*mesh_proto.MeshInsight).GetDataplanesByType()

		standard := dataplanesByType.GetStandard()
		globalInsight.Dataplanes.Standard.Online += int(standard.GetOnline())
		globalInsight.Dataplanes.Standard.Offline += int(standard.GetOffline())
		globalInsight.Dataplanes.Standard.PartiallyDegraded += int(standard.GetPartiallyDegraded())
		globalInsight.Dataplanes.Standard.Total += int(standard.GetTotal())

		gatewayBuiltin := dataplanesByType.GetGatewayBuiltin()
		globalInsight.Dataplanes.GatewayBuiltin.Online += int(gatewayBuiltin.GetOnline())
		globalInsight.Dataplanes.GatewayBuiltin.Offline += int(gatewayBuiltin.GetOffline())
		globalInsight.Dataplanes.GatewayBuiltin.PartiallyDegraded += int(gatewayBuiltin.GetPartiallyDegraded())
		globalInsight.Dataplanes.GatewayBuiltin.Total += int(gatewayBuiltin.GetTotal())

		gatewayDelegated := dataplanesByType.GetGatewayDelegated()
		globalInsight.Dataplanes.GatewayDelegated.Online += int(gatewayDelegated.GetOnline())
		globalInsight.Dataplanes.GatewayDelegated.Offline += int(gatewayDelegated.GetOffline())
		globalInsight.Dataplanes.GatewayDelegated.PartiallyDegraded += int(gatewayDelegated.GetPartiallyDegraded())
		globalInsight.Dataplanes.GatewayDelegated.Total += int(gatewayDelegated.GetTotal())
	}
}

func (gis *defaultGlobalInsightService) aggregatePolicies(
	meshInsights *mesh.MeshInsightResourceList,
	globalInsight *api_types.GlobalInsight,
) {
	for _, meshInsight := range meshInsights.GetItems() {
		policies := meshInsight.GetSpec().(*mesh_proto.MeshInsight).GetPolicies()

		for _, policy := range policies {
			globalInsight.Policies.Total += int(policy.GetTotal())
		}
	}
}

func (gis *defaultGlobalInsightService) aggregateServices(
	ctx context.Context,
	globalInsight *api_types.GlobalInsight,
) error {
	serviceInsights := &mesh.ServiceInsightResourceList{}
	if err := gis.resManager.List(ctx, serviceInsights); err != nil {
		return err
	}

	for _, serviceInsight := range serviceInsights.GetItems() {
		services := serviceInsight.GetSpec().(*mesh_proto.ServiceInsight).GetServices()

		for _, service := range services {
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

func (gis *defaultGlobalInsightService) aggregateZoneControlPlanes(
	ctx context.Context,
	globalInsight *api_types.GlobalInsight,
) error {
	zoneInsights := &system.ZoneInsightResourceList{}
	if err := gis.resManager.List(ctx, zoneInsights); err != nil {
		return err
	}

	for _, zoneInsight := range zoneInsights.GetItems() {
		subscriptions := zoneInsight.GetSpec().(*system_proto.ZoneInsight).GetSubscriptions()

		for _, subscription := range subscriptions {
			globalInsight.Zones.ControlPlanes.Total += 1

			disconnected := subscription.GetDisconnectTime()
			connected := subscription.GetConnectTime().AsTime()

			if disconnected == nil || disconnected.AsTime().Before(connected) {
				globalInsight.Zones.ControlPlanes.Online += 1
			}
		}
	}

	return nil
}

func (gis *defaultGlobalInsightService) aggregateZoneIngresses(
	ctx context.Context,
	globalInsight *api_types.GlobalInsight,
) error {
	zoneIngressInsights := &mesh.ZoneIngressInsightResourceList{}
	if err := gis.resManager.List(ctx, zoneIngressInsights); err != nil {
		return err
	}

	for _, zoneIngressInsight := range zoneIngressInsights.GetItems() {
		subscriptions := zoneIngressInsight.GetSpec().(*mesh_proto.ZoneIngressInsight).GetSubscriptions()

		for _, subscription := range subscriptions {
			globalInsight.Zones.ZoneIngresses.Total += 1

			if subscription.GetDisconnectTime() == nil {
				globalInsight.Zones.ZoneIngresses.Online += 1
			}
		}
	}

	return nil
}

func (gis *defaultGlobalInsightService) aggregateZoneEgresses(
	ctx context.Context,
	globalInsight *api_types.GlobalInsight,
) error {
	zoneEgressInsights := &mesh.ZoneEgressInsightResourceList{}
	if err := gis.resManager.List(ctx, zoneEgressInsights); err != nil {
		return err
	}

	for _, zoneEgressInsight := range zoneEgressInsights.GetItems() {
		subscriptions := zoneEgressInsight.GetSpec().(*mesh_proto.ZoneEgressInsight).GetSubscriptions()

		for _, subscription := range subscriptions {
			globalInsight.Zones.ZoneEgresses.Total += 1

			if subscription.GetDisconnectTime() == nil {
				globalInsight.Zones.ZoneEgresses.Online += 1
			}
		}
	}

	return nil
}
