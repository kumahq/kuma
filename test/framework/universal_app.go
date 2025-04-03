package framework

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	kssh "github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal"
	"github.com/kumahq/kuma/test/framework/utils"
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
	t                   testing.TestingT
	mainApp             *kssh.Session
	mainAppCmd          string
	dpApp               *kssh.Session
	dpAppCmd            string
	ports               map[string]string
	container           string
	containerName       string
	verbose             bool
	mesh                string
	concurrency         int
	clusterName         string
	universalNetworking *universal.Networking
	appName             string
	logger              *logger.Logger
	dockerBackend       DockerBackend
}

func NewUniversalApp(t testing.TestingT, dockerBackend DockerBackend, clusterName, appName, mesh string, mode AppMode, isipv6, verbose bool, caps, volumes []string, containerName string, concurrency int) (*UniversalApp, error) {
	app := &UniversalApp{
		t:             t,
		ports:         map[string]string{},
		verbose:       verbose,
		mesh:          mesh,
		appName:       appName,
		containerName: fmt.Sprintf("%s_%s_%s", clusterName, appName, random.UniqueId()),
		concurrency:   concurrency,
		clusterName:   clusterName,
		dockerBackend: dockerBackend,
	}
	if containerName != "" {
		app.containerName = containerName
	}

	app.allocatePublicPortsFor("22")

	if mode == AppModeCP {
		app.allocatePublicPortsFor("5678", "5680", "5681", "5682", "5685")
	}

	dockerExtraOptions := []string{
		"--network", DockerNetworkName,
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
	app.logger = logger.Discard
	if verbose {
		app.logger = logger.Default
	}
	container, err := app.dockerBackend.RunAndGetIDE(t, Config.GetUniversalImage(), &docker.RunOptions{
		Detach:       true,
		Remove:       true,
		Name:         app.containerName,
		Volumes:      volumes,
		OtherOptions: dockerExtraOptions,
		Logger:       app.logger,
	})
	if err != nil {
		return nil, err
	}

	app.container = container

	if err := app.updatePublishedPorts(); err != nil {
		return nil, err
	}
	app.universalNetworking = &universal.Networking{
		ApiServerPort: app.ports["5681"],
		SshPort:       app.ports[sshPort],
	}
	app.universalNetworking.IP, err = app.getIP(Config.IPV6)
	if err != nil {
		return nil, fmt.Errorf("unable to get Container IP %w", err)
	}

	Logf("Node IP %s", app.universalNetworking.IP)

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

	publishedPorts, err := s.dockerBackend.GetPublishedDockerPorts(s.t, s.logger, s.container, ports)
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

func (s *UniversalApp) GetEnvoyAdminTunnel() envoy_admin.Tunnel {
	return tunnel.NewUniversalEnvoyAdminTunnel(func(cmdName, cmd string) (string, error) {
		session, err := s.newSession("envoytunnel"+cmdName, cmd)
		if err != nil {
			return "", err
		}
		err = session.Run()
		if err != nil {
			return "", err
		}
		b, err := os.ReadFile(session.StdOutFile())
		return string(b), err
	})
}

func (s *UniversalApp) GetIP() string {
	return s.universalNetworking.IP
}

func (s *UniversalApp) Stop() error {
	Logf("Stopping app:%q container:%q", s.appName, s.container)
	for i := 0; i < 10; i++ {
		_, err := s.dockerBackend.StopE(s.t, []string{s.container}, &docker.StopOptions{Time: 1, Logger: s.logger})
		if err != nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	_ = s.universalNetworking.Close()
	return fmt.Errorf("timed out waiting for app:%q container:%q to stop", s.appName, s.container)
}

func (s *UniversalApp) ReStart() error {
	Logf("Restarting app:%q container:%q", s.appName, s.container)
	if err := s.KillMainApp(); err != nil {
		return err
	}
	// No needed but this just in case kill -9 is not instant
	time.Sleep(1 * time.Second)
	return s.StartMainApp()
}

func (s *UniversalApp) KillMainApp() error {
	defer s.mainApp.Close()
	err := s.mainApp.Signal(ssh.SIGKILL, false)
	if err != nil {
		return err
	}
	return nil
}

func (s *UniversalApp) StartMainApp() error {
	Logf("Starting app:%q container:%q", s.appName, s.container)
	s.CreateMainApp(s.mainAppCmd)

	return s.mainApp.Start()
}

func (s *UniversalApp) CreateMainApp(cmd string) {
	s.mainAppCmd = cmd
	var err error
	s.mainApp, err = s.newSession(s.appName, s.mainAppCmd)
	if err != nil {
		panic(err)
	}
}

func (s *UniversalApp) CreateDP(
	token, cpAddress, name, mesh, ip, dpyaml string,
	builtindns bool,
	proxyType string,
	concurrency int,
	envsMap map[string]string,
	transparent bool,
	dpVersion string,
) error {
	cmd := &strings.Builder{}
	// create the token file on the app container
	_, _ = cmd.WriteString("#!/bin/sh\n")
	_, _ = fmt.Fprintf(cmd, "printf %q > /kuma/token-%s\n", token, name)
	if dpVersion != "" {
		// It is important to store installation package in /tmp/kuma/, not /tmp/ otherwise root was taking over /tmp/ and Kuma DP could not store /tmp files
		url := fmt.Sprintf("https://packages.konghq.com/public/kuma-binaries-release/raw/names/kuma-linux-%[2]s/versions/%[1]s/kuma-%[1]s-linux-%[2]s.tar.gz", dpVersion, Config.Arch)
		newPathOut := fmt.Sprintf("/tmp/kuma/kuma-%s/bin", dpVersion)

		_, _ = fmt.Fprintf(cmd, `
mkdir -p /tmp/
curl --no-progress-bar --fail '%s' | tar xvzf - --directory /tmp/kuma/
cp %s/kuma-dp /usr/bin/kuma-dp
cp %s/envoy /usr/bin/envoy
		`, url, newPathOut, newPathOut)
	}
	for k, v := range envsMap {
		_, _ = fmt.Fprintf(cmd, "export %s=%s\n", k, utils.ShellEscape(v))
	}

	// Install transparent proxy
	if transparent {
		extraArgs := []string{}
		if builtindns {
			extraArgs = append(extraArgs, "--redirect-dns")
		}
		_, _ = fmt.Fprintf(cmd, "/usr/bin/kumactl install transparent-proxy --exclude-inbound-ports %s %s\n", sshPort, strings.Join(extraArgs, " "))
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
		dpPath := fmt.Sprintf("/kuma-dp-%s.yaml", name)
		_, _ = fmt.Fprintf(cmd, `cat > %s << 'EOF'
%s
EOF
`, dpPath, dpyaml)

		args = append(args,
			"--dataplane-file="+dpPath,
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

	if Config.Debug {
		args = append(args, "--log-level", "debug")
	}
	_, _ = cmd.WriteString(strings.Join(args, " "))
	s.dpAppCmd = cmd.String()
	var err error
	s.dpApp, err = s.newSession("dp-"+name, s.dpAppCmd)
	return err
}

func (s *UniversalApp) getIP(isipv6 bool) (string, error) {
	stdout, stderr, err := s.universalNetworking.RunCommand(fmt.Sprintf(`
until getent ahosts %q |cut -d" " -f1|sort|uniq; do
	echo "Waiting for getent to return something..."
	sleep 0.5
done
`, s.container[:12]))
	if err != nil {
		return "", errors.Wrapf(err, "cmd failed with %s stderr:%q stdout:%q", err, stderr, stdout)
	}
	// get the first line of the output
	for _, ipStr := range strings.Split(stdout, "\n") {
		ip := strings.TrimSpace(ipStr)
		if isipv6 && govalidator.IsIPv6(ip) {
			return ip, nil
		}
		if !isipv6 && govalidator.IsIPv4(ip) {
			return ip, nil
		}
	}
	return "", errors.Errorf("couldn't find a valid IP address usingV6=%v output=%q", isipv6, stdout)
}

func (s *UniversalApp) newSession(name string, cmd string) (*kssh.Session, error) {
	return s.universalNetworking.NewSession(path.Join(s.clusterName, "universal", "exec", s.containerName), name, s.verbose, cmd)
}
