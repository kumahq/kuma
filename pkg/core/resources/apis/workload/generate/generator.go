package generate

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	mesh_cache "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

const (
	managedByValue = "workload-generator"
)

// Generator generates Workload objects from Dataplane resources created on
// Universal.
type Generator struct {
	logger           logr.Logger
	generateInterval time.Duration
	metric           prometheus.Histogram
	resManager       manager.ResourceManager
	meshCache        *mesh_cache.Cache
	zone             string
}

var _ component.Component = &Generator{}

func New(
	logger logr.Logger,
	generateInterval time.Duration,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	meshCache *mesh_cache.Cache,
	zone string,
) (*Generator, error) {
	metric := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "component_workload_generator",
		Help: "Summary of Workload generation duration",
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Generator{
		logger:           logger,
		generateInterval: generateInterval,
		metric:           metric,
		resManager:       resManager,
		meshCache:        meshCache,
		zone:             zone,
	}, nil
}

// workloadForDataplane extracts and validates the workload identifier from a dataplane
func (g *Generator) workloadForDataplane(dataplane *core_mesh.DataplaneResource) (string, bool) {
	workloadName, ok := dataplane.GetMeta().GetLabels()[metadata.KumaWorkload]
	if !ok || workloadName == "" {
		return "", false
	}
	allErrs := apimachineryvalidation.NameIsDNS1035Label(workloadName, false)
	if len(allErrs) != 0 {
		g.logger.Info("couldn't generate Workload from kuma.io/workload label, contains invalid characters", "value", workloadName, "error", allErrs, "dataplane", dataplane.GetMeta().GetName())
		return "", false
	}
	return workloadName, true
}

func (g *Generator) generate(ctx context.Context, mesh string, dataplanes []*core_mesh.DataplaneResource, workloads []*workload_api.WorkloadResource) {
	log := g.logger.WithValues("mesh", mesh)

	workloadsByName := g.collectWorkloadsByName(dataplanes)

	for _, workload := range workloads {
		workloadName := workload.GetMeta().GetName()
		stillExists := workloadsByName[workloadName]

		if !stillExists {
			log := log.WithValues("Workload", workloadName)
			if err := g.resManager.Delete(ctx, workload_api.NewWorkloadResource(), store.DeleteBy(model.MetaToResourceKey(workload.GetMeta()))); err != nil {
				log.Error(err, "couldn't delete Workload")
			} else {
				log.Info("deleted Workload")
			}
		}
		delete(workloadsByName, workloadName)
	}

	for workloadName := range workloadsByName {
		g.createWorkload(ctx, log, workloadName, mesh)
	}
}

func (g *Generator) collectWorkloadsByName(dataplanes []*core_mesh.DataplaneResource) map[string]bool {
	workloadsByName := map[string]bool{}
	for _, dataplane := range core_mesh.SortDataplanes(dataplanes) {
		if workloadName, ok := g.workloadForDataplane(dataplane); ok {
			workloadsByName[workloadName] = true
		}
	}
	return workloadsByName
}

func (g *Generator) createWorkload(ctx context.Context, log logr.Logger, workloadName, mesh string) {
	log = log.WithValues("Workload", workloadName)
	workload := workload_api.NewWorkloadResource()
	// Workload spec is intentionally empty here - the actual spec is computed
	// separately by the workload status controller based on dataplane state
	workload.Spec = &workload_api.Workload{}
	if err := g.resManager.Create(ctx, workload, store.CreateByKey(workloadName, mesh), store.CreateWithLabels(map[string]string{
		metadata.KumaMeshLabel:    mesh,
		mesh_proto.ManagedByLabel: managedByValue,
		mesh_proto.EnvTag:         mesh_proto.UniversalEnvironment,
		mesh_proto.ZoneTag:        g.zone,
	})); err != nil {
		log.Error(err, "couldn't create Workload")
		return
	}
	log.V(1).Info("created Workload")
}

func (g *Generator) NeedLeaderElection() bool {
	return true
}

func (g *Generator) Start(stop <-chan struct{}) error {
	g.logger.Info("starting")
	ticker := time.NewTicker(g.generateInterval)
	defer ticker.Stop()
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-stop
		cancel()
	}()

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, g.resManager, g.meshCache.GetMeshContext)
			if err != nil {
				// Error from AggregateMeshContexts indicates a fundamental issue
				// with resource manager or mesh cache, so we terminate the component
				g.logger.Error(err, "fundamental issue with resource manager or mesh cache, terminating component")
				return err
			}
			for mesh, meshCtx := range aggregatedMeshCtxs.MeshContextsByName {
				dataplanes := meshCtx.Resources.Dataplanes()
				workloads := meshCtx.Resources.Workloads()
				g.generate(ctx, mesh, dataplanes.Items, workloads.Items)
			}
			g.metric.Observe(float64(time.Since(start).Milliseconds()))

		case <-stop:
			g.logger.Info("stopping")
			return nil
		}
	}
}
