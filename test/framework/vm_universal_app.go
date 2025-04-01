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

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	kssh "github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal"
	"github.com/kumahq/kuma/test/framework/utils"
)

const (
	vmMainAppContainerName = "main-app"
)

type VmUniversalApp struct {
	t          testing.TestingT
	mainAppCmd string
	dpApp      *kssh.Session
	dpAppCmd   string

	verbose bool
	mesh    string

	universalNetworking *universal.Networking
	appName             string

	mainAppLogger *logger.Logger
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

	app.mainAppLogger = logger.Discard
	if verbose {
		app.mainAppLogger = logger.Default
	}

	Logf("App %s running on Node IP %s", appName, app.universalNetworking.IP)

	return app, nil
}

func (s *VmUniversalApp) GetEnvoyAdminTunnel() envoy_admin.Tunnel {
	return tunnel.NewUniversalEnvoyAdminTunnel(func(cmdName, cmd string) (string, error) {
		session, err := s.RunOnHost("envoytunnel"+cmdName, cmd)
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

// Stop stops all running components on this host
func (s *VmUniversalApp) Stop() error {
	Logf("Stopping app:%q", s.appName)

	Logf("Uninstalling the transparent proxy for app:%q", s.appName)
	session, _ := s.RunOnHost(fmt.Sprintf("dp-%s-tp-uninstall", s.appName), "sudo /tmp/kuma/bin/kumactl uninstall transparent-proxy")
	_ = session.Wait()

	_ = s.universalNetworking.Close()

	err := s.KillMainApp()

	return err
}

func (s *VmUniversalApp) ReStart() error {
	Logf("Restarting app:%q", s.appName)
	if err := s.KillMainApp(); err != nil {
		return err
	}
	// No needed but this just in case kill -9 is not instant
	time.Sleep(1 * time.Second)
	return s.RunMainApp(s.mainAppCmd)
}

func (s *VmUniversalApp) KillMainApp() error {
	Logf("Stopping main app:%q on remote host %q", s.appName, s.universalNetworking.RemoteHost.Address)

	dockerCmd := fmt.Sprintf("docker stop %s", vmMainAppContainerName)

	session, err := s.RunOnHost(fmt.Sprintf("dp-%s-main-app-stop", s.appName), dockerCmd)
	if err == nil {
		err = session.Run()
	}
	return err
}

func (s *VmUniversalApp) RunMainApp(cmd string) error {
	Logf("Starting main app:%q on remote host %q", s.appName, s.universalNetworking.RemoteHost.Address)

	s.mainAppCmd = cmd

	dockerCmd := fmt.Sprintf("docker run --detach --rm --name %s --network host --privileged --entrypoint /bin/sh %s -c %s",
		vmMainAppContainerName,
		Config.GetUniversalImage(),
		cmd)

	session, err := s.RunOnHost(fmt.Sprintf("dp-%s-main-app-start", s.appName), dockerCmd)
	if err == nil {
		err = session.Run()
	}
	return err
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
DOWNLOAD_DIR=$(find -maxdepth 1 -type d -name '*-%s')
if [ "$DOWNLOAD_DIR" == "" ]; then
  curl -L --no-progress-bar --fail '%s' | VERSION=%s sh
  DOWNLOAD_DIR=$(find -maxdepth 1 -type d -name '*-%s')
fi
if [ "$DOWNLOAD_DIR" == "" ]; then
  >&2 echo "Could not download the installer"
  exit 1;
fi
cp -f $DOWNLOAD_DIR/bin/* %s/bin/
export PATH=$PATH:%s/bin
		`, Config.KumaInstallerUrl, Config.KumaImageTag, Config.KumaImageTag, workingDir, workingDir)

	for k, v := range envsMap {
		_, _ = fmt.Fprintf(cmd, "export %s=%s\n", k, utils.ShellEscape(v))
	}

	// todo: remove sudo prefix to adapt more distros
	_, _ = fmt.Fprintf(cmd, "sudo useradd --system --no-create-home --shell /sbin/nologin kuma-dp\n")

	// Install transparent proxy
	if transparent {
		extraArgs := []string{}
		if builtindns {
			extraArgs = append(extraArgs, "--redirect-dns")
		}
		_, _ = fmt.Fprintf(cmd, "sudo %s/bin/kumactl install transparent-proxy --exclude-inbound-ports %s %s\n",
			workingDir, sshPort, strings.Join(extraArgs, " "))
	}

	// run the DP as user `envoy` so iptables can distinguish its traffic if needed
	args := []string{
		"sudo", "runuser", "-u", "kuma-dp", "--",
		fmt.Sprintf("%s/bin/kuma-dp", workingDir), "run",
		"--cp-address=" + cpAddress,
		fmt.Sprintf("--dataplane-token-file=%s/token-%s", workingDir, name),
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
	s.dpApp, err = s.RunOnHost("dp-"+name, s.dpAppCmd)
	return err
}

func (s *VmUniversalApp) RunOnHost(name string, cmd string) (*kssh.Session, error) {
	return s.universalNetworking.NewSession(path.Join("vm-universal", "exec"), name, s.verbose, cmd)
}
