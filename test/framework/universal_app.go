package framework

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/testing"

	util_net "github.com/kumahq/kuma/pkg/util/net"
)

type AppMode string

const (
	AppModeCP              = "kuma-cp"
	AppIngress             = "ingress"
	AppModeEchoServer      = "echo-server"
	AppModeHttpsEchoServer = "https-echo-server"
	sshPort                = "22"

	IngressDataplane = `
type: Dataplane
mesh: default
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
	EchoServerDataplane = `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
      kuma.io/protocol: http
`
	EchoServerDataplaneTransparentProxy = `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
      kuma.io/protocol: http
  transparentProxying:
    redirectPortInbound: %s
    redirectPortOutbound: %s
`

	AppModeDemoClient   = "demo-client"
	DemoClientDataplane = `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: demo-client
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
mesh: default
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    tags:
      kuma.io/service: demo-client
  transparentProxying:
    redirectPortInbound: %s
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

func NewUniversalApp(t testing.TestingT, clusterName string, mode AppMode, verbose bool, caps []string) (*UniversalApp, error) {
	app := &UniversalApp{
		t:            t,
		ports:        map[string]string{},
		lastUsedPort: 10204,
		verbose:      verbose,
	}

	app.allocatePublicPortsFor("22")

	if mode == AppModeCP {
		app.allocatePublicPortsFor("5678", "5680", "5681", "5682", "5685")
	} else {
		app.allocatePublicPortsFor("30001") // the envoy admin port
	}

	opts := defaultDockerOptions
	opts.OtherOptions = append(opts.OtherOptions, "--name", clusterName+"_"+string(mode))
	for _, cap := range caps {
		opts.OtherOptions = append(opts.OtherOptions, "--cap-add", cap)
	}
	opts.OtherOptions = append(opts.OtherOptions, "--network", "kind")
	opts.OtherOptions = append(opts.OtherOptions, app.publishPortsForDocker()...)
	container, err := docker.RunAndGetIDE(t, KumaUniversalImage, &opts)
	if err != nil {
		return nil, err
	}

	app.container = container

	retry.DoWithRetry(app.t, "get IP "+app.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			app.ip, err = app.getIP()
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

func (s *UniversalApp) CreateDP(token, cpIp, cpAddress, appname, ip, dpyaml string, transparent bool) {
	if transparent {

		//inboundPort := ""
		//switch appname {
		//case AppModeEchoServer:
		//	inboundPort = "8080"
		//case AppModeDemoClient:
		//	inboundPort = "13000"
		//}

		// make sure transparent iptable rules are set accordingly
		err := NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
			//"/root/run-iptables.sh",
			//"\"",
			"/root/kuma-iptables",
			"-p", "15001", // Specify the envoy port to which redirect all TCP traffic (default $ENVOY_PORT = 15001)
			"-z", "15006", // Port to which all inbound TCP traffic to the pod/VM should be redirected to (default $INBOUND_CAPTURE_PORT = 15006)
			"-u", "5678", // Specify the UID of the user for which the redirection is not applied. Typically, this is the UID of the proxy container
			"-g", "5678", // Specify the GID of the user for which the redirection is not applied. (same default value as -u param)
			//"-d", "22", // Comma separated list of inbound ports to be excluded from redirection to Envoy (optional). Only applies  when all inbound traffic (i.e. "*") is being redirected (default to $ISTIO_LOCAL_EXCLUDE_PORTS)
			//"-o", "12345", // Comma separated list of outbound ports to be excluded from redirection to Envoy
			"-m", "REDIRECT", // The mode used to redirect inbound connections to Envoy, either "REDIRECT" or "TPROXY"
			"-i", "'*'", // Comma separated list of IP ranges in CIDR form to redirect to envoy (optional). The wildcard character "*" can be used to redirect all outbound traffic. An empty list will disable all outbound
			"-b", "'*'", // Comma separated list of inbound ports for which traffic is to be redirected to Envoy (optional). The wildcard character "*" can be used to configure redirection for all ports. An empty list will disable
			//"\"",
		}).Run()
		if err != nil {
			panic(err)
		}

		// add kuma-cp nameserver
		err = NewSshApp(s.verbose, s.ports[sshPort], []string{},
			[]string{
				"cp", "/etc/resolv.conf", "/etc/resolv.conf.orig", "&&",
				"printf", "\"nameserver " + cpIp + "\n\"", ">", "/etc/resolv.conf", "&&",
				"cat", "/etc/resolv.conf.orig", ">>", "/etc/resolv.conf",
			}).Run()
		if err != nil {
			panic(err)
		}
	}
	// create the token file on the app container
	err := NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + token + "\"", ">", "/kuma/token-" + appname}).Run()
	if err != nil {
		panic(err)
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + dpyaml + "\"", ">", "/kuma/dpyaml-" + appname}).Run()
	if err != nil {
		panic(err)
	}
	// run the DP
	s.dpApp = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{
		"runuser", "-u", "envoy", "--",
		"/usr/bin/kuma-dp", "run",
		"--cp-address=" + cpAddress,
		"--dataplane-token-file=/kuma/token-" + appname,
		"--dataplane-file=/kuma/dpyaml-" + appname,
		"--dataplane-var", "name=dp-" + appname,
		"--dataplane-var", "address=" + ip,
		"--binary-path", "/usr/local/bin/envoy",
	})
}

func (s *UniversalApp) getIP() (string, error) {
	cmd := SshCmd(s.ports[sshPort], []string{}, []string{"getent", "ahosts", s.container[:12]})
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "invalid", errors.Wrapf(err, "getent failed with %s", string(bytes))
	}
	lines := strings.Split(string(bytes), "\n")
	// search ipv4
	for _, line := range lines {
		split := strings.Split(line, " ")
		testInput := net.ParseIP(split[0])
		if testInput.To4() != nil {
			return split[0], nil
		}
	}
	return "", errors.Errorf("No IPv4 address found")
}

type SshApp struct {
	cmd    *exec.Cmd
	stdout bytes.Buffer
	stderr bytes.Buffer
	port   string
}

func NewSshApp(verbose bool, port string, env []string, args []string) *SshApp {
	app := &SshApp{
		port: port,
	}
	app.cmd = app.SshCmd(env, args)

	outWriters := []io.Writer{&app.stdout}
	errWriters := []io.Writer{&app.stderr}
	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}
	app.cmd.Stdout = io.MultiWriter(outWriters...)
	app.cmd.Stderr = io.MultiWriter(errWriters...)
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
