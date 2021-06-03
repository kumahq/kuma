package deployments

import (
	"strconv"
	"time"

	"github.com/pkg/errors"

	util_net "github.com/kumahq/kuma/pkg/util/net"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
)

type DockerDeployment struct {
	t          testing.TestingT
	name       string
	container  string
	ports      map[string]string
	envVars    map[string]string
	image      string
	args       map[string][]string
	dockerOpts *docker.RunOptions
}

func NewDockerDeployment() *DockerDeployment {
	return &DockerDeployment{
		ports:   map[string]string{},
		envVars: map[string]string{},
		args:    map[string][]string{},
		dockerOpts: &docker.RunOptions{
			Detach: true,
			Remove: true,
		},
	}
}

func (d *DockerDeployment) applyArgs(name string, values ...string) *DockerDeployment {
	if d.args[name] == nil {
		d.args[name] = []string{}
	}

	finalValues := map[string]struct{}{}
	for _, value := range d.args[name] {
		finalValues[value] = struct{}{}
	}

	for _, value := range values {
		finalValues[value] = struct{}{}
	}

	for value := range finalValues {
		d.args[name] = append(d.args[name], value)
	}

	return d
}

func envVarsForDocker(envVars map[string]string) []string {
	var finalEnvVars []string

	for name, value := range envVars {
		finalEnvVars = append(finalEnvVars, name+"="+value)
	}

	return finalEnvVars
}

func buildArgsForDocker(args map[string][]string, ports map[string]string) []string {
	var opts []string

	for key, values := range args {
		for _, value := range values {
			opts = append(opts, key, value)
		}
	}

	for port, pubPort := range ports {
		opts = append(opts, "--publish="+pubPort+":"+port)
	}

	return opts
}

func (d *DockerDeployment) WithTestingT(t testing.TestingT) *DockerDeployment {
	d.t = t

	return d
}

func (d *DockerDeployment) WithContainer(container string) *DockerDeployment {
	d.container = container

	return d
}

func (d *DockerDeployment) WithEnvVar(name, value string) *DockerDeployment {
	d.envVars[name] = value

	return d
}

func (d *DockerDeployment) WithName(name string) *DockerDeployment {
	d.name = name

	return d
}

func (d *DockerDeployment) WithImage(image string) *DockerDeployment {
	d.image = image

	return d
}

func (d *DockerDeployment) WithNetwork(network string) *DockerDeployment {
	return d.applyArgs("--network", network)
}

func (d *DockerDeployment) GetContainerID() string {
	return d.container
}

func (d *DockerDeployment) GetIP() (string, error) {
	args := []string{
		"container",
		"inspect",
		"-f",
		"{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}",
		d.container,
	}

	cmd := shell.Command{
		Command: "docker",
		Args:    args,
		Logger:  logger.Discard,
	}

	return shell.RunCommandAndGetStdOutE(d.t, cmd)
}

func (d *DockerDeployment) GetContainerName() (string, error) {
	args := []string{
		"container",
		"inspect",
		"-f",
		"{{.Name}}",
		d.container,
	}

	cmd := shell.Command{
		Command: "docker",
		Args:    args,
		Logger:  logger.Discard,
	}

	out, err := shell.RunCommandAndGetStdOutE(d.t, cmd)
	if err != nil {
		return "", err
	}

	return out[1:], nil
}

func (d *DockerDeployment) AllocatePublicPortsFor(ports ...string) error {
	for i, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", uint32(33204+i), uint32(34204+i))
		if err != nil {
			return err
		}

		d.ports[port] = strconv.Itoa(int(pubPortUInt32))
	}

	return nil
}

func (d *DockerDeployment) RunContainer() error {
	d.dockerOpts.Name = d.name
	d.dockerOpts.EnvironmentVariables = envVarsForDocker(d.envVars)
	d.dockerOpts.OtherOptions = buildArgsForDocker(d.args, d.ports)

	container, err := docker.RunAndGetIDE(d.t, d.image, d.dockerOpts)
	if err != nil {
		return err
	}

	d.WithContainer(container)

	return nil
}

func (d *DockerDeployment) StopContainer() error {
	retry.DoWithRetry(d.t, "stop "+d.container, 30, 3*time.Second,
		func() (string, error) {
			_, err := docker.StopE(d.t, []string{d.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})

	return nil
}
