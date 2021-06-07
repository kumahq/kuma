package postgres

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"

	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/deployments"
)

type universalDeployment struct {
	name      string
	container *DockerContainer
	envVars   map[string]string
}

func (u *universalDeployment) GetEnvVars() map[string]string {
	return u.envVars
}

var _ Deployment = &universalDeployment{}

func NewUniversalDeployment(cluster Cluster, name string) *universalDeployment {
	container, err := NewDockerContainer(
		AllocatePublicPortsFor(DefaultPostgresPort),
		WithContainerName(cluster.Name()+"_"+AppPostgres),
		WithTestingT(cluster.GetTesting()),
		WithNetwork("kind"),
		WithImage(PostgresImage),
		WithEnvVar(PostgresEnvVarUser, DefaultPostgresUser),
		WithEnvVar(PostgresEnvVarPassword, DefaultPostgresPassword),
		WithEnvVar(PostgresEnvVarDB, DefaultPostgresDBName),
	)
	if err != nil {
		panic(err)
	}

	return &universalDeployment{
		name:      name,
		container: container,
		envVars: map[string]string{
			EnvStoreType:             "postgres",
			EnvStorePostgresUser:     DefaultPostgresUser,
			EnvStorePostgresPassword: DefaultPostgresPassword,
			EnvStorePostgresDBName:   DefaultPostgresDBName,
		},
	}
}

func (u *universalDeployment) Name() string {
	return AppPostgres + u.name
}

func (u *universalDeployment) Deploy(cluster Cluster) error {
	if err := u.container.Run(); err != nil {
		return err
	}

	ip, err := u.container.GetIP()
	if err != nil {
		return err
	}

	u.envVars[EnvStorePostgresHost] = ip
	u.envVars[EnvStorePostgresPort] = strconv.Itoa(int(DefaultPostgresPort))

	return u.waitTillReady(cluster.GetTesting())
}

func (u *universalDeployment) Delete(Cluster) error {
	return u.container.Stop()
}

func (u *universalDeployment) waitTillReady(t testing.TestingT) error {
	containerID := u.container.GetID()
	r := regexp.MustCompile("database system is ready to accept connections")

	retry.DoWithRetry(t, "logs "+containerID, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			var stdout bytes.Buffer

			cmd := exec.Command("docker", "logs", containerID)
			cmd.Stdout = &stdout

			if err := cmd.Run(); err != nil {
				return "docker logs command failed", err
			}

			if !r.Match(stdout.Bytes()) {
				return "Postgres is not ready yet", fmt.Errorf("failed to match against %q", stdout.String())
			}

			return "Postgres is ready", nil
		})

	return nil
}
