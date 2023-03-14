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

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal_logs"
)

type UniversalDeployment struct {
	container string
	ip        string
	ports     map[uint32]uint32
	name      string
	cert      string
	commands  []Command
	app       *ssh.App
	verbose   bool
}

var _ Deployment = &UniversalDeployment{}

func NewUniversalDeployment() *UniversalDeployment {
	return &UniversalDeployment{
		ports: map[uint32]uint32{},
	}
}

func (u *UniversalDeployment) WithName(name string) *UniversalDeployment {
	u.name = name

	return u
}

func (u *UniversalDeployment) WithCommands(commands ...Command) *UniversalDeployment {
	u.commands = commands

	return u
}

func (u *UniversalDeployment) WithVerbose(verbose bool) *UniversalDeployment {
	u.verbose = verbose

	return u
}

var UniversalEchoServer = func(port int, tls bool) Command {
	args := []string{
		"test-server", "echo",
		"--instance", fmt.Sprintf("echo-%d", port),
		"--port", fmt.Sprintf("%d", port),
	}
	if tls {
		args = append(args, "--crt", "/server-cert.pem", "--key", "/server-key.pem", "--tls")
	}
	return args
}

var (
	UniversalAppEchoServer      = ExternalServiceCommand(80, "Echo 80")
	UniversalAppHttpsEchoServer = Command([]string{
		"ncat",
		"-lk", "-p", "443",
		"--ssl", "--ssl-cert", "/server-cert.pem", "--ssl-key", "/server-key.pem",
		"--sh-exec", "'echo \"HTTP/1.1 200 OK\n\n HTTPS Echo\n\"'",
	})
)
var UniversalTCPSink = Command([]string{"ncat", "-lk", "9999", ">", "/nc.out"})

var ExternalServiceCommand = func(port uint32, message string) Command {
	return []string{
		"ncat", "-lk", "-p", fmt.Sprintf("%d", port), "--sh-exec",
		fmt.Sprintf("'echo \"HTTP/1.1 200 OK\n\n%s\n\"'", message),
	}
}

func (u *UniversalDeployment) Name() string {
	return DeploymentName + u.name
}

func (u *UniversalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.allocatePublicPortsFor(22); err != nil {
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

	if err := u.updatePublishedPorts(cluster.GetTesting()); err != nil {
		return err
	}

	port := strconv.Itoa(int(u.ports[22]))

	// certificates
	cert, key, err := framework.CreateCertsFor("localhost", ip, name)
	if err != nil {
		return err
	}

	err = ssh.NewApp(u.name, "", u.verbose, port, nil, []string{"printf ", "--", "\"" + cert + "\"", ">", "/server-cert.pem"}).Run()
	if err != nil {
		panic(err)
	}

	err = ssh.NewApp(u.name, "", u.verbose, port, nil, []string{"printf ", "--", "\"" + key + "\"", ">", "/server-key.pem"}).Run()
	if err != nil {
		panic(err)
	}

	u.cert = cert
	for _, arg := range u.commands {
		u.app = ssh.NewApp(u.name, universal_logs.LogsPath(framework.Config.UniversalE2ELogsPath), u.verbose, port, nil, arg)
		err = u.app.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UniversalDeployment) Exec(cmd ...string) (string, string, error) {
	port := strconv.Itoa(int(u.ports[22]))
	sshApp := ssh.NewApp(u.name, "", u.verbose, port, nil, cmd)
	err := sshApp.Run()
	return sshApp.Out(), sshApp.Err(), err
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

func (j *UniversalDeployment) allocatePublicPortsFor(ports ...uint32) error {
	for _, port := range ports {
		j.ports[port] = 0
	}
	return nil
}

func (j *UniversalDeployment) publishPortsForDocker() []string {
	var args []string
	for port := range j.ports {
		args = append(args, "--publish="+strconv.Itoa(int(port)))
	}
	return args
}

func (j *UniversalDeployment) updatePublishedPorts(t testing.TestingT) error {
	var ports []uint32
	for port := range j.ports {
		ports = append(ports, port)
	}
	publishedPorts, err := framework.GetPublishedDockerPorts(t, j.container, ports)
	if err != nil {
		return err
	}
	j.ports = publishedPorts
	return nil
}

func (u *UniversalDeployment) Delete(cluster framework.Cluster) error {
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

func (u *UniversalDeployment) GetExternalAppAddress() string {
	return u.ip
}

func (u *UniversalDeployment) GetExternalPort(internalPort uint32) (uint32, bool) {
	port, ok := u.ports[internalPort]
	return port, ok
}

func (u *UniversalDeployment) GetCert() string {
	return u.cert
}
