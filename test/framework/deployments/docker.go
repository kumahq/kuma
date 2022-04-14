package deployments

import (
	"strconv"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type DockerContainer struct {
	t          testing.TestingT
	name       string
	id         string
	ports      map[uint32]uint32
	envVars    map[string]string
	image      string
	args       map[string][]string
	dockerOpts *docker.RunOptions
}

type DockerContainerOptFn func(d *DockerContainer) error

func NewDockerContainer(fs ...DockerContainerOptFn) (*DockerContainer, error) {
	d := &DockerContainer{
		ports:   map[uint32]uint32{},
		envVars: map[string]string{},
		args:    map[string][]string{},
		dockerOpts: &docker.RunOptions{
			Detach: true,
			Remove: true,
		},
	}

	for _, f := range fs {
		if err := f(d); err != nil {
			return nil, errors.Errorf("couldn't create docker containter: %s", err)
		}
	}

	d.dockerOpts.Name = d.name
	d.dockerOpts.EnvironmentVariables = envVarsForDocker(d.envVars)
	d.dockerOpts.OtherOptions = buildArgsForDocker(d.args, d.ports)

	return d, nil
}

func (d *DockerContainer) applyArgs(name string, values ...string) *DockerContainer {
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

func buildArgsForDocker(args map[string][]string, ports map[uint32]uint32) []string {
	var opts []string

	for key, values := range args {
		for _, value := range values {
			opts = append(opts, key, value)
		}
	}

	for port := range ports {
		port := strconv.Itoa(int(port))
		opts = append(opts, "--publish="+port)
	}

	return opts
}

func WithTestingT(t testing.TestingT) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		d.t = t

		return nil
	}
}

func WithEnvVar(name, value string) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		d.envVars[name] = value

		return nil
	}
}

func WithContainerName(name string) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		d.name = name

		return nil
	}
}

func WithImage(image string) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		d.image = image

		return nil
	}
}

func WithNetwork(network string) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		d.applyArgs("--network", network)

		return nil
	}
}

func (d *DockerContainer) addID(id string) {
	d.id = id
}

func (d *DockerContainer) GetID() string {
	return d.id
}

func (d *DockerContainer) GetIP() (string, error) {
	args := []string{
		"container",
		"inspect",
		"-f",
		"{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}",
		d.id,
	}

	cmd := shell.Command{
		Command: "docker",
		Args:    args,
		Logger:  logger.Discard,
	}

	return shell.RunCommandAndGetStdOutE(d.t, cmd)
}

func (d *DockerContainer) GetName() (string, error) {
	args := []string{
		"container",
		"inspect",
		"-f",
		"{{.Name}}",
		d.id,
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

func AllocatePublicPortsFor(ports ...uint32) DockerContainerOptFn {
	return func(d *DockerContainer) error {
		for _, port := range ports {
			d.ports[port] = uint32(0)
		}

		return nil
	}
}

func (d *DockerContainer) Run() error {
	container, err := docker.RunAndGetIDE(d.t, d.image, d.dockerOpts)
	if err != nil {
		return err
	}

	d.addID(container)
	if err := d.updatePorts(); err != nil {
		return err
	}

	return nil
}

func (d *DockerContainer) updatePorts() error {
	var ports []uint32
	for port := range d.ports {
		ports = append(ports, port)
	}
	publishedPorts, err := framework.GetPublishedDockerPorts(d.t, d.id, ports)
	if err != nil {
		return err
	}
	d.ports = publishedPorts
	return nil
}

func (d *DockerContainer) Stop() error {
	retry.DoWithRetry(d.t, "stop "+d.id, 30, 3*time.Second,
		func() (string, error) {
			_, err := docker.StopE(d.t, []string{d.id}, &docker.StopOptions{Time: 1})
			if err != nil {
				return "Container still running", err
			}

			return "Container stopped", nil
		})

	return nil
}
