package framework

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/asaskevich/govalidator"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal_logs"
)

type AppMode string

const (
	AppModeCP              = "kuma-cp"
	AppIngress             = "ingress"
	AppEgress              = "egress"
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
name: %s
networking:
  address: {{ address }}
  advertisedAddress: %s
  advertisedPort: %d
  port: %d
`
	ZoneEgress = `
type: ZoneEgress
name: egress
networking:
  address: {{ address }}
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
    redirectPortOutbound: %s
`

	AppModeTcpSink      = "tcp-sink"
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
%s
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
%s
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
%s
  transparentProxying:
    redirectPortInbound: %s
    redirectPortOutbound: %s
    reachableServices: [%s]
`
)

type UniversalApp struct {
	t             testing.TestingT
	logsPath      string
	mainApp       *ssh.App
	mainAppEnv    map[string]string
	mainAppArgs   []string
	dpApp         *ssh.App
	dpEnv         map[string]string
	ports         map[string]string
	container     string
	containerName string
	ip            string
	verbose       bool
	mesh          string
	concurrency   int
}

func NewUniversalApp(t testing.TestingT, clusterName, dpName, mesh string, mode AppMode, isipv6, verbose bool, caps, volumes []string, containerName string, concurrency int) (*UniversalApp, error) {
	app := &UniversalApp{
		t:             t,
		logsPath:      universal_logs.CurrentLogsPath(Config.UniversalE2ELogsPath),
		ports:         map[string]string{},
		verbose:       verbose,
		mesh:          mesh,
		containerName: fmt.Sprintf("%s_%s_%s", clusterName, dpName, random.UniqueId()),
		concurrency:   concurrency,
	}
	if containerName != "" {
		app.containerName = containerName
	}

	app.allocatePublicPortsFor("22")

	if mode == AppModeCP {
		app.allocatePublicPortsFor("5678", "5680", "5681", "5682", "5685", "9901")
	}

	if dpName == AppEgress {
		app.allocatePublicPortsFor("9901")
	}

	dockerExtraOptions := []string{
		"--network", "kind",
	}
	dockerExtraOptions = append(dockerExtraOptions, app.publishPortsForDocker()...)
	if !isipv6 {
		// For now supporting mixed environments with IPv4 and IPv6 addresses is challenging, specifically with
		// builtin DNS. This is due to our mix of CoreDNS and Envoy DNS architecture.
		// Here we make sure the IPv6 address is not allocated to the container unless explicitly requested.
		dockerExtraOptions = append(dockerExtraOptions, "--sysctl", "net.ipv6.conf.all.disable_ipv6=1")
	}
	for _, c := range caps {
		dockerExtraOptions = append(dockerExtraOptions, "--cap-add", c)
	}
	container, err := docker.RunAndGetIDE(t, Config.GetUniversalImage(), &docker.RunOptions{
		Detach:       true,
		Remove:       true,
		Name:         app.containerName,
		Volumes:      volumes,
		OtherOptions: dockerExtraOptions,
	})
	if err != nil {
		return nil, err
	}

	app.container = container

	if err := app.updatePublishedPorts(); err != nil {
		return nil, err
	}

	retry.DoWithRetry(app.t, "get IP "+app.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			app.ip, err = app.getIP(Config.IPV6)
			if err != nil {
				return "Unable to get Container IP", err
			}
			return "Success", nil
		})

	Logf("Node IP %s", app.ip)

	return app, nil
}

func (s *UniversalApp) updatePublishedPorts() error {
	var ports []uint32
	for portStr := range s.ports {
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			return err
		}
		ports = append(ports, uint32(port))
	}

	publishedPorts, err := GetPublishedDockerPorts(s.t, s.container, ports)
	if err != nil {
		return err
	}

	for _, port := range ports {
		publishedPort := publishedPorts[port]
		s.ports[strconv.Itoa(int(port))] = strconv.Itoa(int(publishedPort))
	}
	return nil
}

func (s *UniversalApp) allocatePublicPortsFor(ports ...string) {
	for _, port := range ports {
		s.ports[port] = "0"
	}
}

func (s *UniversalApp) publishPortsForDocker() []string {
	// If we aren't using IPv6 in the container then we only want to listen on
	// IPv4 interfaces to prevent resolving 'localhost' to the IPv6 address of
	// the container and having the container not respond.
	// We only use the ipv4 address to access the container services
	// even if ipv6 is enabled in the container to prevent port number conflicts
	// between ipv4 and ipv6 addresses.
	ipv4Ip := "0.0.0.0::"
	var args []string
	for port := range s.ports {
		args = append(args, "--publish="+ipv4Ip+port)
	}
	return args
}

func (s *UniversalApp) GetPublicPort(port string) string {
	return s.ports[port]
}

func (s *UniversalApp) GetContainerName() string {
	return s.containerName
}

func (s *UniversalApp) GetEnvoyAdminTunnel() (envoy_admin.Tunnel, error) {
	t, err := tunnel.NewUniversalEnvoyAdminTunnel(s.t, s.ports["22"], s.verbose)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *UniversalApp) GetIP() string {
	return s.ip
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
	if err := s.KillMainApp(); err != nil {
		return err
	}
	return s.StartMainApp()
}

func (s *UniversalApp) KillMainApp() error {
	return s.mainApp.Signal(syscall.SIGKILL, true)
}

func (s *UniversalApp) StartMainApp() error {
	s.CreateMainApp(s.mainAppEnv, s.mainAppArgs)

	return s.mainApp.Start()
}

func (s *UniversalApp) CreateMainApp(env map[string]string, args []string) {
	s.mainAppEnv = env
	s.mainAppArgs = args
	s.mainApp = ssh.NewApp(s.containerName, s.logsPath, s.verbose, s.ports[sshPort], env, args)
}

func (s *UniversalApp) OverrideDpVersion(version string) error {
	// It is important to store installation package in /tmp/kuma/, not /tmp/ otherwise root was taking over /tmp/ and Kuma DP could not store /tmp files
	err := ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{
		"wget",
		fmt.Sprintf("https://packages.konghq.com/public/kuma-binaries-release/raw/names/kuma-linux-%[2]s/versions/%[1]s/kuma-%[1]s-linux-%[2]s.tar.gz", version, Config.Arch),
		"-O",
		fmt.Sprintf("/tmp/kuma-%s-ubuntu-amd64.tar.gz", version),
	}).Run()
	if err != nil {
		return err
	}

	err = ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{
		"mkdir",
		"-p",
		"/tmp/kuma/",
	}).Run()
	if err != nil {
		return err
	}

	err = ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{
		"tar",
		"xvzf",
		fmt.Sprintf("/tmp/kuma-%s-ubuntu-amd64.tar.gz", version),
		"-C",
		"/tmp/kuma/",
	}).Run()
	if err != nil {
		return err
	}

	err = ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{
		"cp",
		fmt.Sprintf("/tmp/kuma/kuma-%s/bin/kuma-dp", version),
		"/usr/bin/kuma-dp",
	}).Run()
	if err != nil {
		return err
	}

	err = ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{
		"cp",
		fmt.Sprintf("/tmp/kuma/kuma-%s/bin/envoy", version),
		"/usr/local/bin/envoy",
	}).Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *UniversalApp) CreateDP(
	token, cpAddress, name, mesh, ip, dpyaml string,
	builtindns bool,
	proxyType string,
	concurrency int,
	envsMap map[string]string,
) {
	// create the token file on the app container
	err := ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{"printf ", "\"" + token + "\"", ">", "/kuma/token-" + name}).Run()
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
		err = ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{"printf ", "\"" + dpyaml + "\"", ">", "/kuma/dpyaml-" + name}).Run()
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

	if concurrency > 0 {
		args = append(args, "--concurrency", strconv.Itoa(concurrency))
	}

	if builtindns {
		args = append(args, "--dns-enabled")
	}

	if proxyType != "" {
		args = append(args, "--proxy-type", proxyType)
	}

	s.dpApp = ssh.NewApp(s.containerName, s.logsPath, s.verbose, s.ports[sshPort], envsMap, args)
}

func (s *UniversalApp) setupTransparent(builtindns bool) {
	args := []string{
		"/usr/bin/kumactl", "install", "transparent-proxy",
		"--kuma-dp-user", "kuma-dp",
		"--skip-dns-conntrack-zone-split",
		"--exclude-inbound-ports", "22",
	}

	if builtindns {
		args = append(args,
			"--redirect-dns",
		)
	}

	app := ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, args)
	err := app.Run()
	if err != nil {
		panic(fmt.Sprintf("err: %s\nstderr :%s\nstdout %s", err.Error(), app.Err(), app.Out()))
	}
}

func (s *UniversalApp) getIP(isipv6 bool) (string, error) {
	cmd := ssh.NewApp(s.containerName, "", s.verbose, s.ports[sshPort], nil, []string{"getent", "ahosts", s.container[:12]})
	err := cmd.Run()
	if err != nil {
		return "invalid", errors.Wrapf(err, "getent failed with %s", cmd.Err())
	}
	lines := strings.Split(cmd.Out(), "\n")
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
	return "", errors.New(errString)
}
