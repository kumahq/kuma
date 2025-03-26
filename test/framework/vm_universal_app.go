package framework

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/testing"
	"golang.org/x/crypto/ssh"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	kssh "github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal"
	"github.com/kumahq/kuma/test/framework/utils"
)

type VmUniversalApp struct {
	t          testing.TestingT
	mainApp    *kssh.Session
	mainAppCmd string
	dpApp      *kssh.Session
	dpAppCmd   string

	verbose bool
	mesh    string

	universalNetworking *universal.Networking
	appName             string
	logger              *logger.Logger
}

func NewVmUniversalApp(t testing.TestingT, appName, mesh string, host *VmClusterHost, verbose bool) (*VmUniversalApp, error) {
	app := &VmUniversalApp{
		t:       t,
		verbose: verbose,
		mesh:    mesh,
		appName: appName,
	}

	app.universalNetworking = &universal.Networking{
		IP:         host.InternalIPAddress,
		RemoteHost: &host.Host,
	}

	Logf("App %s running on Node IP %s", appName, app.universalNetworking.IP)

	return app, nil
}

func (s *VmUniversalApp) GetEnvoyAdminTunnel() envoy_admin.Tunnel {
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

func (s *VmUniversalApp) GetIP() string {
	return s.universalNetworking.IP
}

func (s *VmUniversalApp) Stop() error {
	Logf("Stopping app:%q", s.appName)
	_ = s.universalNetworking.Close()
	return fmt.Errorf("timed out waiting for app:%q to stop", s.appName)
}

func (s *VmUniversalApp) ReStart() error {
	Logf("Restarting app:%q", s.appName)
	if err := s.KillMainApp(); err != nil {
		return err
	}
	// No needed but this just in case kill -9 is not instant
	time.Sleep(1 * time.Second)
	return s.StartMainApp()
}

func (s *VmUniversalApp) KillMainApp() error {
	defer s.mainApp.Close()
	err := s.mainApp.Signal(ssh.SIGKILL, false)
	if err != nil {
		return err
	}
	return nil
}

func (s *VmUniversalApp) StartMainApp() error {
	Logf("Starting app:%q", s.appName)
	s.CreateMainApp(s.mainAppCmd)

	return s.mainApp.Start()
}

func (s *VmUniversalApp) CreateMainApp(cmd string) {
	s.mainAppCmd = cmd
	var err error
	s.mainApp, err = s.newSession(s.appName, s.mainAppCmd)
	if err != nil {
		panic(err)
	}
}

func (s *VmUniversalApp) CreateDP(
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

func (s *VmUniversalApp) newSession(name string, cmd string) (*kssh.Session, error) {
	return s.universalNetworking.NewSession(path.Join("vm-universal", "exec"), name, s.verbose, cmd)
}
