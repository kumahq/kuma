package tracing

import (
	"fmt"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	util_net "github.com/kumahq/kuma/pkg/util/net"
	"github.com/kumahq/kuma/test/framework"
)

type universalDeployment struct {
	container string
	ip        string
	ports     map[string]string
}

var _ Deployment = &universalDeployment{}

func (u *universalDeployment) Name() string {
	return DeploymentName
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.allocatePublicPortsFor("16686"); err != nil {
		return err
	}

	opts := docker.RunOptions{
		Detach:               true,
		Remove:               true,
		EnvironmentVariables: []string{"COLLECTOR_ZIPKIN_HOST_PORT=9411"},
		OtherOptions:         append([]string{"--network", "kind"}, u.publishPortsForDocker()...),
	}
	container, err := docker.RunAndGetIDE(cluster.GetTesting(), "jaegertracing/all-in-one:1.22", &opts)
	if err != nil {
		return err
	}

	ip, err := getIP(cluster.GetTesting(), container)
	if err != nil {
		return err
	}

	u.ip = ip
	u.container = container
	return nil
}

func getIP(t testing.TestingT, container string) (string, error) {
	cmd := shell.Command{
		Command: "docker",
		Args:    []string{"container", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container},
		Logger:  logger.Discard,
	}
	return shell.RunCommandAndGetStdOutE(t, cmd)
}

func (j *universalDeployment) allocatePublicPortsFor(ports ...string) error {
	for i, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", uint32(33204+i), uint32(34204+i))
		if err != nil {
			return err
		}
		j.ports[port] = strconv.Itoa(int(pubPortUInt32))
	}
	return nil
}

func (j *universalDeployment) publishPortsForDocker() (args []string) {
	for port, pubPort := range j.ports {
		args = append(args, "--publish="+pubPort+":"+port)
	}
	return
}

func (u *universalDeployment) Delete(cluster framework.Cluster) error {
	retry.DoWithRetry(cluster.GetTesting(), "stop "+u.container, framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			_, err := docker.StopE(cluster.GetTesting(), []string{u.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})
	return nil
}

func (u *universalDeployment) ZipkinCollectorURL() string {
	return fmt.Sprintf("http://%s:9411/api/v2/spans", u.ip)
}

func (u *universalDeployment) TracedServices() ([]string, error) {
	return tracedServices(fmt.Sprintf("http://localhost:%s", u.ports["16686"]))
}
