package logs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// Current limitations:
// 1) Right now we only generate and place access logs for outbound listeners
// 2) We match all tags in source section of TrafficLog but only service tag on destination
// 3) Let's assume we've got following dataplanes:
//    Dataplane 1 with services: kong and kong-admin
//    Dataplane 2 with services: backend
//    If we define rule kong->backend, it is also applied for kong-admin because there is no way to differentiate
//    traffic from services that are using one dataplane.
type TrafficLogsMatcher struct {
	ResourceManager manager.ReadOnlyResourceManager
}

func (m *TrafficLogsMatcher) Match(ctx context.Context, dataplane *core_mesh.DataplaneResource) (core_xds.TrafficLogMap, error) {
	logs := &core_mesh.TrafficLogResourceList{}
	if err := m.ResourceManager.List(ctx, logs, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic logs")
	}
	return BuildTrafficLogMap(dataplane, logs.Items), nil
}

func BuildTrafficLogMap(dataplane *core_mesh.DataplaneResource, trafficLogs []*core_mesh.TrafficLogResource) core_xds.TrafficLogMap {
	policies := make([]policy.ConnectionPolicy, len(trafficLogs))
	for i, log := range trafficLogs {
		policies[i] = log
	}
	policyMap := policy.SelectOutboundConnectionPolicies(dataplane, policies)

	trafficLogMap := core_xds.TrafficLogMap{}
	for service, connectionPolicy := range policyMap {
		trafficLogMap[service] = connectionPolicy.(*core_mesh.TrafficLogResource)
	}
	return trafficLogMap
}
