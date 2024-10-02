package generate

import (
	"context"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

const (
	managedByValue string = "meshservice-generator"
)

// Generator generates MeshService objects from Dataplane resources created on
// universal.
type Generator struct {
	logger              logr.Logger
	generateInterval    time.Duration
	deletionGracePeriod time.Duration
	metric              prometheus.Summary
	resManager          manager.ResourceManager
	meshCache           *mesh.Cache
}

var _ component.Component = &Generator{}

func New(
	logger logr.Logger,
	generateInterval time.Duration,
	deletionGracePeriod time.Duration,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	meshCache *mesh.Cache,
) (*Generator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_meshservice_generator",
		Help:       "Summary of MeshService generation duration",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Generator{
		logger:              logger,
		generateInterval:    generateInterval,
		deletionGracePeriod: deletionGracePeriod,
		metric:              metric,
		resManager:          resManager,
		meshCache:           meshCache,
	}, nil
}

func (g *Generator) meshServicesForDataplane(dataplane *core_mesh.DataplaneResource) map[string]*meshservice_api.MeshService {
	log := g.logger.WithValues("mesh", dataplane.GetMeta().GetMesh(), "Dataplane", dataplane.GetMeta().GetName())
	portsByService := map[string][]meshservice_api.Port{}
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		serviceTagValue := inbound.GetTags()[mesh_proto.ServiceTag]
		if !core_mesh.NameCharacterSet.MatchString(serviceTagValue) {
			log.Info("couldn't generate MeshService from kuma.io/service, contains invalid characters", "value", serviceTagValue, "regex", core_mesh.NameCharacterSet)
			continue
		}
		appProtocol, hasProtocol := inbound.GetTags()[mesh_proto.ProtocolTag]
		if !hasProtocol {
			appProtocol = core_mesh.ProtocolTCP
		}
		portName := inbound.Name
		if portName == "" {
			portName = strconv.Itoa(int(inbound.Port))
		}
		port := meshservice_api.Port{
			Name:        portName,
			Port:        inbound.Port,
			TargetPort:  intstr.FromInt(int(inbound.Port)),
			AppProtocol: core_mesh.Protocol(appProtocol),
		}
		portsByService[serviceTagValue] = append(portsByService[serviceTagValue], port)
	}

	services := map[string]*meshservice_api.MeshService{}
	for serviceTag, ports := range portsByService {
		ms := meshservice_api.MeshService{
			Selector: meshservice_api.Selector{
				DataplaneTags: meshservice_api.DataplaneTags{
					mesh_proto.ServiceTag: serviceTag,
				},
			},
			Ports: ports,
		}
		services[serviceTag] = &ms
	}
	return services
}

type dataplaneAndMeshService struct {
	dataplane   model.ResourceMeta
	meshService *meshservice_api.MeshService
}

// checkMeshServicesConsistency returns a list of dataplanes that conflict and
// the chosen meshService
func checkMeshServicesConsistency(
	meshService *meshservice_api.MeshService,
	generated []dataplaneAndMeshService,
) ([]dataplaneAndMeshService, *meshservice_api.MeshService) {
	if len(generated) == 0 {
		return nil, nil
	}
	if meshService == nil {
		meshService = generated[0].meshService
	}
	var conflicting []dataplaneAndMeshService
	for _, generatedMeshService := range generated {
		if servicesDiffer(meshService, generatedMeshService.meshService) {
			conflicting = append(conflicting, generatedMeshService)
		}
	}
	if len(generated) == len(conflicting) {
		// None of the generated MeshServices match the existing, drop it and
		// try again with the first generated. We know we won't recurse
		// indefinitely because generated[0] won't conflict
		return checkMeshServicesConsistency(generated[0].meshService, generated)
	}
	return conflicting, meshService
}

func sortDataplanes(dps []*core_mesh.DataplaneResource) []*core_mesh.DataplaneResource {
	sorted := slices.Clone(dps)
	slices.SortFunc(sorted, func(a, b *core_mesh.DataplaneResource) int {
		if a, b := a.Meta.GetCreationTime(), b.Meta.GetCreationTime(); a.Before(b) {
			return -1
		} else if a.After(b) {
			return 1
		}
		return strings.Compare(a.Meta.GetName(), b.Meta.GetName())
	})
	return sorted
}

func servicesDiffer(a, b *meshservice_api.MeshService) bool {
	return !reflect.DeepEqual(a.Ports, b.Ports)
}

func (g *Generator) generate(ctx context.Context, mesh string, dataplanes []*core_mesh.DataplaneResource, meshServices []*meshservice_api.MeshServiceResource) {
	log := g.logger.WithValues("mesh", mesh)
	meshservicesByName := map[string][]dataplaneAndMeshService{}
	for _, dataplane := range sortDataplanes(dataplanes) {
		for name, ms := range g.meshServicesForDataplane(dataplane) {
			meshservicesByName[name] = append(meshservicesByName[name], dataplaneAndMeshService{
				dataplane:   dataplane.GetMeta(),
				meshService: ms,
			})
		}
	}

	for _, meshService := range meshServices {
		if managedBy, ok := meshService.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]; !ok || managedBy != managedByValue {
			continue
		}
		log := log.WithValues("MeshService", meshService.GetMeta().GetName())
		conflicting, newMeshService := checkMeshServicesConsistency(meshService.Spec, meshservicesByName[meshService.GetMeta().GetName()])
		var dps []string
		for _, dp := range conflicting {
			dps = append(dps, dp.dataplane.GetName())
		}
		if len(conflicting) > 0 {
			log.Info("Port conflict for a kuma.io/service tag, ports must be identical across Dataplane inbounds for a given kuma.io/service", "dps", dps)
		}
		delete(meshservicesByName, meshService.GetMeta().GetName())
		gracePeriodStartedAtText, hasGracePeriodLabel := meshService.GetMeta().GetLabels()[mesh_proto.DeletionGracePeriodStartedLabel]

		servicesDiffer := newMeshService != nil && servicesDiffer(meshService.Spec, newMeshService)
		if newMeshService != nil && (servicesDiffer || hasGracePeriodLabel) {
			meta := meshService.GetMeta()
			meshService = meshservice_api.NewMeshServiceResource()
			meshService.Meta = meta
			meshService.Spec = newMeshService

			// Unset the grace period by deleting the label
			newLabels := maps.Clone(meshService.GetMeta().GetLabels())
			delete(newLabels, mesh_proto.DeletionGracePeriodStartedLabel)

			if err := g.resManager.Update(ctx, meshService, store.UpdateWithLabels(newLabels)); err != nil {
				log.Error(err, "couldn't update MeshService")
				continue
			}
			var reasons []string
			if servicesDiffer {
				reasons = append(reasons, "spec changed")
			}
			if hasGracePeriodLabel {
				reasons = append(reasons, "no longer scheduled for deletion")
			}
			log.Info("updated MeshService", "reasons", reasons)
		} else if newMeshService == nil {
			gracePeriodStartedAt, err := time.Parse(time.RFC3339, gracePeriodStartedAtText)
			if hasGracePeriodLabel && err == nil {
				// If we have a valid grace period set, check if it's expired
				if time.Since(gracePeriodStartedAt) > g.deletionGracePeriod {
					if err := g.resManager.Delete(ctx, meshservice_api.NewMeshServiceResource(), store.DeleteBy(model.MetaToResourceKey(meshService.GetMeta()))); err != nil {
						log.Error(err, "couldn't delete MeshService")
						continue
					}
					log.Info("deleted MeshService")
				}
			} else {
				// Start the grace period if we don't have the label or it's invalid
				if hasGracePeriodLabel && err != nil {
					log.Info("couldn't parse grace period label, ignoring", "value", gracePeriodStartedAtText)
				}
				nowText, err := time.Now().MarshalText()
				if err != nil {
					log.Error(err, "couldn't marshal time.Now as text, this shouldn't be possible")
					continue
				}
				newLabels := maps.Clone(meshService.GetMeta().GetLabels())
				newLabels[mesh_proto.DeletionGracePeriodStartedLabel] = string(nowText)
				if err := g.resManager.Update(ctx, meshService, store.UpdateWithLabels(newLabels)); err != nil {
					log.Error(err, "couldn't update MeshService")
					continue
				}
				log.Info("MeshService deletion grace period started", "period", g.deletionGracePeriod.String())
			}
		}
	}
	for name, meshServices := range meshservicesByName {
		log := log.WithValues("MeshService", name)
		conflicting, newMeshService := checkMeshServicesConsistency(nil, meshServices)
		meshService := meshservice_api.NewMeshServiceResource()
		meshService.Spec = newMeshService
		var dps []string
		for _, dp := range conflicting {
			dps = append(dps, dp.dataplane.GetName())
		}
		if len(conflicting) > 0 {
			log.Info("Port conflict for a kuma.io/service tag, ports must be identical across Dataplane inbounds for a given kuma.io/service", "dps", dps)
		}
		if err := g.resManager.Create(ctx, meshService, store.CreateByKey(name, mesh), store.CreateWithLabels(map[string]string{
			metadata.KumaMeshLabel:         mesh,
			mesh_proto.DisplayName:         name,
			mesh_proto.ManagedByLabel:      managedByValue,
			mesh_proto.EnvTag:              mesh_proto.UniversalEnvironment,
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
		})); err != nil {
			log.Error(err, "couldn't create MeshService")
			continue
		}
		log.Info("created MeshService")
	}
}

func (g *Generator) NeedLeaderElection() bool {
	return true
}

func (g *Generator) Start(stop <-chan struct{}) error {
	g.logger.Info("starting")
	ticker := time.NewTicker(g.generateInterval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, g.resManager, g.meshCache.GetMeshContext)
			if err != nil {
				return err
			}
			for mesh, meshCtx := range aggregatedMeshCtxs.MeshContextsByName {
				if meshCtx.Resource.Spec.MeshServicesMode() != mesh_proto.Mesh_MeshServices_Disabled {
					dataplanes := meshCtx.Resources.Dataplanes()
					meshServices := meshCtx.Resources.MeshServices()
					g.generate(ctx, mesh, dataplanes.Items, meshServices.Items)
				}
			}
			g.metric.Observe(float64(time.Since(start).Milliseconds()))

		case <-stop:
			g.logger.Info("stopping")
			return nil
		}
	}
}
