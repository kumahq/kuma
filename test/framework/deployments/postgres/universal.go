package postgres

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/deployments"
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
	container *DockerContainer
}

var _ Deployment = &universalDeployment{}

func NewUniversalDeployment(cluster framework.Cluster) *universalDeployment {
	name := cluster.Name() + "_" + AppPostgres

	container, err := NewDockerContainer(
		AllocatePublicPortsFor(DefaultPostgresPort),
		Name(name),
		TestingT(cluster.GetTesting()),
		Network("kind"),
		Image(PostgresImage),
		EnvVar(PostgresEnvVarUser, DefaultPostgresUser),
		EnvVar(PostgresEnvVarPassword, DefaultPostgresPassword),
		EnvVar(PostgresEnvVarDB, DefaultPostgresDBName),
	)
	if err != nil {
		panic(err)
	}

	return &universalDeployment{
		container: container,
	}
}

func (u *universalDeployment) Name() string {
	return AppPostgres
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.container.Run(); err != nil {
		return err
	}

	ip, err := u.container.GetIP()
	if err != nil {
		return err
	}

	port := strconv.Itoa(int(DefaultPostgresPort))

	cluster.
		WithEnvVar(AppKumaCP, EnvStoreType, string(StoreTypePostgres)).
		WithEnvVar(AppKumaCP, EnvStorePostgresUser, DefaultPostgresUser).
		WithEnvVar(AppKumaCP, EnvStorePostgresPassword, DefaultPostgresPassword).
		WithEnvVar(AppKumaCP, EnvStorePostgresDBName, DefaultPostgresDBName).
		WithEnvVar(AppKumaCP, EnvStorePostgresPort, port).
		WithEnvVar(AppKumaCP, EnvStorePostgresHost, ip).
		WithHookFn(AppKumaCP, AfterCreateMainApp, RunDBMigration)

	return u.waitTillReady(cluster.GetTesting())
}

func (u *universalDeployment) Delete(framework.Cluster) error {
	return u.container.Stop()
}

func (u *universalDeployment) waitTillReady(t testing.TestingT) error {
	containerID := u.container.GetID()

	r := regexp.MustCompile("database system is ready to accept connections")

	retry.DoWithRetry(t, "logs "+containerID, framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd := exec.Command("docker", "logs", containerID)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				return "docker logs command failed", err
			}

			if stderr.Len() != 0 {
				return "command returned stderr", errors.New(stderr.String())
			}

			if !r.Match(stdout.Bytes()) {
				return "Postgres is not ready yet", fmt.Errorf("failed to match against %q", stdout.String())
			}

			return "Postgres is ready", nil
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
