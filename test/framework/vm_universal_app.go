package framework

import (
	"fmt"
	"golang.org/x/crypto/ssh"
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
	t                     testing.TestingT
	mainAppCmd            string
	mainAppContainerImage string
	dpAppCmd              string
	longRunningSessions   map[string]*kssh.Session

	verbose bool
	mesh    string

	universalNetworking *universal.Networking
	appName             string

	mainAppLogger *logger.Logger
}

func NewVmUniversalApp(t testing.TestingT, appName, mesh string, host *VmClusterHost, verbose bool) (*VmUniversalApp, error) {
	app := &VmUniversalApp{
		t:                   t,
		verbose:             verbose,
		mesh:                mesh,
		appName:             appName,
		longRunningSessions: map[string]*kssh.Session{},
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
		session, err := s.CreateHostProcess("envoytunnel"+cmdName, cmd, false)
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

	if s.mainAppCmd != "" {
		_ = s.KillMainApp()
	}

	for _, session := range s.longRunningSessions {
		_ = session.Signal(ssh.SIGKILL, false)
		_ = session.Close()
	}

	if s.dpAppCmd != "" {
		Logf("Uninstalling the transparent proxy for app:%q", s.appName)
		session, _ := s.CreateHostProcess(fmt.Sprintf("dp-%s-tp-uninstall", s.appName), "sudo /tmp/kuma/bin/kumactl uninstall transparent-proxy", false)
		_ = session.Start()
		_ = session.Wait()
	}

	_ = s.universalNetworking.Close()

	return nil
}

func (s *VmUniversalApp) ReStart() error {
	Logf("Restarting app:%q", s.appName)
	if err := s.KillMainApp(); err != nil {
		return err
	}
	// No needed but this just in case kill -9 is not instant
	time.Sleep(1 * time.Second)
	return s.RunMainApp(s.mainAppCmd, s.mainAppContainerImage)
}

func (s *VmUniversalApp) KillMainApp() error {
	Logf("Stopping main app:%q on remote host %q", s.appName, s.universalNetworking.RemoteHost.Address)

	dockerCmd := fmt.Sprintf("docker stop %s", vmMainAppContainerName)

	session, err := s.CreateHostProcess(fmt.Sprintf("dp-%s-main-app-stop", s.appName), dockerCmd, false)
	if err == nil {
		err = session.Run()
		_ = session.Close()
	}
	return err
}

func (s *VmUniversalApp) RunMainApp(cmd, containerImage string) error {
	Logf("Starting main app:%q on remote host %q", s.appName, s.universalNetworking.RemoteHost.Address)

	s.mainAppCmd = cmd

	dockerCmd := fmt.Sprintf("docker run --name %s --rm --network host --privileged --entrypoint /bin/sh %s -c '%s'",
		vmMainAppContainerName,
		containerImage,
		cmd)

	session, err := s.CreateHostProcess(fmt.Sprintf("dp-%s-main-app-start", s.appName), dockerCmd, true)
	if err == nil {
		err = session.Start()
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
) (*kssh.Session, error) {
	cmd := &strings.Builder{}

	workingDir := fmt.Sprintf("/tmp/kuma")
	// create the token file on the app container
	_, _ = cmd.WriteString("#!/bin/bash\n")
	_, _ = fmt.Fprintf(cmd, "mkdir -p %s/bin\n", workingDir)
	_, _ = fmt.Fprintf(cmd, "printf %q > %s/token-%s\n", token, workingDir, name)

	_, _ = fmt.Fprintf(cmd, `
DOWNLOAD_DIR=$(find -maxdepth 1 -type d -name '*-%s')
if [[ "$DOWNLOAD_DIR" == "" ]]; then
  curl -L --no-progress-bar --fail '%s' | VERSION=%s sh
  DOWNLOAD_DIR=$(find -maxdepth 1 -type d -name '*-%s')
fi
if [[ "$DOWNLOAD_DIR" == "" ]]; then
  >&2 echo "Could not download the installer"
  exit 1;
fi
cp -f $DOWNLOAD_DIR/bin/* %s/bin/
export PATH=$PATH:%s/bin
		`, Config.KumaImageTag, Config.KumaInstallerUrl, Config.KumaImageTag, Config.KumaImageTag, workingDir, workingDir)

	for k, v := range envsMap {
		_, _ = fmt.Fprintf(cmd, "export %s=%s\n", k, utils.ShellEscape(v))
	}

	// todo: remove sudo prefix to adapt more distros
	_, _ = fmt.Fprintf(cmd, "sudo useradd --system --no-create-home --shell /sbin/nologin kuma-dp\n")

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
	dpSession, err := s.CreateHostProcess("dp-"+name, s.dpAppCmd, true)

	// Install transparent proxy after the kuma-dp to make it fast, otherwise the runuser will trigger a DNS resolutin timeout
	if err == nil && transparent {
		var extraArgs []string
		if builtindns {
			extraArgs = append(extraArgs, "--redirect-dns")
		}
		tpSession, _ := s.CreateHostProcess(fmt.Sprintf("dp-%s-tp-install", s.appName), fmt.Sprintf("sudo %s/bin/kumactl install transparent-proxy --exclude-inbound-ports %s %s\n",
			workingDir, sshPort, strings.Join(extraArgs, " ")), false)
		err = tpSession.Start()
		if err == nil {
			_ = tpSession.Wait()
		}
	}

	if err != nil {
		if dpSession != nil {
			_ = dpSession.Close()
		}
		return nil, err
	}
	return dpSession, nil
}

func (s *VmUniversalApp) CreateHostProcess(name string, cmd string, longRunning bool) (*kssh.Session, error) {
	session, err := s.universalNetworking.NewSession(path.Join("vm-universal", "exec"), name, s.verbose, cmd)
	if err != nil {
		return nil, err
	}

	if longRunning {
		s.longRunningSessions[name] = session
	}
	return session, nil
}
