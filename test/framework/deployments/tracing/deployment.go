package tracing

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type Tracing interface {
	ZipkinCollectorURL() string
	TracedServices() ([]string, error)
}

type Deployment interface {
	framework.Deployment
	Tracing
}

const DeploymentName = "tracing"

func From(cluster framework.Cluster) Tracing {
	return cluster.Deployment(DeploymentName).(Tracing)
}

func Install() framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{}
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				ports: map[uint32]uint32{},
			}
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}
