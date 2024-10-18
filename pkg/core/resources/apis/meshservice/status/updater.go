package status

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/util/maps"
	util_time "github.com/kumahq/kuma/pkg/util/time"
)

type StatusUpdater struct {
	roResManager manager.ReadOnlyResourceManager
	resManager   manager.ResourceManager
	logger       logr.Logger
	metric       prometheus.Summary
	interval     time.Duration
	localZone    string
}

var _ component.Component = &StatusUpdater{}

func NewStatusUpdater(
	logger logr.Logger,
	roResManager manager.ReadOnlyResourceManager,
	resManager manager.ResourceManager,
	interval time.Duration,
	metrics core_metrics.Metrics,
	localZone string,
) (component.Component, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_ms_status_updater",
		Help:       "Summary of Inter CP Heartbeat component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &StatusUpdater{
		roResManager: roResManager,
		resManager:   resManager,
		logger:       logger,
		metric:       metric,
		interval:     interval,
		localZone:    localZone,
	}, nil
}

func (s *StatusUpdater) Start(stop <-chan struct{}) error {
	// sleep to mitigate update conflicts with other components
	util_time.SleepUpTo(s.interval)
	s.logger.Info("starting")
	ticker := time.NewTicker(s.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := s.updateStatus(ctx); err != nil {
				s.logger.Error(err, "could not update status of mesh services")
			}
			s.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			s.logger.Info("stopping")
			return nil
		}
	}
}

func (s *StatusUpdater) updateStatus(ctx context.Context) error {
	msList := &meshservice_api.MeshServiceResourceList{}
	if err := s.roResManager.List(ctx, msList); err != nil {
		return errors.Wrap(err, "could not list of Dataplanes")
	}
	if len(msList.Items) == 0 {
		// skip fetching other resources if MeshService is not used
		return nil
	}
	dpList := core_mesh.DataplaneResourceList{}
	if err := s.roResManager.List(ctx, &dpList); err != nil {
		return errors.Wrap(err, "could not list of Dataplanes")
	}

	dpInsightsList := core_mesh.DataplaneInsightResourceList{}
	if err := s.roResManager.List(ctx, &dpInsightsList); err != nil {
		return errors.Wrap(err, "could not list of DataplaneInsights")
	}

	meshList := core_mesh.MeshResourceList{}
	if err := s.roResManager.List(ctx, &meshList); err != nil {
		return errors.Wrap(err, "could not list of Meshes")
	}

	insightsByKey := core_model.IndexByKey(dpInsightsList.Items)
	meshByKey := core_model.IndexByKey(meshList.Items)

	dppsForMs := meshservice.MatchDataplanesWithMeshServices(dpList.Items, msList.Items, false)

	for ms, dpps := range dppsForMs {
		if !ms.IsLocalMeshService() {
			// identities are already computed by the other zone
			continue
		}
		log := s.logger.WithValues("meshservice", ms.GetMeta().GetName(), "mesh", ms.GetMeta().GetMesh())

		var changeReasons []string

		identities := buildIdentities(dpps)
		if !reflect.DeepEqual(ms.Spec.Identities, identities) {
			changeReasons = append(changeReasons, "identities")
			ms.Spec.Identities = identities
		}

		mesh := meshByKey[core_model.ResourceKey{Name: ms.Meta.GetMesh()}]
		tls := buildTLS(ms.Status.TLS, dpps, insightsByKey, mesh)
		if !reflect.DeepEqual(ms.Status.TLS, tls) {
			changeReasons = append(changeReasons, "tls status")
			ms.Status.TLS = tls
		}

		dataplaneProxies := buildDataplaneProxies(dpps, insightsByKey, ms.Spec.Ports)
		if !reflect.DeepEqual(ms.Status.DataplaneProxies, dataplaneProxies) {
			changeReasons = append(changeReasons, "data plane proxies")
			ms.Status.DataplaneProxies = dataplaneProxies
		}

		state := meshservice_api.StateUnavailable
		if dataplaneProxies.Healthy > 0 {
			state = meshservice_api.StateAvailable
		}
		if ms.Spec.State != state {
			changeReasons = append(changeReasons, "availability state")
			ms.Spec.State = state
		}

		if len(changeReasons) > 0 {
			log.Info("updating mesh service", "reason", changeReasons)
			if err := s.resManager.Update(ctx, ms); err != nil {
				if errors.Is(err, &store.ResourceConflictError{}) {
					log.Info("couldn't update mesh service, because it was modified in another place. Will try again in the next interval", "interval", s.interval)
				} else {
					log.Error(err, "could not update mesh service", "reason", changeReasons)
				}
			}
		}
	}
	return nil
}

func buildDataplaneProxies(
	dpps []*core_mesh.DataplaneResource,
	insightsByKey map[core_model.ResourceKey]*core_mesh.DataplaneInsightResource,
	ports []meshservice_api.Port,
) meshservice_api.DataplaneProxies {
	result := meshservice_api.DataplaneProxies{}
	for _, dpp := range dpps {
		result.Total++
		if insight := insightsByKey[core_model.MetaToResourceKey(dpp.Meta)]; insight != nil {
			if insight.Spec.IsOnline() {
				result.Connected++
			}
		}
		healthyInbounds := 0
		for _, port := range ports {
			if inbound := dpInboundForMeshServicePort(dpp.Spec.GetNetworking().Inbound, port); inbound != nil {
				if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ready {
					healthyInbounds++
				}
			}
		}
		if healthyInbounds == len(ports) {
			result.Healthy++
		}
	}
	return result
}

func dpInboundForMeshServicePort(inbounds []*mesh_proto.Dataplane_Networking_Inbound, port meshservice_api.Port) *mesh_proto.Dataplane_Networking_Inbound {
	for _, inbound := range inbounds {
		if port.Name != "" && inbound.Name == port.Name {
			return inbound
		}
		if port.Port == inbound.Port {
			return inbound
		}
	}
	return nil
}

func buildTLS(
	existing meshservice_api.TLS,
	dpps []*core_mesh.DataplaneResource,
	insightsByName map[core_model.ResourceKey]*core_mesh.DataplaneInsightResource,
	mesh *core_mesh.MeshResource,
) meshservice_api.TLS {
	if !mesh.MTLSEnabled() {
		return meshservice_api.TLS{
			Status: meshservice_api.TLSNotReady,
		}
	}
	if mesh.MTLSEnabled() && existing.Status == meshservice_api.TLSReady {
		// If mTLS is enabled, the status should go only one way.
		// Every new instance always starts with mTLS, so we don't want to count issued backends.
		// Otherwise, we could get into race when new Dataplane did not receive cert yet,
		// We would flip TLS to NotReady for a short period of time.
		return existing
	}

	issuedBackends := 0
	for _, dpp := range dpps {
		if insight := insightsByName[core_model.MetaToResourceKey(dpp.Meta)]; insight != nil {
			// Cert issued by any backend means that mTLS cert was issued to the DP
			// We don't want to check specific backend value, because we might be in a middle of CA rotation.
			if insight.Spec.GetMTLS().GetIssuedBackend() != "" {
				issuedBackends++
			}
		}
	}
	if issuedBackends == len(dpps) {
		return meshservice_api.TLS{
			Status: meshservice_api.TLSReady,
		}
	} else {
		return meshservice_api.TLS{
			Status: meshservice_api.TLSNotReady,
		}
	}
}

func buildIdentities(dpps []*core_mesh.DataplaneResource) []meshservice_api.MeshServiceIdentity {
	serviceTagIdentities := map[string]struct{}{}
	for _, dpp := range dpps {
		for service := range dpp.Spec.TagSet()[mesh_proto.ServiceTag] {
			serviceTagIdentities[service] = struct{}{}
		}
	}
	var identites []meshservice_api.MeshServiceIdentity
	for _, identity := range maps.SortedKeys(serviceTagIdentities) {
		identites = append(identites, meshservice_api.MeshServiceIdentity{
			Type:  meshservice_api.MeshServiceIdentityServiceTagType,
			Value: identity,
		})
	}
	return identites
}

func (s *StatusUpdater) NeedLeaderElection() bool {
	return true
}
