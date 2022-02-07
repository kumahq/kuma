package externalservice

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
	name      string
	cert      string
	commands  []Command
	app       *framework.SshApp
	verbose   bool
}

var _ Deployment = &universalDeployment{}

var UniversalAppEchoServer = ExternalServiceCommand(80, "Echo 80")
var UniversalAppEchoServer81 = ExternalServiceCommand(81, "Echo 81")
var UniversalAppHttpsEchoServer = Command([]string{"ncat",
	"-lk", "-p", "443",
	"--ssl", "--ssl-cert", "/server-cert.pem", "--ssl-key", "/server-key.pem",
	"--sh-exec", "'echo \"HTTP/1.1 200 OK\n\n HTTPS Echo\n\"'"})

var ExternalServiceCommand = func(port uint32, message string) Command {
	return []string{"ncat", "-lk", "-p", fmt.Sprintf("%d", port), "--sh-exec",
		fmt.Sprintf("'echo \"HTTP/1.1 200 OK\n\n%s\n\"'", message)}
}

func (u *universalDeployment) Name() string {
	return DeploymentName + u.name
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.allocatePublicPortsFor("22"); err != nil {
		return err
	}

	dockerOpts := docker.RunOptions{
		Detach:               true,
		Remove:               true,
		EnvironmentVariables: []string{},
		OtherOptions:         append([]string{"--name", cluster.Name() + "_" + u.Name(), "--network", "kind"}, u.publishPortsForDocker()...),
	}
	container, err := docker.RunAndGetIDE(cluster.GetTesting(), framework.Config.GetUniversalImage(), &dockerOpts)
	if err != nil {
		return err
	}

	ip, err := getIP(cluster.GetTesting(), container)
	if err != nil {
		return err
	}

	name, err := getName(cluster.GetTesting(), container)
	if err != nil {
		return err
	}

	u.ip = ip
	u.container = container

	port := u.ports["22"]

	// ceritficates
	cert, key, err := framework.CreateCertsFor("localhost", ip, name)
	if err != nil {
		return err
	}

	err = framework.NewSshApp(u.verbose, port, nil, []string{"printf ", "--", "\"" + cert + "\"", ">", "/server-cert.pem"}).Run()
	if err != nil {
		panic(err)
	}

	err = framework.NewSshApp(u.verbose, port, nil, []string{"printf ", "--", "\"" + key + "\"", ">", "/server-key.pem"}).Run()
	if err != nil {
		panic(err)
	}

	u.cert = cert
	for _, arg := range u.commands {
		u.app = framework.NewSshApp(u.verbose, port, nil, arg)
		err = u.app.Start()
		if err != nil {
			return err
		}
	}

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

func getName(t testing.TestingT, container string) (string, error) {
	cmd := shell.Command{
		Command: "docker",
		Args:    []string{"container", "inspect", "-f", "{{.Name}}", container},
		Logger:  logger.Discard,
	}
	out, err := shell.RunCommandAndGetStdOutE(t, cmd)
	if err != nil {
		return "", err
	}

	return out[1:], nil
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

func (u *universalDeployment) GetExternalAppAddress() string {
	return u.ip
}

func (u *universalDeployment) GetCert() string {
	return u.cert
}
