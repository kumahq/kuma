package observability

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type Observability interface {
	ZipkinCollectorURL() string
	TracedServices() ([]string, error)
	Name() string
}

type Deployment interface {
	framework.Deployment
	Observability
}

type deployOptions struct {
	namespace      string
	deploymentName string
	components     []string
}
type deployOptionsFunc func(*deployOptions)

func From(deploymentName string, cluster framework.Cluster) Observability {
	return cluster.Deployment(deploymentName).(Observability)
}

func WithNamespace(namespace string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.namespace = namespace
	}
}

type Component string

const (
	JaegerComponent     Component = "jaeger"
	PrometheusComponent Component = "prometheus"
	GrafanaComponent    Component = "grafana"
	LokiComponent       Component = "loki"
)

func WithComponents(components ...Component) deployOptionsFunc {
	return func(o *deployOptions) {
		for _, c := range components {
			o.components = append(o.components, string(c))
		}
	}
}

func Install(name string, optFns ...deployOptionsFunc) framework.InstallFunc {
	opts := &deployOptions{deploymentName: name, namespace: framework.Config.DefaultObservabilityNamespace}
	for _, optFn := range optFns {
		optFn(opts)
	}
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{
				namespace:      opts.namespace,
				deploymentName: opts.deploymentName,
				components:     opts.components,
			}
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				deploymentName: name,
				ports:          map[uint32]uint32{},
			}
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}
