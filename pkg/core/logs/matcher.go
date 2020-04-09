package logs

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"

	"github.com/pkg/errors"
)

var logger = core.Log.WithName("logs")

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

func (m *TrafficLogsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (core_xds.LogMap, error) {
	logs := &mesh_core.TrafficLogResourceList{}
	if err := m.ResourceManager.List(ctx, logs, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic logs")
	}
	backends, err := m.backendsByName(ctx, dataplane)
	if err != nil {
		return nil, err
	}

	policies := make([]policy.ConnectionPolicy, len(logs.Items))
	for i, log := range logs.Items {
		policies[i] = log
	}
	policyMap := policy.SelectOutboundConnectionPolicies(dataplane, policies)

	logMap := core_xds.LogMap{}
	for service, policy := range policyMap {
		log := policy.(*mesh_core.TrafficLogResource)
		backend, found := backends[log.Spec.GetConf().GetBackend()]
		if !found {
			logger.Info("Logging backend is not found. Ignoring.", "name", log.Spec.GetConf().GetBackend(), "trafficLog", log.GetMeta())
			continue
		}
		logMap[service] = backend
	}
	return logMap, nil
}

func (m *TrafficLogsMatcher) backendsByName(ctx context.Context, dataplane *mesh_core.DataplaneResource) (map[string]*mesh_proto.LoggingBackend, error) {
	mesh := mesh_core.MeshResource{}
	if err := m.ResourceManager.Get(ctx, &mesh, store.GetByKey(dataplane.GetMeta().GetMesh(), dataplane.GetMeta().GetMesh())); err != nil {
		return nil, err
	}
	backendsByName := map[string]*mesh_proto.LoggingBackend{}
	for _, backend := range mesh.Spec.GetLogging().GetBackends() {
		backendsByName[backend.Name] = backend
	}
	defaultBackend := mesh.Spec.GetLogging().GetDefaultBackend()
	if defaultBackend != "" {
		backendsByName[""] = backendsByName[defaultBackend]
	}
	return backendsByName, nil
}
