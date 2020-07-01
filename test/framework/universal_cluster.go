package framework

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"go.uber.org/multierr"

	config_mode "github.com/Kong/kuma/pkg/config/mode"
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
}

func NewUniversalCluster(t *TestingT, name string, verbose bool) *UniversalCluster {
	return &UniversalCluster{
		t:       t,
		name:    name,
		apps:    map[string]*UniversalApp{},
		verbose: verbose,
	}
}

func (c *UniversalCluster) DismissCluster() (errs error) {
	for _, app := range c.apps {
		err := app.Stop()
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return
}

func (c *UniversalCluster) DeployKuma(mode ...string) error {
	if len(mode) == 0 {
		mode = []string{config_mode.Standalone}
	}
	c.controlplane = NewUniversalControlPlane(c.t, mode[0], c.name, c, c.verbose)

	c.apps[AppModeCP] = NewUniversalApp(c.t, AppModeCP, true, []string{}, []string{"kuma-cp", "run"})
	err := c.apps[AppModeCP].mainApp.Start()
	if err != nil {
		return err
	}

	kumacpURL := "http://localhost:5681"
	return c.controlplane.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
}

func (c *UniversalCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *UniversalCluster) VerifyKuma() error {
	return c.controlplane.kumactl.RunKumactl("get", "dataplanes")
}

func (c *UniversalCluster) RestartKuma() error {
	mode := c.controlplane.mode
	_ = c.DeleteKuma()
	return c.DeployKuma(mode)
}

func (c *UniversalCluster) DeleteKuma() error {
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

	app := NewUniversalApp(c.t, AppMode(appname), c.verbose, []string{}, args)
	err := app.mainApp.Start()
	if err != nil {
		return err
	}
	ip := app.ip

	dpyaml := ""
	switch appname {
	case AppModeEchoServer:
		dpyaml = fmt.Sprintf(EchoServerDataplane, "dp-"+appname, ip, "8080", "80")
	case AppModeDemoClient:
		dpyaml = fmt.Sprintf(DemoClientDataplane, "dp-"+appname, ip, "13000", "3000")
	}

	// apply the dataplane
	err = c.controlplane.kumactl.KumactlApplyFromString(fmt.Sprint(dpyaml))
	if err != nil {
		return err
	}

	// generate the token on the CP node
	token, _ := NewSshApp(c.verbose, c.apps[AppModeCP].ip, []string{}, []string{"curl",
		"-H", "\"Content-Type: application/json\"",
		"--data", "'{\"name\": \"dp-" + appname + "\", \"mesh\": \"default\"}'",
		"http://localhost:5679/tokens"}).cmd.Output()

	app.CreateDP(string(token), c.apps[AppModeCP].ip, appname)
	err = app.dpApp.Start()
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

func (c *UniversalCluster) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("not implementedv")
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
				return "", errors.Errorf("App %s not found for deletion", appname)
			}
			sshApp := NewSshApp(false, app.sshPort, []string{}, cmd)
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
