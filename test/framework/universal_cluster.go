package framework

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/go-errors/errors"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"go.uber.org/multierr"
)

const (
	kumaUniversalImage = "kuma-universal"
)

type UniversalCluster struct {
	t            testing.TestingT
	name         string
	controlplane *UniversalControlPlane
	apps         map[string]*UniversalApp
	verbose      bool
	deployments  map[string]Deployment
}

func NewUniversalCluster(t *TestingT, name string, verbose bool) *UniversalCluster {
	return &UniversalCluster{
		t:           t,
		name:        name,
		apps:        map[string]*UniversalApp{},
		verbose:     verbose,
		deployments: map[string]Deployment{},
	}
}

func (c *UniversalCluster) DismissCluster() (errs error) {
	for _, app := range c.apps {
		err := app.Stop()
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	for _, deployment := range c.deployments {
		if err := deployment.Delete(c); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return
}

func (c *UniversalCluster) DeployKuma(mode string, fs ...DeployOptionsFunc) error {
	c.controlplane = NewUniversalControlPlane(c.t, mode, c.name, c, c.verbose)
	opts := newDeployOpt(fs...)

	if opts.installationMode != KumactlInstallationMode {
		return errors.Errorf("universal clusters only support the '%s' installation mode but got '%s'", KumactlInstallationMode, opts.installationMode)
	}

	cmd := []string{"kuma-cp", "run"}
	env := []string{"KUMA_MODE=" + mode}
	if opts.globalAddress != "" {
		env = append(env, "KUMA_MULTICLUSTER_REMOTE_GLOBAL_ADDRESS="+opts.globalAddress)
	}

	switch mode {
	case core.Remote:
		env = append(env, "KUMA_MULTICLUSTER_REMOTE_ZONE="+c.name)
	case core.Global:
		cmd = append(cmd, "--config-file", confPath)
	}

	app, err := NewUniversalApp(c.t, c.name, AppModeCP, true, env, cmd)
	if err != nil {
		return err
	}

	err = app.mainApp.Start()
	if err != nil {
		return err
	}

	kumacpURL := "http://localhost:" + app.ports["5681"]
	err = c.controlplane.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
	if err != nil {
		return err
	}

	c.apps[AppModeCP] = app

	switch mode {
	case core.Remote:
		dpyaml := fmt.Sprintf(IngressDataplane, app.ip, kdsPort)
		err = c.CreateDP(app, "ingress", dpyaml)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *UniversalCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *UniversalCluster) VerifyKuma() error {
	return c.controlplane.kumactl.RunKumactl("get", "dataplanes")
}

func (c *UniversalCluster) RestartKuma() error {
	return c.apps[AppModeCP].ReStart()
}

func (c *UniversalCluster) DeleteKuma(opts ...DeployOptionsFunc) error {
	err := c.apps[AppModeCP].Stop()
	delete(c.apps, AppModeCP)
	c.controlplane = nil
	return err
}

func (c *UniversalCluster) InjectDNS() error {
	return nil
}

func (c *UniversalCluster) GetKumactlOptions() *KumactlOptions {
	return c.controlplane.kumactl
}

// K8s
func (c *UniversalCluster) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	return nil
}

func (c *UniversalCluster) CreateNamespace(namespace string) error {
	return nil
}

func (c *UniversalCluster) DeleteNamespace(namespace string) error {
	return nil
}

func (c *UniversalCluster) CreateDP(app *UniversalApp, appname, dpyaml string) error {
	// generate the token on the CP node
	sshApp := NewSshApp(c.verbose, c.apps[AppModeCP].ports["22"], []string{}, []string{"curl",
		"-H", "\"Content-Type: application/json\"",
		"--data", "'{\"name\": \"dp-" + appname + "\", \"mesh\": \"default\"}'",
		"http://localhost:5679/tokens"})
	if err := sshApp.Run(); err != nil {
		return err
	}

	token := sshApp.Out()

	cpAddress := "http://" + c.apps[AppModeCP].ip + ":5681"
	app.CreateDP(token, cpAddress, appname, dpyaml)

	return app.dpApp.Start()
}

func (c *UniversalCluster) DeployApp(namespace, appname string) error {
	var args []string
	switch appname {
	case AppModeEchoServer:
		args = []string{"nc", "-lk", "-p", "80", "-e", "echo", "-e", "\"HTTP/1.1 200 OK\n\n Echo\n\""}
	case AppModeDemoClient:
		args = []string{"nc", "-lvk", "-p", "3000"}
	default:
		return errors.Errorf("not supported app type %s", appname)
	}

	app, err := NewUniversalApp(c.t, c.name, AppMode(appname), c.verbose, []string{}, args)
	if err != nil {
		return err
	}

	err = app.mainApp.Start()
	if err != nil {
		return err
	}
	ip := app.ip

	dpyaml := ""
	switch appname {
	case AppModeEchoServer:
		dpyaml = fmt.Sprintf(EchoServerDataplane, "dp-"+appname, ip, "8080", "80", "8080")
	case AppModeDemoClient:
		dpyaml = fmt.Sprintf(DemoClientDataplane, "dp-"+appname, ip, "13000", "3000", "80", "8080")
	}

	err = c.CreateDP(app, appname, dpyaml)
	if err != nil {
		return err
	}

	c.apps[appname] = app

	return nil
}

func (c *UniversalCluster) DeleteApp(namespace, appname string) error {
	app, ok := c.apps[appname]
	if !ok {
		return errors.Errorf("App %s not found for deletion", appname)
	}
	return app.Stop()
}

func (c *UniversalCluster) Exec(namespace, podName, appname string, cmd ...string) (string, string, error) {
	app, ok := c.apps[appname]
	if !ok {
		return "", "", errors.Errorf("App %s not found", appname)
	}
	sshApp := NewSshApp(false, app.ports[sshPort], []string{}, cmd)
	return sshApp.Out(), sshApp.Err(), sshApp.Run()
}

func (c *UniversalCluster) ExecWithRetries(namespace, podName, appname string, cmd ...string) (string, string, error) {
	var stdout string
	var stderr string
	_, err := retry.DoWithRetryE(
		c.t,
		fmt.Sprintf("Trying %s", strings.Join(cmd, " ")),
		DefaultRetries/3,
		DefaultTimeout,
		func() (string, error) {
			app, ok := c.apps[appname]
			if !ok {
				return "", errors.Errorf("App %s not found", appname)
			}
			sshApp := NewSshApp(false, app.ports[sshPort], []string{}, cmd)
			err := sshApp.Run()
			stdout = sshApp.Out()
			stderr = sshApp.Err()

			return stdout, err
		},
	)

	return stdout, stderr, err
}

func (c *UniversalCluster) GetTesting() testing.TestingT {
	return c.t
}

func (c *UniversalCluster) Deployment(name string) Deployment {
	return c.deployments[name]
}

func (c *UniversalCluster) Deploy(deployment Deployment) error {
	c.deployments[deployment.Name()] = deployment
	return deployment.Deploy(c)
}
