package postgres

import (
	"bytes"
	"os/exec"
	"regexp"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments"
)

const (
	AppKumaCP                = framework.AppKumaCP
	AppPostgres              = framework.AppPostgres
	StoreTypePostgres        = framework.StoreTypePostgres
	PostgresImage            = framework.PostgresImage
	PostgresEnvVarUser       = framework.PostgresEnvVarUser
	PostgresEnvVarPassword   = framework.PostgresEnvVarPassword
	PostgresEnvVarDB         = framework.PostgresEnvVarDB
	EnvStoreType             = framework.EnvStoreType
	EnvStorePostgresUser     = framework.EnvStorePostgresUser
	EnvStorePostgresPassword = framework.EnvStorePostgresPassword
	EnvStorePostgresDBName   = framework.EnvStorePostgresDBName
	EnvStorePostgresPort     = framework.EnvStorePostgresPort
	EnvStorePostgresHost     = framework.EnvStorePostgresHost
	DefaultPostgresPort      = framework.DefaultPostgresPort
	DefaultPostgresUser      = framework.DefaultPostgresUser
	DefaultPostgresPassword  = framework.DefaultPostgresPassword
	DefaultPostgresDBName    = framework.DefaultPostgresDBName
	AfterCreateMainApp       = framework.AfterCreateMainApp
)

type universalDeployment struct {
	*deployments.DockerDeployment
}

var _ Deployment = &universalDeployment{}

func (u *universalDeployment) Name() string {
	return AppPostgres
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.AllocatePublicPortsFor("5432"); err != nil {
		return err
	}

	err := u.
		WithName(cluster.Name()+"_"+u.Name()).
		WithTestingT(cluster.GetTesting()).
		WithNetwork("kind").
		WithImage(PostgresImage).
		WithEnvVar(PostgresEnvVarUser, DefaultPostgresUser).
		WithEnvVar(PostgresEnvVarPassword, DefaultPostgresPassword).
		WithEnvVar(PostgresEnvVarDB, DefaultPostgresDBName).
		RunContainer()
	if err != nil {
		return err
	}

	ip, err := u.GetIP()
	if err != nil {
		return err
	}

	cluster.
		WithEnvVar(AppKumaCP, EnvStoreType, string(StoreTypePostgres)).
		WithEnvVar(AppKumaCP, EnvStorePostgresUser, DefaultPostgresUser).
		WithEnvVar(AppKumaCP, EnvStorePostgresPassword, DefaultPostgresPassword).
		WithEnvVar(AppKumaCP, EnvStorePostgresDBName, DefaultPostgresDBName).
		WithEnvVar(AppKumaCP, EnvStorePostgresPort, DefaultPostgresPort).
		WithEnvVar(AppKumaCP, EnvStorePostgresHost, ip).
		WithHookFn(AppKumaCP, AfterCreateMainApp, RunDBMigration)

	return u.waitTillReady(cluster.GetTesting())
}

func (u *universalDeployment) Delete(cluster framework.Cluster) error {
	return u.StopContainer()
}

func (u *universalDeployment) waitTillReady(t testing.TestingT) error {
	retry.DoWithRetry(t, "logs "+u.GetContainerID(), framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd := exec.Command("docker", "logs", u.GetContainerID())
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				return "docker logs command failed", err
			}

			if stderr.Len() != 0 {
				return "command returned stderr", errors.New(stderr.String())
			}

			matched, err := regexp.Match("database system is ready to accept connections", stdout.Bytes())
			if matched {
				return "Postgres is ready", nil
			} else {
				return "Postgres is not ready yet", err
			}
		})

	return nil
}

var RunDBMigration = framework.HookFn(func(cluster framework.Cluster) (framework.Cluster, error) {
	args := []string{
		"/usr/bin/kuma-cp", "migrate", "up",
	}

	c, ok := cluster.(*framework.UniversalCluster)
	if !ok {
		return nil, errors.New("unsupported cluster type")
	}

	kumaCP := c.GetApp(AppKumaCP)
	if kumaCP == nil {
		return nil, errors.New("missing kuma-cp deployment in the cluster")
	}

	sshPort := kumaCP.GetPublicPort("22")
	if sshPort == "" {
		return nil, errors.New("missing public port: 22")
	}

	app := framework.NewSshApp(true, sshPort, c.GetEnvVars(AppKumaCP), args)
	if err := app.Run(); err != nil {
		return nil, errors.Errorf("db migration err: %s\nstderr :%s\nstdout %s", err.Error(), app.Err(), app.Out())
	}

	return cluster, nil
})
