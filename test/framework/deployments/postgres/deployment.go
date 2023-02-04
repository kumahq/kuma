package postgres

import (
	"errors"

	. "github.com/kumahq/kuma/test/framework"
)

const (
	EnvStoreType             = "KUMA_STORE_TYPE"
	EnvStorePostgresHost     = "KUMA_STORE_POSTGRES_HOST"
	EnvStorePostgresPort     = "KUMA_STORE_POSTGRES_PORT"
	EnvStorePostgresUser     = "KUMA_STORE_POSTGRES_USER"
	EnvStorePostgresPassword = "KUMA_STORE_POSTGRES_PASSWORD"
	EnvStorePostgresDBName   = "KUMA_STORE_POSTGRES_DB_NAME"

	DefaultPostgresPort     = uint32(5432)
	DefaultPostgresUser     = "kuma"
	DefaultPostgresPassword = "kuma"
	DefaultPostgresDBName   = "kuma"

	PostgresImage = "postgres"

	PostgresEnvVarUser     = "POSTGRES_USER"
	PostgresEnvVarPassword = "POSTGRES_PASSWORD"
	PostgresEnvVarDB       = "POSTGRES_DB"

	AppPostgres = "postgres"
)

type Postgres interface {
	GetEnvVars() map[string]string
}

type PostgresDeployment interface {
	Postgres
	Deployment
}

type deployOptions struct {
	namespace      string
	deploymentName string
}
type deployOptionsFunc func(*deployOptions)

func From(cluster Cluster, name string) Postgres {
	return cluster.Deployment(AppPostgres + name).(Postgres)
}

func Install(name string, optFns ...deployOptionsFunc) InstallFunc {
	opts := &deployOptions{deploymentName: name, namespace: Config.DefaultPostgresNamespace}
	for _, optFn := range optFns {
		optFn(opts)
	}
	return func(cluster Cluster) error {
		var deployment PostgresDeployment
		switch cluster.(type) {
		case *UniversalCluster:
			deployment = NewUniversalDeployment(cluster, name)
		case *K8sCluster:
			deployment = NewK8SDeployment(opts)
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
