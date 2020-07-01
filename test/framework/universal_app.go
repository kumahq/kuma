package framework

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/testing"

	util_net "github.com/Kong/kuma/pkg/util/net"
)

type AppMode string

const (
	AppModeCP           = "kuma-cp"
	AppModeEchoServer   = "echo-server"
	EchoServerDataplane = `
type: Dataplane
mesh: default
name: %s
networking:
  address: %s
  inbound:
  - port: %s
    servicePort: %s
    tags:
      service: echo-server
      protocol: http
`
	AppModeDemoClient   = "demo-client"
	DemoClientDataplane = `
type: Dataplane
mesh: default
name: %s
networking:
  address: %s
  inbound:
  - port: %s
    servicePort: %s
    tags:
      service: demo-client
  outbound:
  - port: 4000
    service: echo-server
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
	t         testing.TestingT
	mainApp   *sshApp
	dpApp     *sshApp
	sshPort   string
	container string
	ip        string
	verbose   bool
}

func NewUniversalApp(t testing.TestingT, mode AppMode, verbose bool, env []string, args []string) *UniversalApp {
	port, err := util_net.PickTCPPort("", 10240, 11204)
	if err != nil {
		panic(err)
	}
	sshPort := strconv.Itoa(int(port))

	opts := defaultDockerOptions
	opts.OtherOptions = append(opts.OtherOptions, "--publish="+sshPort+":22")

	if mode == AppModeCP {
		opts.OtherOptions = append(opts.OtherOptions,
			"--publish=5678:5678",
			"--publish=5679:5679",
			"--publish=5680:5680",
			"--publish=5681:5681",
			"--publish=5682:5682",
			"--publish=5683:5683",
			"--publish=5684:5684",
		)
	}
	container := docker.RunAndGetID(t, kumaUniversalImage, &opts)

	app := &UniversalApp{
		t:         t,
		sshPort:   sshPort,
		container: container,
		verbose:   verbose,
	}

	app.ip = app.getIP()
	fmt.Printf("Node IP %s\n", app.ip)

	if mode == AppModeCP {
		args = append([]string{"KUMA_GENERAL_ADVERTISED_HOSTNAME=" + app.ip}, args...)
	}

	app.CreateMainApp(env, args)

	return app
}

func (s *UniversalApp) Stop() error {
	docker.Stop(s.t, []string{s.container}, &docker.StopOptions{})
	return nil
}

func (s *UniversalApp) ReStart() error {
	if err := s.mainApp.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.mainApp.cmd.Process.Wait(); err != nil {
		return err
	}
	if err := s.mainApp.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (s *UniversalApp) CreateMainApp(env []string, args []string) {
	s.mainApp = NewSshApp(s.verbose, s.sshPort, env, args)
}

func (s *UniversalApp) CreateDP(token, cpIP, appname string) {
	// and echo it to the Application Node
	err := NewSshApp(s.verbose, s.sshPort, []string{}, []string{"echo", token, ">", "/kuma/token-"+appname}).Run()
	if err != nil {
		panic(err)
	}

	// run the DP
	s.dpApp = NewSshApp(s.verbose, s.sshPort, []string{}, []string{"kuma-dp", "run",
		"--name=dp-"+appname,
		"--mesh=default",
		"--cp-address=http://"+cpIP+":5681",
		"--dataplane-token-file=/kuma/token-"+appname,
		"--binary-path", "/usr/local/bin/envoy"})
}

func (s *UniversalApp) getIP() string {
	cmd := SshCmd(s.sshPort, []string{}, []string{"getent", "hosts", s.container[:12]})
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		panic(string(bytes))
	}
	split := strings.Split(string(bytes), " ")
	return split[0]
}

type sshApp struct {
	cmd    *exec.Cmd
	stdout bytes.Buffer
	stderr bytes.Buffer
	port   string
}

func NewSshApp(verbose bool, port string, env []string, args []string) *sshApp {
	app := &sshApp{
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

func (s *sshApp) Run() error {
	fmt.Printf("Running %v\n", s.cmd)
	return s.cmd.Run()
}

func (s *sshApp) Start() error {
	fmt.Printf("Starting %v\n", s.cmd)
	return s.cmd.Start()
}

func (s *sshApp) Stop() error {
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.cmd.Process.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *sshApp) Wait() error {
	return s.cmd.Wait()
}

func (s *sshApp) Out() string {
	return s.stdout.String()
}

func (s *sshApp) Err() string {
	return s.stderr.String()
}

func (s *sshApp) SshCmd(env []string, args []string) *exec.Cmd {
	return SshCmd(s.port, env, args)
}

func SshCmd(port string, env []string, args []string) *exec.Cmd {
	sshArgs := append([]string{
		"-q",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"root@localhost", "-p", port}, env...)
	sshArgs = append(sshArgs, args...)

	cmd := exec.Command("ssh", sshArgs...)
	return cmd
}
