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
		switch c := cluster.(type) {
		case *framework.K8sCluster:
			deployment := &k8SDeployment{}
			c.Deployments[DeploymentName] = deployment
			if err := deployment.Deploy(cluster); err != nil {
				return err
			}
		case *framework.UniversalCluster:
			deployment := &universalDeployment{
				ports: map[string]string{},
			}
			c.Deployments[DeploymentName] = deployment
			if err := deployment.Deploy(cluster); err != nil {
				return err
			}
		default:
			return errors.New("invalid cluster")
		}
		return nil
	}
}
