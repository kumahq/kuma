package logs

import (
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"sort"
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
	ResourceManager manager.ResourceManager
}

type MatchedLogs struct {
	Outbounds map[string][]*mesh_proto.LoggingBackend
}

func NewMatchedLogs() *MatchedLogs {
	return &MatchedLogs{
		Outbounds: map[string][]*mesh_proto.LoggingBackend{},
	}
}

func (m *MatchedLogs) AddForOutbound(outbound string, backend *mesh_proto.LoggingBackend) {
	if _, ok := m.Outbounds[outbound]; !ok {
		m.Outbounds[outbound] = []*mesh_proto.LoggingBackend{}
	}
	m.Outbounds[outbound] = append(m.Outbounds[outbound], backend)
	// sort the slice for stability of envoy configuration
	sort.Slice(m.Outbounds[outbound], func(i, j int) bool {
		return m.Outbounds[outbound][i].Name > m.Outbounds[outbound][j].Name
	})
}

func (m *TrafficLogsMatcher) Match(ctx context.Context, dataplane *mesh_core.DataplaneResource) (*MatchedLogs, error) {
	logs := &mesh_core.TrafficLogResourceList{}
	if err := m.ResourceManager.List(ctx, logs, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve traffic logs")
	}
	backends, err := m.backendsByName(ctx, dataplane)
	if err != nil {
		return nil, err
	}
	matchedLog := NewMatchedLogs()
	for outbound, backendsNames := range matchBackends(&dataplane.Spec, logs) {
		for backendName := range backendsNames {
			backend, found := backends[backendName]
			if !found {
				logger.Info("Logging backend is not found. Ignoring.", "name", backendName)
				continue
			}
			matchedLog.AddForOutbound(outbound, backend)
		}
	}
	return matchedLog, nil
}

func (m *TrafficLogsMatcher) backendsByName(ctx context.Context, dataplane *mesh_core.DataplaneResource) (map[string]*mesh_proto.LoggingBackend, error) {
	meshes := &mesh_core.MeshResourceList{}
	// todo(jakubydszkiewicz) simplify to Get after we solve namespace problem
	if err := m.ResourceManager.List(ctx, meshes, store.ListByMesh(dataplane.GetMeta().GetMesh())); err != nil {
		return nil, errors.Wrap(err, "could not retrieve meshes")
	}
	if len(meshes.Items) != 1 {
		return nil, errors.Errorf("found %d meshes. There should be only one mesh of name %s", len(meshes.Items), dataplane.GetMeta().GetMesh())
	}
	backendsByName := map[string]*mesh_proto.LoggingBackend{}
	for _, backend := range meshes.Items[0].Spec.GetLogging().GetBackends() {
		backendsByName[backend.Name] = backend
	}
	defaultBackend := meshes.Items[0].Spec.GetLogging().GetDefaultBackend()
	if defaultBackend != "" {
		backendsByName[""] = backendsByName[defaultBackend]
	}
	return backendsByName, nil
}

func matchBackends(dataplane *mesh_proto.Dataplane, logs *mesh_core.TrafficLogResourceList) map[string]map[string]bool {
	outboundToBackend := map[string]map[string]bool{}
	for _, outbound := range dataplane.GetNetworking().GetOutbound() {
		for _, logRes := range matchOutbound(outbound, dataplane.Networking.Inbound, logs.Items) {
			for _, rule := range logRes.Spec.Rules {
				if _, ok := outboundToBackend[outbound.Interface]; !ok {
					outboundToBackend[outbound.Interface] = map[string]bool{}
				}
				outboundToBackend[outbound.Interface][rule.Conf.GetBackend()] = true
			}
		}
	}
	return outboundToBackend
}

// To Match outbound, we need to match service tag of outbound and all tags of any inbound interface
func matchOutbound(outbound *mesh_proto.Dataplane_Networking_Outbound, inbounds []*mesh_proto.Dataplane_Networking_Inbound, logs []*mesh_core.TrafficLogResource) []*mesh_core.TrafficLogResource {
	matchedLogs := []*mesh_core.TrafficLogResource{}
	for _, log := range logs {
		matchedRules := []*mesh_proto.TrafficLog_Rule{}

		for _, rule := range log.Spec.Rules {
			if !anySelectorMatchAnyInbound(rule.Sources, inbounds) {
				continue
			}
			for _, dest := range rule.Destinations {
				if outbound.MatchTags(dest.Match) {
					matchedRules = append(matchedRules, rule)
				}
			}
		}

		if len(matchedRules) > 0 {
			// construct copy of the resource but only with matched rules
			matchedLogs = append(matchedLogs, &mesh_core.TrafficLogResource{
				Meta: log.Meta,
				Spec: mesh_proto.TrafficLog{
					Rules: matchedRules,
				},
			})
		}
	}
	return matchedLogs
}

func anySelectorMatchAnyInbound(selectors []*mesh_proto.Selector, inbounds []*mesh_proto.Dataplane_Networking_Inbound) bool {
	for _, inbound := range inbounds {
		for _, selector := range selectors {
			if inbound.MatchTags(selector.Match) {
				return true
			}
		}
	}
	return false
}
