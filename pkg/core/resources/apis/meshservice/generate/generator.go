package generate

import (
	"context"
	"reflect"
	"slices"
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
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// Generator generates MeshService objects from Dataplane resources created on
// universal.
type Generator struct {
	logger     logr.Logger
	interval   time.Duration
	metric     prometheus.Summary
	resManager manager.ResourceManager
	meshCache  *mesh.Cache
}

var _ component.Component = &Generator{}

func New(
	logger logr.Logger,
	interval time.Duration,
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
		logger:     logger,
		interval:   interval,
		metric:     metric,
		resManager: resManager,
		meshCache:  meshCache,
	}, nil
}

func (g *Generator) meshServicesForDataplane(dataplane *mesh_proto.Dataplane) map[string]*meshservice_api.MeshService {
	portsByService := map[string][]meshservice_api.Port{}
	for _, inbound := range dataplane.Networking.Inbound {
		serviceTag := inbound.GetTags()[mesh_proto.ServiceTag]
		appProtocol, hasProtocol := inbound.GetTags()[mesh_proto.ProtocolTag]
		if !hasProtocol {
			appProtocol = core_mesh.ProtocolTCP
		}
		port := meshservice_api.Port{
			Port:        inbound.Port,
			TargetPort:  intstr.FromInt(int(inbound.Port)),
			AppProtocol: core_mesh.Protocol(appProtocol),
		}
		portsByService[serviceTag] = append(portsByService[serviceTag], port)
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
		if !slices.EqualFunc(meshService.Ports, generatedMeshService.meshService.Ports, func(a, b meshservice_api.Port) bool {
			return reflect.DeepEqual(a, b)
		}) {
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

func (g *Generator) Start(stop <-chan struct{}) error {
	g.logger.Info("starting")
	ticker := time.NewTicker(g.interval)
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
				dataplanes := meshCtx.Resources.Dataplanes()
				g.generate(ctx, mesh, dataplanes.Items)
			}
			g.metric.Observe(float64(time.Since(start).Milliseconds()))

		case <-stop:
			g.logger.Info("stopping")
			return nil
		}
	}
}

func (g *Generator) generate(ctx context.Context, mesh string, dataplanes []*core_mesh.DataplaneResource) {
	meshservicesByName := map[string][]dataplaneAndMeshService{}
	for _, dataplane := range dataplanes {
		// TODO order these
		for name, ms := range g.meshServicesForDataplane(dataplane.Spec) {
			meshservicesByName[name] = append(meshservicesByName[name], dataplaneAndMeshService{
				dataplane:   dataplane.GetMeta(),
				meshService: ms,
			})
		}
	}

	msList := meshservice_api.MeshServiceResourceList{}
	if err := g.resManager.List(ctx, &msList, store.ListByMesh(mesh)); err != nil {
		g.logger.Error(err, "failed to list MeshServices")
	}

	for _, meshService := range msList.Items {
		conflicting, newMeshService := checkMeshServicesConsistency(meshService.Spec, meshservicesByName[meshService.GetMeta().GetName()])
		var dps []string
		for _, dp := range conflicting {
			dps = append(dps, dp.dataplane.GetName())
		}
		if len(conflicting) > 0 {
			g.logger.Info("conflicting for MeshService", "MeshService", meshService.GetMeta().GetName(), "dps", dps)
		}
		delete(meshservicesByName, meshService.GetMeta().GetName())
		if newMeshService != nil && !reflect.DeepEqual(meshService.Spec, newMeshService) {
			meshService.Spec = newMeshService
			if err := g.resManager.Update(ctx, meshService); err != nil {
				g.logger.Error(err, "couldn't update MeshService", "name", meshService.GetMeta().GetName())
				continue
			}
		} else if newMeshService == nil {
			if err := g.resManager.Delete(ctx, meshService, store.DeleteBy(model.MetaToResourceKey(meshService.GetMeta()))); err != nil {
				log.Error(err, "couldn't delete MeshService")
				continue
			}
			log.Info("deleted MeshService")
		}
	}
	for name, meshServices := range meshservicesByName {
		conflicting, newMeshService := checkMeshServicesConsistency(nil, meshServices)
		meshService := meshservice_api.NewMeshServiceResource()
		meshService.Spec = newMeshService
		var dps []string
		for _, dp := range conflicting {
			dps = append(dps, dp.dataplane.GetName())
		}
		if len(conflicting) > 0 {
			g.logger.Info("conflicting for MeshService", "MeshService", meshService.GetMeta().GetName(), "dps", dps)
		}
		if err := g.resManager.Create(ctx, meshService, store.CreateByKey(name, mesh), store.CreateWithLabels(map[string]string{
			mesh_proto.ManagedByLabel:      "meshservice-generator",
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
		})); err != nil {
			g.logger.Error(err, "couldn't create MeshService", "name", name)
			continue
		}
	}
}

func (g *Generator) NeedLeaderElection() bool {
	return true
}
