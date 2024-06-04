package hostname

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/util/maps"
)

type Generator struct {
	logger     logr.Logger
	interval   time.Duration
	metric     prometheus.Summary
	resManager manager.ResourceManager
}

var _ component.Component = &Generator{}

func NewGenerator(
	logger logr.Logger,
	metrics core_metrics.Metrics,
	resManager manager.ResourceManager,
	interval time.Duration,
) (*Generator, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_hostname_generator_ms",
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
				return err
			}
			g.metric.Observe(float64(time.Since(start).Milliseconds()))
		case <-stop:
			g.logger.Info("stopping")
			return nil
		}
	}
}

func apply(generator *hostnamegenerator_api.HostnameGeneratorResource, service *meshservice_api.MeshServiceResource) (string, error) {
	if !generator.Spec.Selector.MeshService.Matches(service.Meta.GetLabels()) {
		return "", nil
	}
	sb := strings.Builder{}
	tmpl := template.New("").Funcs(
		map[string]any{
			"label": func(key string) (string, error) {
				val, ok := service.GetMeta().GetLabels()[key]
				if !ok {
					return "", errors.Errorf("label %s not found", key)
				}
				return val, nil
			},
		},
	)
	tmpl, err := tmpl.Parse(generator.Spec.Template)
	if err != nil {
		return "", fmt.Errorf("failed compiling gotemplate error=%q", err.Error())
	}
	type meshedName struct {
		Name      string
		Namespace string
		Mesh      string
	}
	err = tmpl.Execute(&sb, meshedName{
		Name:      service.GetMeta().GetNameExtensions()[model.K8sNameComponent],
		Namespace: service.GetMeta().GetNameExtensions()[model.K8sNamespaceComponent],
		Mesh:      service.GetMeta().GetMesh(),
	})
	if err != nil {
		return "", fmt.Errorf("pre evaluation of template with parameters failed with error=%q", err.Error())
	}
	return sb.String(), nil
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
	services := &meshservice_api.MeshServiceResourceList{}
	if err := g.resManager.List(ctx, services); err != nil {
		return errors.Wrap(err, "could not list MeshServices")
	}

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
		conditions []meshservice_api.Condition
	}
	generatedHostnames := map[string]serviceKey{}
	newStatuses := map[serviceKey]map[string]status{}
	for _, generator := range sortGenerators(generators.Items) {
		for _, service := range services.Items {
			serviceKey := serviceKey{
				name: service.GetMeta().GetName(),
				mesh: service.GetMeta().GetMesh(),
			}
			generatorStatuses, ok := newStatuses[serviceKey]
			if !ok {
				generatorStatuses = map[string]status{}
			}

			generated, err := apply(generator, service)

			var conditions []meshservice_api.Condition
			if generated != "" || err != nil {
				generationConditionStatus := kube_meta.ConditionUnknown
				reason := "Pending"
				var message string
				if err != nil {
					generationConditionStatus = kube_meta.ConditionFalse
					reason = meshservice_api.TemplateErrorReason
					message = err.Error()
				}
				if generated != "" {
					if key, ok := generatedHostnames[generated]; ok && key != serviceKey {
						generationConditionStatus = kube_meta.ConditionFalse
						reason = meshservice_api.CollisionReason
						message = fmt.Sprintf("Hostname collision with MeshService %s", serviceKey.name)
						generated = ""
					} else {
						generationConditionStatus = kube_meta.ConditionTrue
						reason = meshservice_api.GeneratedReason
						generatedHostnames[generated] = serviceKey
					}
				}
				condition := meshservice_api.Condition{
					Type:    meshservice_api.GeneratedCondition,
					Status:  generationConditionStatus,
					Reason:  reason,
					Message: message,
				}
				conditions = []meshservice_api.Condition{
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

	for _, service := range services.Items {
		statuses := newStatuses[serviceKey{
			name: service.GetMeta().GetName(),
			mesh: service.GetMeta().GetMesh(),
		}]
		var addresses []meshservice_api.Address
		var generatorStatuses []meshservice_api.HostnameGeneratorStatus

		for _, generator := range maps.SortedKeys(statuses) {
			status := statuses[generator]
			ref := meshservice_api.HostnameGeneratorRef{
				CoreName: generator,
			}
			if status.hostname == "" && len(status.conditions) == 0 {
				continue
			}
			if status.hostname != "" {
				addresses = append(
					addresses,
					meshservice_api.Address{
						Hostname:             status.hostname,
						Origin:               meshservice_api.OriginGenerator,
						HostnameGeneratorRef: ref,
					},
				)
			}
			generatorStatuses = append(generatorStatuses, meshservice_api.HostnameGeneratorStatus{
				HostnameGeneratorRef: ref,
				Conditions:           status.conditions,
			})
		}
		if reflect.DeepEqual(addresses, service.Status.Addresses) && reflect.DeepEqual(generatorStatuses, service.Status.HostnameGenerators) {
			continue
		}
		service.Status.Addresses = addresses
		service.Status.HostnameGenerators = generatorStatuses
		if err := g.resManager.Update(ctx, service); err != nil {
			return errors.Wrap(err, "couldn't update MeshService status")
		}
	}

	return nil
}

func (g *Generator) NeedLeaderElection() bool {
	return true
}
