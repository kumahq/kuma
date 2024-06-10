package hostname

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/util/maps"
)

type HostnameGenerator interface {
	GetResources(context.Context) (model.ResourceList, error)
	UpdateResourceStatus(context.Context, model.Resource, []hostnamegenerator_api.HostnameGeneratorStatus, []hostnamegenerator_api.Address) error
	HasStatusChanged(model.Resource, []hostnamegenerator_api.HostnameGeneratorStatus, []hostnamegenerator_api.Address) (bool, error)
	GenerateHostname(*hostnamegenerator_api.HostnameGeneratorResource, model.Resource) (string, error)
}

type Generator struct {
	logger     logr.Logger
	interval   time.Duration
	metric     prometheus.Summary
	resManager manager.ResourceManager
	generators []HostnameGenerator
}

var _ component.Component = &Generator{}

func NewGenerator(
	logger logr.Logger,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	interval time.Duration,
	generators []HostnameGenerator,
) (*Generator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_hostname_generator",
		Help:       "Summary of hostname generator interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Generator{
		logger:     logger,
		resManager: resManager,
		interval:   interval,
		metric:     metric,
		generators: generators,
	}, nil
}

func (g *Generator) Start(stop <-chan struct{}) error {
	g.logger.Info("starting")
	ticker := time.NewTicker(g.interval)
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if err := g.generateHostnames(ctx); err != nil {
				g.logger.Error(err, "couldn't generate hostnames")
			}
			g.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			g.logger.Info("stopping")
			return nil
		}
	}
}

func sortGenerators(generators []*hostnamegenerator_api.HostnameGeneratorResource) []*hostnamegenerator_api.HostnameGeneratorResource {
	sorted := slices.Clone(generators)
	slices.SortFunc(sorted, func(a, b *hostnamegenerator_api.HostnameGeneratorResource) int {
		if a, b := a.Meta.GetLabels()[mesh_proto.ResourceOriginLabel], b.Meta.GetLabels()[mesh_proto.ResourceOriginLabel]; a != b {
			if a == string(mesh_proto.ZoneResourceOrigin) {
				return -1
			} else if b == string(mesh_proto.ZoneResourceOrigin) {
				return 1
			}
		}
		if a, b := a.Meta.GetCreationTime(), b.Meta.GetCreationTime(); a.Before(b) {
			return -1
		} else if a.After(b) {
			return 1
		}
		return strings.Compare(a.Meta.GetName(), b.Meta.GetName())
	})
	return sorted
}

func (g *Generator) generateHostnames(ctx context.Context) error {
	generators := &hostnamegenerator_api.HostnameGeneratorResourceList{}
	if err := g.resManager.List(ctx, generators); err != nil {
		return errors.Wrap(err, "could not list HostnameGenerators")
	}
	type serviceKey struct {
		name string
		mesh string
	}
	type status struct {
		hostname   string
		conditions []hostnamegenerator_api.Condition
	}
	for _, generatorType := range g.generators {
		resources, err := generatorType.GetResources(ctx)
		if err != nil {
			g.logger.Error(err, "couldn't get resources", "type", generatorType)
			continue
		}
		type meshName string
		type serviceName string
		type hostname string
		generatedHostnames := map[meshName]map[hostname]serviceName{}
		newStatuses := map[serviceKey]map[string]status{}
		for _, generator := range sortGenerators(generators.Items) {
			for _, service := range resources.GetItems() {
				serviceKey := serviceKey{
					name: service.GetMeta().GetName(),
					mesh: service.GetMeta().GetMesh(),
				}
				generatorStatuses, ok := newStatuses[serviceKey]
				if !ok {
					generatorStatuses = map[string]status{}
				}

				generated, err := generatorType.GenerateHostname(generator, service)

				var conditions []hostnamegenerator_api.Condition
				if generated != "" || err != nil {
					generationConditionStatus := kube_meta.ConditionUnknown
					reason := "Pending"
					var message string
					if err != nil {
						generationConditionStatus = kube_meta.ConditionFalse
						reason = hostnamegenerator_api.TemplateErrorReason
						message = err.Error()
					}
					if generated != "" {
						if svcName, ok := generatedHostnames[meshName(serviceKey.mesh)][hostname(generated)]; ok && string(svcName) != serviceKey.name {
							generationConditionStatus = kube_meta.ConditionFalse
							reason = hostnamegenerator_api.CollisionReason
							message = fmt.Sprintf("Hostname collision with %s: %s", resources.GetItemType(), serviceKey.name)
							generated = ""
						} else {
							generationConditionStatus = kube_meta.ConditionTrue
							reason = hostnamegenerator_api.GeneratedReason
							meshHostnames, ok := generatedHostnames[meshName(serviceKey.mesh)]
							if !ok {
								meshHostnames = map[hostname]serviceName{}
							}
							meshHostnames[hostname(generated)] = serviceName(serviceKey.name)
							generatedHostnames[meshName(serviceKey.mesh)] = meshHostnames
						}
					}
					condition := hostnamegenerator_api.Condition{
						Type:    hostnamegenerator_api.GeneratedCondition,
						Status:  generationConditionStatus,
						Reason:  reason,
						Message: message,
					}
					conditions = []hostnamegenerator_api.Condition{
						condition,
					}
				}

				generatorStatuses[generator.GetMeta().GetName()] = status{
					hostname:   generated,
					conditions: conditions,
				}
				newStatuses[serviceKey] = generatorStatuses
			}
		}
		for _, service := range resources.GetItems() {
			statuses := newStatuses[serviceKey{
				name: service.GetMeta().GetName(),
				mesh: service.GetMeta().GetMesh(),
			}]
			var addresses []hostnamegenerator_api.Address
			var generatorStatuses []hostnamegenerator_api.HostnameGeneratorStatus

			for _, generator := range maps.SortedKeys(statuses) {
				status := statuses[generator]
				ref := hostnamegenerator_api.HostnameGeneratorRef{
					CoreName: generator,
				}
				if status.hostname == "" && len(status.conditions) == 0 {
					continue
				}
				if status.hostname != "" {
					addresses = append(
						addresses,
						hostnamegenerator_api.Address{
							Hostname:             status.hostname,
							Origin:               hostnamegenerator_api.OriginGenerator,
							HostnameGeneratorRef: ref,
						},
					)
				}
				generatorStatuses = append(generatorStatuses, hostnamegenerator_api.HostnameGeneratorStatus{
					HostnameGeneratorRef: ref,
					Conditions:           status.conditions,
				})
			}
			changed, changedErr := generatorType.HasStatusChanged(service, generatorStatuses, addresses)
			if changedErr != nil {
				g.logger.Error(err, "couldn't check status", "type", resources.GetItemType())
				continue
			}
			if !changed {
				continue
			}
			if err := generatorType.UpdateResourceStatus(ctx, service, generatorStatuses, addresses); err != nil {
				g.logger.Error(err, "couldn't update status", "type", resources.GetItemType())
				continue
			}
		}
	}
	return nil
}

func (g *Generator) NeedLeaderElection() bool {
	return true
}
