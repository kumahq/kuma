package framework

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"strconv"
)

type UniversalJaeger struct {
	container string
	ip string
	ports map[string]string
}

type jaegerServicesOutput struct {
	Data []string `json:"data"`
}

func (j *UniversalJaeger) TracedServices() ([]string, error) {
	port := j.ports["16686"]
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

func DeployJaegerInDocker(t testing.TestingT) (*UniversalJaeger, error) {
	jaeger := &UniversalJaeger{
		ports: map[string]string{},
	}

	jaeger.allocatePublicPortsFor("16686")

	opts := docker.RunOptions{
		Detach: true,
		Remove: true,
		EnvironmentVariables: []string{"COLLECTOR_ZIPKIN_HTTP_PORT=9411"},
		OtherOptions: append([]string{"--network", "kind"}, jaeger.publishPortsForDocker()...),
	}
	container, err := docker.RunAndGetIDE(t, "jaegertracing/all-in-one:1.18", &opts)
	if err != nil {
		return nil, err
	}

	ip, err := getIP(t, container)
	if err != nil {
		return nil, err
	}

	jaeger.ip = ip
	jaeger.container = container
	return jaeger, nil
}

func StopJaegerDocker(t testing.TestingT, jaeger *UniversalJaeger) {
	retry.DoWithRetry(t, "stop "+jaeger.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			_, err := docker.StopE(t, []string{jaeger.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})
}

func getIP(t testing.TestingT, container string) (string, error) {
	cmd := shell.Command{
		Command: "docker",
		Args:    []string{"container", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container},
		Logger: logger.Discard,
	}
	return shell.RunCommandAndGetStdOutE(t, cmd)
}

func (j *UniversalJaeger) ZipkinCollectorURL() string {
	return fmt.Sprintf("http://%s:9411/api/v2/spans", j.ip)
}

func (j *UniversalJaeger) allocatePublicPortsFor(ports ...string) {
	for i, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", uint32(33204 + i), uint32(34204 + i))
		if err != nil {
			panic(err)
		}
		j.ports[port] = strconv.Itoa(int(pubPortUInt32))
	}
}

func (j *UniversalJaeger) publishPortsForDocker() (args []string) {
	for port, pubPort := range j.ports {
		args = append(args, "--publish="+pubPort+":"+port)
	}
	return
}
