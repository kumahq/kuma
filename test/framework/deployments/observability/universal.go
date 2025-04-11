package observability

import (
	"fmt"
	"net"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type universalDeployment struct {
	deploymentName string
	container      string
	ip             string
	ports          map[uint32]uint32
	logger         *logger.Logger
	dockerBackend  framework.DockerBackend
}

var _ Deployment = &universalDeployment{}

func (u *universalDeployment) Name() string {
	return u.deploymentName
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	u.dockerBackend = cluster.(*framework.UniversalCluster).GetDockerBackend()
	u.allocatePublicPortsFor(16686)

	opts := docker.RunOptions{
		Detach:               true,
		Remove:               true,
		EnvironmentVariables: []string{"COLLECTOR_ZIPKIN_HOST_PORT=9411"},
		OtherOptions:         append([]string{"--network", framework.DockerNetworkName}, u.publishPortsForDocker()...),
		Logger:               u.logger,
	}
	container, err := u.dockerBackend.RunAndGetIDE(cluster.GetTesting(), "jaegertracing/all-in-one:1.67.0", &opts)
	if err != nil {
		return err
	}

	ip, err := u.getIP(cluster.GetTesting(), container)
	if err != nil {
		return err
	}

	u.ip = ip
	u.container = container
	if err := u.updatePublishedPorts(cluster.GetTesting()); err != nil {
		return err
	}
	return nil
}

func (u *universalDeployment) getIP(t testing.TestingT, container string) (string, error) {
	args := []string{"container", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container}
	return u.dockerBackend.RunCommandAndGetStdOutE(t, "container-inspect", args, logger.Discard)
}

func (u *universalDeployment) allocatePublicPortsFor(ports ...uint32) {
	for _, port := range ports {
		u.ports[port] = 0
	}
}

func (u *universalDeployment) publishPortsForDocker() []string {
	var args []string
	for port := range u.ports {
		args = append(args, "--publish=0.0.0.0::"+strconv.Itoa(int(port)))
	}
	return args
}

func (u *universalDeployment) updatePublishedPorts(t testing.TestingT) error {
	var ports []uint32
	for port := range u.ports {
		ports = append(ports, port)
	}
	publishedPorts, err := u.dockerBackend.GetPublishedDockerPorts(t, u.logger, u.container, ports)
	if err != nil {
		return err
	}
	u.ports = publishedPorts
	return nil
}

func (u *universalDeployment) Delete(cluster framework.Cluster) error {
	retry.DoWithRetry(cluster.GetTesting(), "stop "+u.container, framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			_, err := u.dockerBackend.StopE(cluster.GetTesting(), []string{u.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})
	return nil
}

func (u *universalDeployment) ZipkinCollectorURL() string {
	return fmt.Sprintf("http://%s/api/v2/spans", net.JoinHostPort(u.ip, "9411"))
}

func (u *universalDeployment) TracedServices() ([]string, error) {
	return tracedServices(fmt.Sprintf("http://localhost:%d", u.ports[16686]))
}
