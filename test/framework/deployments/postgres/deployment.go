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
	EnvStorePostgresPassword = "KUMA_STORE_POSTGRES_PASSWORD" // #nosec G101 -- That's the env var not the pwd
	EnvStorePostgresDBName   = "KUMA_STORE_POSTGRES_DB_NAME"

	DefaultPostgresPort     = uint32(5432)
	DefaultPostgresUser     = "kuma"
	DefaultPostgresPassword = "kuma"
	DefaultPostgresDBName   = "kuma"

	PostgresImage = "postgres"

	PostgresEnvVarUser     = "POSTGRES_USER"
	PostgresEnvVarPassword = "POSTGRES_PASSWORD" // #nosec G101 -- Env var not actual password
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
	namespace        string
	deploymentName   string
	username         string
	password         string
	database         string
	primaryName      string
	postgresPassword string
	initScript       string
}
type DeployOptionsFunc func(*deployOptions)

func From(cluster Cluster, name string) Postgres {
	return cluster.Deployment(AppPostgres + name).(Postgres)
}

func Install(name string, optFns ...DeployOptionsFunc) InstallFunc {
	opts := &deployOptions{deploymentName: name, namespace: Config.KumaNamespace}
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

func WithK8sNamespace(namespace string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.namespace = namespace
	}
}

func WithUsername(username string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.username = username
	}
}

func WithPassword(password string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.password = password
	}
}

func WithDatabase(database string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.database = database
	}
}

func WithPrimaryName(primaryName string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.primaryName = primaryName
	}
}

func WithPostgresPassword(postgresPassword string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.postgresPassword = postgresPassword
	}
}

func WithInitScript(initScript string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.initScript = initScript
	}
}
