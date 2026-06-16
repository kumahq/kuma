package postgres

import (
	"errors"

	. "github.com/kumahq/kuma/v2/test/framework"
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
	DefaultPreloadImages    = true

	PostgresImage = "postgres:latest@sha256:29ee7bb30d804447dc9a91fd0d74322ae1dc3a4072cc6346f70a5ed6e783b565"

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
	initScripts      []string
	preloadImages    bool
}
type DeployOptionsFunc func(*deployOptions)

func From(cluster Cluster, name string) Postgres {
	return cluster.Deployment(AppPostgres + name).(Postgres)
}

func Install(name string, optFns ...DeployOptionsFunc) InstallFunc {
	opts := &deployOptions{
		deploymentName:   name,
		namespace:        Config.KumaNamespace,
		primaryName:      AppPostgres,
		database:         DefaultPostgresDBName,
		username:         DefaultPostgresUser,
		password:         DefaultPostgresPassword,
		postgresPassword: DefaultPostgresPassword,
		preloadImages:    DefaultPreloadImages,
	}

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
		o.initScripts = append(o.initScripts, initScript)
	}
}

func WithoutPreloadImages() DeployOptionsFunc {
	return func(o *deployOptions) {
		o.preloadImages = false
	}
}
