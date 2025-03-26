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
	_ = s.KillMainApp()
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
) error {
	cmd := &strings.Builder{}

	workingDir := fmt.Sprintf("/tmp/kuma")
	// create the token file on the app container
	_, _ = cmd.WriteString("#!/bin/sh\n")
	_, _ = fmt.Fprintf(cmd, "mkdir -p %s/bin\n", workingDir)
	_, _ = fmt.Fprintf(cmd, "printf %q > %s/token-%s\n", token, workingDir, name)

	_, _ = fmt.Fprintf(cmd, `
curl -L --no-progress-bar --fail '%s' | VERSION=%s sh
DOWNLOAD_DIR=$(find -maxdepth 1 -type d -name '*-%s')
mv -f $DOWNLOAD_DIR/bin/* %s/bin/
export PATH=$PATH:%s/bin
		`, Config.KumaInstallerUrl, Config.KumaImageTag, Config.KumaImageTag, workingDir, workingDir)

	for k, v := range envsMap {
		_, _ = fmt.Fprintf(cmd, "export %s=%s\n", k, utils.ShellEscape(v))
	}

	// todo: remove sudo prefix to adapt more distros

	// Install transparent proxy
	if transparent {
		extraArgs := []string{}
		if builtindns {
			extraArgs = append(extraArgs, "--redirect-dns")
		}
		_, _ = fmt.Fprintf(cmd, "sudo %s/bin/kumactl install transparent-proxy --exclude-inbound-ports %s %s\n",
			workingDir, sshPort, strings.Join(extraArgs, " "))
		_, _ = fmt.Fprintf(cmd, `function uninstall_transparent_proxy(){
    sudo %s/bin/kumactl install transparent-proxy 
}
trap uninstall_transparent_proxy EXIT INT QUIT TERM
\n`,
			workingDir)
	}

	_, _ = fmt.Fprintf(cmd, "sudo useradd --system --no-create-home --shell /sbin/nologin kuma-dp\n")
	// run the DP as user `envoy` so iptables can distinguish its traffic if needed
	args := []string{
		"sudo", "runuser", "-u", "kuma-dp", "--",
		"kuma-dp", "run",
		"--cp-address=" + cpAddress,
		fmt.Sprintf("--dataplane-token-file=%s/kuma/token-%s", workingDir, name),
		"--binary-path", fmt.Sprintf("%s/bin/envoy", workingDir),
	}

	if dpyaml != "" {
		dpPath := fmt.Sprintf("%s/kuma-dp-%s.yaml", workingDir, name)
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
