package postgres

import (
	"errors"

	. "github.com/kumahq/kuma/test/framework"
)

type Postgres interface {
	GetEnvVars() map[string]string
}

type PostgresDeployment interface {
	Postgres
	Deployment
}

func From(cluster Cluster, name string) Postgres {
	return cluster.Deployment(AppPostgres + name).(Postgres)
}

func Install(name string) InstallFunc {
	return func(cluster Cluster) error {
		var deployment PostgresDeployment
		switch cluster.(type) {
		case *UniversalCluster:
			deployment = NewUniversalDeployment(cluster, name)
		case *K8sCluster:
			return errors.New("kubernetes cluster not supported for postgres deployment")
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
