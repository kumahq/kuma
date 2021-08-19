package framework

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	util_net "github.com/kumahq/kuma/pkg/util/net"
)

type AppMode string

const (
	AppModeCP              = "kuma-cp"
	AppIngress             = "ingress"
	AppModeEchoServer      = "echo-server"
	AppModeHttpsEchoServer = "https-echo-server"
	sshPort                = "22"

	IngressDataplaneOldType = `
type: Dataplane
mesh: %s
name: dp-ingress
networking:
  address: {{ address }}
  ingress:
    publicAddress: %s
    publicPort: %d
  inbound:
  - port: %d
    tags:
      kuma.io/service: ingress
`
	ZoneIngress = `
type: ZoneIngress
name: ingress
networking:
  address: {{ address }}
  advertisedAddress: %s
  advertisedPort: %d
  port: %d
`

	EchoServerDataplane = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      team: server-owners
      version: %s
`

	EchoServerDataplaneWithServiceProbe = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    serviceProbe:
      tcp: {}
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      team: server-owners
      version: %s
`

	EchoServerDataplaneTransparentProxy = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      team: server-owners
      version: %s
  transparentProxying:
    redirectPortInbound: %s
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`

	AppModeDemoClient   = "demo-client"
	DemoClientDataplane = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      team: client-owners
  outbound:
  - port: 4000
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 4001
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 5000
    tags:
      kuma.io/service: external-service
`

	DemoClientDataplaneWithServiceProbe = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    serviceProbe:
      tcp: {}
    tags:
      kuma.io/service: %s
      team: client-owners
  outbound:
  - port: 4000
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 4001
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 5000
    tags:
      kuma.io/service: external-service
`

	DemoClientDataplaneTransparentProxy = `
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    tags:
      kuma.io/service: %s
      team: client-owners
  transparentProxying:
    redirectPortInbound: %s
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`
)

var defaultDockerOptions = docker.RunOptions{
	Command: nil,
	Detach:  true,
	//Entrypoint:           "",
	EnvironmentVariables: nil,
	Init:                 false,
	//Name:                 "",
	Privileged: false,
	Remove:     true,
	Tty:        false,
	//User:                 "",
	Volumes:      nil,
	OtherOptions: []string{},
	Logger:       nil,
}

type UniversalApp struct {
	t            testing.TestingT
	mainApp      *SshApp
	mainAppEnv   []string
	mainAppArgs  []string
	dpApp        *SshApp
	ports        map[string]string
	lastUsedPort uint32
	container    string
	ip           string
	verbose      bool
}

func NewUniversalApp(t testing.TestingT, clusterName, dpName string, mode AppMode, isipv6, verbose bool, caps []string) (*UniversalApp, error) {
	app := &UniversalApp{
		t:            t,
		ports:        map[string]string{},
		lastUsedPort: 10204,
		verbose:      verbose,
	}

	app.allocatePublicPortsFor("22")

	if mode == AppModeCP {
		app.allocatePublicPortsFor("5678", "5680", "5681", "5682", "5685")
	}

	opts := defaultDockerOptions
	opts.OtherOptions = append(opts.OtherOptions, "--name", clusterName+"_"+dpName)
	for _, cap := range caps {
		opts.OtherOptions = append(opts.OtherOptions, "--cap-add", cap)
	}
	opts.OtherOptions = append(opts.OtherOptions, "--network", "kind")
	if !isipv6 {
		// For now supporting mixed environments with IPv4 and IPv6 addresses is challenging, specifically with
		// builtin DNS. This is due to our mix of CoreDNS and Envoy DNS architecture.
		// Here we make sure the IPv6 address is not allocated to the container unless explicitly requested.
		opts.OtherOptions = append(opts.OtherOptions, "--sysctl", "net.ipv6.conf.all.disable_ipv6=1")
	}
	opts.OtherOptions = append(opts.OtherOptions, app.publishPortsForDocker()...)
	container, err := docker.RunAndGetIDE(t, GetUniversalImage(), &opts)
	if err != nil {
		return nil, err
	}

	app.container = container

	retry.DoWithRetry(app.t, "get IP "+app.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			app.ip, err = app.getIP(IsIPv6())
			if err != nil {
				return "Unable to get Container IP", err
			}
			return "Success", nil
		})

	fmt.Printf("Node IP %s\n", app.ip)

	return app, nil
}

func (s *UniversalApp) allocatePublicPortsFor(ports ...string) {
	for _, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", s.lastUsedPort+1, 11204)
		if err != nil {
			panic(err)
		}
		s.ports[port] = strconv.Itoa(int(pubPortUInt32))
		s.lastUsedPort = pubPortUInt32
	}
}

func (s *UniversalApp) publishPortsForDocker() (args []string) {
	for port, pubPort := range s.ports {
		args = append(args, "--publish="+pubPort+":"+port)
	}
	return
}

func (s *UniversalApp) GetPublicPort(port string) string {
	return s.ports[port]
}

func (s *UniversalApp) Stop() error {
	out, err := docker.StopE(s.t, []string{s.container}, &docker.StopOptions{Time: 1})
	if err != nil {
		return errors.Wrapf(err, "Returned %s", out)
	}

	retry.DoWithRetry(s.t, "stop "+s.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			_, err := docker.StopE(s.t, []string{s.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})

	return nil
}

func (s *UniversalApp) ReStart() error {
	if err := s.mainApp.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.mainApp.cmd.Process.Wait(); err != nil {
		return err
	}

	s.CreateMainApp(s.mainAppEnv, s.mainAppArgs)

	if err := s.mainApp.Start(); err != nil {
		return err
	}
	return nil
}

func (s *UniversalApp) CreateMainApp(env []string, args []string) {
	s.mainAppEnv = env
	s.mainAppArgs = args
	s.mainApp = NewSshApp(s.verbose, s.ports[sshPort], env, args)
}

func (s *UniversalApp) OverrideDpVersion(version string) error {
	// It is important to store installation package in /tmp/kuma/, not /tmp/ otherwise root was taking over /tmp/ and Kuma DP could not store /tmp files
	err := NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"wget",
		fmt.Sprintf("https://download.konghq.com/mesh-alpine/kuma-%s-ubuntu-amd64.tar.gz", version),
		"-O",
		fmt.Sprintf("/tmp/kuma-%s-ubuntu-amd64.tar.gz", version),
	}).Run()
	if err != nil {
		return err
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"mkdir",
		"-p",
		"/tmp/kuma/",
	}).Run()
	if err != nil {
		return err
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"tar",
		"xvzf",
		fmt.Sprintf("/tmp/kuma-%s-ubuntu-amd64.tar.gz", version),
		"-C",
		"/tmp/kuma/",
	}).Run()
	if err != nil {
		return err
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"cp",
		fmt.Sprintf("/tmp/kuma/kuma-%s/bin/kuma-dp", version),
		"/usr/bin/kuma-dp",
	}).Run()
	if err != nil {
		return err
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"cp",
		fmt.Sprintf("/tmp/kuma/kuma-%s/bin/envoy", version),
		"/usr/local/bin/envoy",
	}).Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *UniversalApp) CreateDP(token, cpAddress, name, mesh, ip, dpyaml string, builtindns, ingress bool) {
	// create the token file on the app container
	err := NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + token + "\"", ">", "/kuma/token-" + name}).Run()
	if err != nil {
		panic(err)
	}

	// run the DP as user `envoy` so iptables can distinguish its traffic if needed
	args := []string{
		"runuser", "-u", "kuma-dp", "--",
		"/usr/bin/kuma-dp", "run",
		"--cp-address=" + cpAddress,
		"--dataplane-token-file=/kuma/token-" + name,
		"--binary-path", "/usr/local/bin/envoy",
	}

	if dpyaml != "" {
		err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + dpyaml + "\"", ">", "/kuma/dpyaml-" + name}).Run()
		if err != nil {
			panic(err)
		}
		args = append(args,
			"--dataplane-file=/kuma/dpyaml-"+name,
			"--dataplane-var", "name="+name,
			"--dataplane-var", "address="+ip)
	} else {
		args = append(args,
			"--name="+name,
			"--mesh="+mesh)
	}

	if builtindns {
		args = append(args, "--dns-enabled")
	}
	if ingress {
		args = append(args, "--proxy-type=ingress")
	}
	s.dpApp = NewSshApp(s.verbose, s.ports[sshPort], []string{}, args)
}

func (s *UniversalApp) setupTransparent(cpIp string, builtindns bool) {
	args := []string{
		"/usr/bin/kumactl", "install", "transparent-proxy",
		"--kuma-dp-user", "kuma-dp",
		"--kuma-cp-ip", cpIp,
	}

	if builtindns {
		args = append(args,
			"--skip-resolv-conf",
			"--redirect-dns",
			"--redirect-dns-upstream-target-chain", "DOCKER_OUTPUT")
	}

	app := NewSshApp(s.verbose, s.ports[sshPort], []string{}, args)
	err := app.Run()
	if err != nil {
		panic(fmt.Sprintf("err: %s\nstderr :%s\nstdout %s", err.Error(), app.Err(), app.Out()))
	}
}

func (s *UniversalApp) getIP(isipv6 bool) (string, error) {
	cmd := SshCmd(s.ports[sshPort], []string{}, []string{"getent", "ahosts", s.container[:12]})
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "invalid", errors.Wrapf(err, "getent failed with %s", string(bytes))
	}
	lines := strings.Split(string(bytes), "\n")
	// search for the requested IP
	for _, line := range lines {
		split := strings.Split(line, " ")
		ip := split[0]
		if isipv6 {
			if govalidator.IsIPv6(ip) {
				return ip, nil
			}
		} else if govalidator.IsIPv4(ip) {
			return ip, nil
		}
	}
	errString := "No IPv4 address found"
	if isipv6 {
		errString = "No IPv6 address found"
	}
	return "", errors.Errorf(errString)
}

type SshApp struct {
	cmd    *exec.Cmd
	stdin  bytes.Buffer
	stdout bytes.Buffer
	stderr bytes.Buffer
	port   string
}

func NewSshApp(verbose bool, port string, env []string, args []string) *SshApp {
	app := &SshApp{
		port: port,
	}
	app.cmd = app.SshCmd(env, args)

	inWriters := []io.Reader{&app.stdin}
	outWriters := []io.Writer{&app.stdout}
	errWriters := []io.Writer{&app.stderr}
	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}
	app.cmd.Stdout = io.MultiWriter(outWriters...)
	app.cmd.Stderr = io.MultiWriter(errWriters...)
	app.cmd.Stdin = io.MultiReader(inWriters...)
	return app
}

func (s *SshApp) Run() error {
	fmt.Printf("Running %v\n", s.cmd)
	return s.cmd.Run()
}

func (s *SshApp) Start() error {
	fmt.Printf("Starting %v\n", s.cmd)
	return s.cmd.Start()
}

func (s *SshApp) Stop() error {
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.cmd.Process.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *SshApp) Wait() error {
	return s.cmd.Wait()
}

func (s *SshApp) Out() string {
	return s.stdout.String()
}

func (s *SshApp) Err() string {
	return s.stderr.String()
}

func (s *SshApp) SshCmd(env []string, args []string) *exec.Cmd {
	return SshCmd(s.port, env, args)
}

func SshCmd(port string, env []string, args []string) *exec.Cmd {
	sshArgs := append([]string{
		"-q", "-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"root@localhost", "-p", port}, env...)
	sshArgs = append(sshArgs, args...)

	cmd := exec.Command("ssh", sshArgs...)
	return cmd
}
