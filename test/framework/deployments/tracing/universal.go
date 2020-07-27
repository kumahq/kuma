package tracing

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	"github.com/kumahq/kuma/test/framework"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"strconv"
)

type universalDeployment struct {
	container string
	ip string
	ports map[string]string
}

var _ Deployment = &universalDeployment{}

type jaegerServicesOutput struct {
	Data []string `json:"data"`
}

func (u *universalDeployment) Name() string {
	panic("implement me")
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	u.allocatePublicPortsFor("16686")

	opts := docker.RunOptions{
		Detach: true,
		Remove: true,
		EnvironmentVariables: []string{"COLLECTOR_ZIPKIN_HTTP_PORT=9411"},
		OtherOptions: append([]string{"--network", "kind"}, u.publishPortsForDocker()...),
	}
	container, err := docker.RunAndGetIDE(cluster.GetTesting(), "jaegertracing/all-in-one:1.18", &opts)
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
		Logger: logger.Discard,
	}
	return shell.RunCommandAndGetStdOutE(t, cmd)
}

func (j *universalDeployment) allocatePublicPortsFor(ports ...string) {
	for i, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", uint32(33204 + i), uint32(34204 + i))
		if err != nil {
			panic(err)
		}
		j.ports[port] = strconv.Itoa(int(pubPortUInt32))
	}
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
	port := u.ports["16686"]
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/services", port))
	if err != nil {
		return nil, err
	}
	output := &jaegerServicesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	sort.Strings(output.Data)
	return output.Data, nil
}
