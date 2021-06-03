package postgres

import (
	"errors"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments"
)

type Deployment interface {
	framework.Deployment
}

func Install() framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				DockerDeployment: deployments.NewDockerDeployment(),
			}
		case *framework.K8sCluster:
			return errors.New("kubernetes cluster not supported for postgres deployment")
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
