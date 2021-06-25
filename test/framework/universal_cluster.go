package framework

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/go-errors/errors"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"go.uber.org/multierr"
)

type UniversalCluster struct {
	t              testing.TestingT
	name           string
	controlplane   *UniversalControlPlane
	apps           map[string]*UniversalApp
	verbose        bool
	deployments    map[string]Deployment
	defaultTimeout time.Duration
	defaultRetries int
}

func NewUniversalCluster(t *TestingT, name string, verbose bool) *UniversalCluster {
	return &UniversalCluster{
		t:              t,
		name:           name,
		apps:           map[string]*UniversalApp{},
		verbose:        verbose,
		deployments:    map[string]Deployment{},
		defaultRetries: GetDefaultRetries(),
		defaultTimeout: GetDefaultTimeout(),
	}
}

func (c *UniversalCluster) WithTimeout(timeout time.Duration) Cluster {
	c.defaultTimeout = timeout

	return c
}

func (c *UniversalCluster) WithRetries(retries int) Cluster {
	c.defaultRetries = retries

	return c
}

func (c *UniversalCluster) Name() string {
	return c.name
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
	env := []string{"KUMA_MODE=" + mode, "KUMA_DNS_SERVER_PORT=53"}

	caps := []string{}
	for k, v := range opts.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	if opts.globalAddress != "" {
		env = append(env, "KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS="+opts.globalAddress)
	}
	if opts.hdsDisabled {
		env = append(env, "KUMA_DP_SERVER_HDS_ENABLED=false")
	}

	if HasApiVersion() {
		env = append(env, "KUMA_BOOTSTRAP_SERVER_API_VERSION="+GetApiVersion())
	}

	if opts.isipv6 {
		env = append(env, fmt.Sprintf("KUMA_DNS_SERVER_CIDR=\"%s\"", cidrIPv6))
	}

	switch mode {
	case core.Zone:
		env = append(env, "KUMA_MULTIZONE_ZONE_NAME="+c.name)
	case core.Global:
		cmd = append(cmd, "--config-file", confPath)
	}

	app, err := NewUniversalApp(c.t, c.name, AppModeCP, AppModeCP, opts.isipv6, true, caps)
	if err != nil {
		return err
	}

	app.CreateMainApp(env, cmd)

	if opts.runPostgresMigration {
		if err := runPostgresMigration(app, env); err != nil {
			return err
		}
	}

	if err := app.mainApp.Start(); err != nil {
		return err
	}

	kumacpURL := "http://localhost:" + app.ports["5681"]
	err = c.controlplane.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
	if err != nil {
		return err
	}

	c.apps[AppModeCP] = app

	return nil
}

func (c *UniversalCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *UniversalCluster) VerifyKuma() error {
	return c.controlplane.kumactl.RunKumactl("get", "dataplanes")
}

func (c *UniversalCluster) DeleteKuma(opts ...DeployOptionsFunc) error {
	err := c.apps[AppModeCP].Stop()
	delete(c.apps, AppModeCP)
	c.controlplane = nil
	return err
}

func (c *UniversalCluster) InjectDNS(namespace ...string) error {
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

func (c *UniversalCluster) CreateDP(app *UniversalApp, appname, ip, dpyaml, token string, builtindns bool) error {
	cpIp := c.apps[AppModeCP].ip
	cpAddress := "https://" + net.JoinHostPort(cpIp, "5678")
	app.CreateDP(token, cpAddress, appname, ip, dpyaml, builtindns, false)
	return app.dpApp.Start()
}

func (c *UniversalCluster) CreateZoneIngress(app *UniversalApp, appname, ip, dpyaml, token string, builtindns bool) error {
	cpIp := c.apps[AppModeCP].ip
	cpAddress := "https://" + net.JoinHostPort(cpIp, "5678")
	app.CreateDP(token, cpAddress, appname, ip, dpyaml, builtindns, true)
	return app.dpApp.Start()
}

func (c *UniversalCluster) DeployApp(fs ...DeployOptionsFunc) error {
	opts := newDeployOpt(fs...)
	appname := opts.appname
	token := opts.token
	transparent := opts.transparent
	dpyaml := opts.appYaml
	args := opts.appArgs

	if opts.mesh == "" {
		opts.mesh = "default"
	}

	caps := []string{}
	if transparent {
		caps = append(caps, "NET_ADMIN", "NET_RAW")
	}

	fmt.Printf("IPV6 is %v\n", opts.isipv6)
	app, err := NewUniversalApp(c.t, c.name, opts.name, AppMode(appname), opts.isipv6, c.verbose, caps)
	if err != nil {
		return err
	}

	builtindns := opts.builtindns == nil || *opts.builtindns
	if transparent {
		app.setupTransparent(c.apps[AppModeCP].ip, builtindns)
	}

	ip := app.ip

	if opts.dpVersion != "" {
		if err := app.OverrideDpVersion(opts.dpVersion); err != nil {
			return err
		}
	}

	err = c.CreateDP(app, opts.name, ip, dpyaml, token, builtindns)
	if err != nil {
		return err
	}

	if !opts.proxyOnly {
		app.CreateMainApp([]string{}, args)
		err = app.mainApp.Start()
		if err != nil {
			return err
		}
	}

	c.apps[opts.name] = app

	return nil
}

func runPostgresMigration(kumaCP *UniversalApp, envVars []string) error {
	args := []string{
		"/usr/bin/kuma-cp", "migrate", "up",
	}

	sshPort := kumaCP.GetPublicPort("22")
	if sshPort == "" {
		return errors.New("missing public port: 22")
	}

	app := NewSshApp(true, sshPort, envVars, args)
	if err := app.Run(); err != nil {
		return errors.Errorf("db migration err: %s\nstderr :%s\nstdout %s", err.Error(), app.Err(), app.Out())
	}

	return nil
}

func (c *UniversalCluster) GetApp(appName string) *UniversalApp {
	return c.apps[appName]
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
	err := sshApp.Run()
	return sshApp.Out(), sshApp.Err(), err
}

func (c *UniversalCluster) ExecWithRetries(namespace, podName, appname string, cmd ...string) (string, string, error) {
	var stdout string
	var stderr string
	_, err := retry.DoWithRetryE(
		c.t,
		fmt.Sprintf("Trying %s", strings.Join(cmd, " ")),
		c.defaultRetries/3,
		c.defaultTimeout,
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
