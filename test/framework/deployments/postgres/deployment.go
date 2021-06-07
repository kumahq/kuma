package postgres

import (
	"errors"

	. "github.com/kumahq/kuma/test/framework"
)

func Install() InstallFunc {
	return func(cluster Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *UniversalCluster:
			deployment = NewUniversalDeployment(cluster)
		case *K8sCluster:
			return errors.New("kubernetes cluster not supported for postgres deployment")
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
