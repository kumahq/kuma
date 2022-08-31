package framework

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/template"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/test/framework/ssh"
)

type UniversalNetworking struct {
	IP            string `json:"ip"` // IP inside a docker network
	ApiServerPort string `json:"apiServerPort"`
	SshPort       string `json:"sshPort"`
}

func (u UniversalNetworking) BootstrapAddress() string {
	return "https://" + net.JoinHostPort(u.IP, "5678")
}

type UniversalNetworkingState struct {
	ZoneEgress  UniversalNetworking `json:"zoneEgress"`
	ZoneIngress UniversalNetworking `json:"zoneIngress"`
	KumaCp      UniversalNetworking `json:"kumaCp"`
}

type UniversalCluster struct {
	t              testing.TestingT
	name           string
	controlplane   *UniversalControlPlane
	apps           map[string]*UniversalApp
	verbose        bool
	deployments    map[string]Deployment
	defaultTimeout time.Duration
	defaultRetries int
	opts           kumaDeploymentOptions

	envoyTunnels map[string]envoy_admin.Tunnel
	networking   map[string]UniversalNetworking
}

var _ Cluster = &UniversalCluster{}

func NewUniversalCluster(t *TestingT, name string, verbose bool) *UniversalCluster {
	return &UniversalCluster{
		t:              t,
		name:           name,
		apps:           map[string]*UniversalApp{},
		verbose:        verbose,
		deployments:    map[string]Deployment{},
		defaultRetries: Config.DefaultClusterStartupRetries,
		defaultTimeout: Config.DefaultClusterStartupTimeout,
		envoyTunnels:   map[string]envoy_admin.Tunnel{},
		networking:     map[string]UniversalNetworking{},
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
	for name, deployment := range c.deployments {
		if err := deployment.Delete(c); err != nil {
			errs = multierr.Append(errs, err)
		}
		delete(c.deployments, name)
	}
	return
}

func (c *UniversalCluster) Verbose() bool {
	return c.verbose
}

func (c *UniversalCluster) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
	if mode == core.Zone {
		opt = append([]KumaDeploymentOption{WithEnvs(Config.KumaZoneUniversalEnvVars)}, opt...)
	} else {
		opt = append([]KumaDeploymentOption{WithEnvs(Config.KumaUniversalEnvVars)}, opt...)
	}
	c.opts.apply(opt...)
	if c.opts.installationMode != KumactlInstallationMode {
		return errors.Errorf("universal clusters only support the '%s' installation mode but got '%s'", KumactlInstallationMode, c.opts.installationMode)
	}

	env := map[string]string{"KUMA_MODE": mode, "KUMA_DNS_SERVER_PORT": "53"}

	for k, v := range c.opts.env {
		env[k] = v
	}
	if c.opts.globalAddress != "" {
		env["KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS"] = c.opts.globalAddress
	}
	if c.opts.hdsDisabled {
		env["KUMA_DP_SERVER_HDS_ENABLED"] = "false"
	}

	if Config.XDSApiVersion != "" {
		env["KUMA_BOOTSTRAP_SERVER_API_VERSION"] = Config.XDSApiVersion
	}

	if Config.CIDR != "" {
		env["KUMA_DNS_SERVER_CIDR"] = Config.CIDR
	}

	cmd := []string{"kuma-cp", "run"}
	switch mode {
	case core.Zone:
		env["KUMA_MULTIZONE_ZONE_NAME"] = c.name
	case core.Global:
		cmd = append(cmd, "--config-file", "/kuma/kuma-cp.conf")
	}

	app, err := NewUniversalApp(c.t, c.name, AppModeCP, "", AppModeCP, c.opts.isipv6, true, []string{}, nil, "")
	if err != nil {
		return err
	}

	app.CreateMainApp(env, cmd)

	if c.opts.runPostgresMigration {
		if err := runPostgresMigration(app, env); err != nil {
			return err
		}
	}

	if err := app.mainApp.Start(); err != nil {
		return err
	}

	c.apps[AppModeCP] = app

	pf := UniversalNetworking{
		IP:            app.ip,
		ApiServerPort: app.ports["5681"],
		SshPort:       app.ports["22"],
	}
	c.controlplane, err = NewUniversalControlPlane(c.t, mode, c.name, c.verbose, pf)
	if err != nil {
		return err
	}

	for name, updateFuncs := range c.opts.meshUpdateFuncs {
		for _, f := range updateFuncs {
			Logf("applying update function to mesh %q", name)
			err := c.controlplane.kumactl.KumactlUpdateObject("mesh", name,
				func(resource core_model.Resource) core_model.Resource {
					mesh := resource.(*core_mesh.MeshResource)
					mesh.Spec = f(mesh.Spec)
					return mesh
				})
			if err != nil {
				return err
			}
		}
	}

	return c.VerifyKuma()
}

func (c *UniversalCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *UniversalCluster) GetKumaCPLogs() (string, error) {
	return c.apps[AppModeCP].mainApp.Out(), nil
}

func (c *UniversalCluster) VerifyKuma() error {
	return c.controlplane.kumactl.RunKumactl("get", "dataplanes")
}

func (c *UniversalCluster) DeleteKuma() error {
	err := c.apps[AppModeCP].Stop()
	delete(c.apps, AppModeCP)
	c.controlplane = nil
	return err
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

func (c *UniversalCluster) CreateDP(app *UniversalApp, name, mesh, ip, dpyaml, token string, builtindns bool, concurrency int) error {
	cpIp := c.controlplane.Networking().IP
	cpAddress := "https://" + net.JoinHostPort(cpIp, "5678")
	app.CreateDP(token, cpAddress, name, mesh, ip, dpyaml, builtindns, "", concurrency)
	return app.dpApp.Start()
}

func (c *UniversalCluster) CreateZoneIngress(app *UniversalApp, name, ip, dpyaml, token string, builtindns bool) error {
	app.CreateDP(token, c.controlplane.Networking().BootstrapAddress(), name, "", ip, dpyaml, builtindns, "ingress", 0)

	if err := c.addIngressEnvoyTunnel(); err != nil {
		return err
	}

	return app.dpApp.Start()
}

func (c *UniversalCluster) CreateZoneEgress(
	app *UniversalApp,
	name, ip, dpYAML, token string,
	builtinDNS bool,
) error {
	app.CreateDP(token, c.controlplane.Networking().BootstrapAddress(), name, "", ip, dpYAML, builtinDNS, "egress", 0)

	if err := c.addEgressEnvoyTunnel(); err != nil {
		return err
	}

	return app.dpApp.Start()
}

func (c *UniversalCluster) DeployApp(opt ...AppDeploymentOption) error {
	var opts appDeploymentOptions
	opts.apply(opt...)
	appname := opts.appname
	token := opts.token
	transparent := opts.transparent != nil && *opts.transparent // default false
	args := opts.appArgs

	if opts.verbose == nil {
		opts.verbose = &c.verbose
	}

	caps := []string{}
	if transparent {
		caps = append(caps, "NET_ADMIN", "NET_RAW")
	}

	Logf("IPV6 is %v", opts.isipv6)

	app, err := NewUniversalApp(c.t, c.name, opts.name, opts.mesh, AppMode(appname), opts.isipv6, *opts.verbose, caps, opts.dockerVolumes, opts.dockerContainerName)
	if err != nil {
		return err
	}

	// We need to record the app before running any other options,
	// since those options might fail. If they do, we have a running
	// container that isn't fully configured, and we need it to be
	// recorded so that DismissCluster can clean it up.
	Logf("Started universal app %q in container %q", opts.name, app.container)
	c.apps[opts.name] = app

	if !opts.omitDataplane {
		if opts.kumactlFlow {
			dataplaneResource := template.Render(opts.appYaml, map[string]string{
				"name":    opts.name,
				"address": app.ip,
			})
			err := c.GetKumactlOptions().KumactlApplyFromString(string(dataplaneResource))
			if err != nil {
				return err
			}
		}

		if opts.dpVersion != "" {
			// override needs to be before setting up transparent proxy.
			// Otherwise, we won't be able to fetch specific Kuma DP version.
			if err := app.OverrideDpVersion(opts.dpVersion); err != nil {
				return err
			}
		}

		builtindns := opts.builtindns == nil || *opts.builtindns
		if transparent {
			app.setupTransparent(c.controlplane.Networking().IP, builtindns)
		}

		ip := app.ip

		var dataplaneResource string
		if opts.kumactlFlow {
			dataplaneResource = ""
		} else {
			dataplaneResource = opts.appYaml
		}

		if opts.mesh == "" {
			opts.mesh = "default"
		}
		if err := c.CreateDP(app, opts.name, opts.mesh, ip, dataplaneResource, token, builtindns, opts.concurrency); err != nil {
			return err
		}
	}

	if opts.boundToContainerIp {
		args = append(args, "--ip", app.ip)
	}

	if !opts.proxyOnly {
		app.CreateMainApp(nil, args)
		err = app.mainApp.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func runPostgresMigration(kumaCP *UniversalApp, envVars map[string]string) error {
	args := []string{
		"/usr/bin/kuma-cp", "migrate", "up",
	}

	sshPort := kumaCP.GetPublicPort("22")
	if sshPort == "" {
		return errors.New("missing public port: 22")
	}

	app := ssh.NewApp(kumaCP.containerName, kumaCP.verbose, sshPort, envVars, args)
	if err := app.Run(); err != nil {
		return errors.Errorf("db migration err: %s\nstderr :%s\nstdout %s", err.Error(), app.Err(), app.Out())
	}

	return nil
}

func (c *UniversalCluster) GetApp(appName string) *UniversalApp {
	return c.apps[appName]
}

func (c *UniversalCluster) DeleteApp(appname string) error {
	app, ok := c.apps[appname]
	if !ok {
		return errors.Errorf("App %s not found for deletion", appname)
	}
	if err := app.Stop(); err != nil {
		return err
	}
	delete(c.apps, appname)
	return nil
}

func (c *UniversalCluster) DeleteMesh(mesh string) error {
	return c.GetKumactlOptions().KumactlDelete("mesh", mesh, "")
}

func (c *UniversalCluster) DeleteMeshApps(mesh string) error {
	for name := range c.apps {
		if c.GetApp(name).mesh == mesh {
			if err := c.DeleteApp(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *UniversalCluster) Exec(namespace, podName, appname string, cmd ...string) (string, string, error) {
	app, ok := c.apps[appname]
	if !ok {
		return "", "", errors.Errorf("App %s not found", appname)
	}
	sshApp := ssh.NewApp(app.containerName, c.verbose, app.ports[sshPort], nil, cmd)
	err := sshApp.Run()
	return sshApp.Out(), sshApp.Err(), err
}

func (c *UniversalCluster) ExecWithRetries(namespace, podName, appname string, cmd ...string) (string, string, error) {
	return c.ExecWithCustomRetries(namespace, podName, appname, c.defaultRetries, c.defaultTimeout, cmd...)
}

func (c *UniversalCluster) ExecWithCustomRetries(namespace, podName, appname string, retries int, timeout time.Duration, cmd ...string) (string, string, error) {
	var stdout string
	var stderr string
	_, err := retry.DoWithRetryE(
		c.t,
		fmt.Sprintf("Trying %s", strings.Join(cmd, " ")),
		retries,
		timeout,
		func() (string, error) {
			app, ok := c.apps[appname]
			if !ok {
				return "", errors.Errorf("App %s not found", appname)
			}
			sshApp := ssh.NewApp(app.containerName, c.verbose, app.ports[sshPort], nil, cmd)
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

func (c *UniversalCluster) DeleteDeployment(name string) error {
	deployment, ok := c.deployments[name]
	if !ok {
		return errors.Errorf("deployment %s not found", name)
	}
	if err := deployment.Delete(c); err != nil {
		return err
	}
	delete(c.deployments, name)
	return nil
}

func (c *UniversalCluster) GetZoneIngressNetworking() UniversalNetworking {
	return c.networking[Config.ZoneIngressApp]
}

func (c *UniversalCluster) GetZoneEgressNetworking() UniversalNetworking {
	return c.networking[Config.ZoneEgressApp]
}

func (c *UniversalCluster) AddNetworking(networking UniversalNetworking, name string) error {
	c.networking[name] = networking
	err := c.createEnvoyTunnel(name)
	if err != nil {
		return err
	}
	return nil
}

func (c *UniversalCluster) addEgressEnvoyTunnel() error {
	app := c.apps[AppEgress]
	c.networking[Config.ZoneEgressApp] = UniversalNetworking{
		IP:            "localhost",
		ApiServerPort: app.GetPublicPort(sshPort),
		SshPort:       sshPort,
	}

	err := c.createEnvoyTunnel(Config.ZoneEgressApp)
	if err != nil {
		return err
	}
	return nil
}

func (c *UniversalCluster) addIngressEnvoyTunnel() error {
	app := c.apps[AppIngress]
	c.networking[Config.ZoneIngressApp] = UniversalNetworking{
		IP:            "localhost",
		ApiServerPort: app.GetPublicPort(sshPort),
		SshPort:       sshPort,
	}

	err := c.createEnvoyTunnel(Config.ZoneIngressApp)
	if err != nil {
		return err
	}
	return nil
}

func (c *UniversalCluster) createEnvoyTunnel(name string) error {
	t, err := tunnel.NewUniversalEnvoyAdminTunnel(c.t, c.networking[name].ApiServerPort, c.verbose)
	if err != nil {
		return err
	}
	c.envoyTunnels[name] = t
	return nil
}

func (c *UniversalCluster) GetZoneEgressEnvoyTunnel() envoy_admin.Tunnel {
	t, err := c.GetZoneEgressEnvoyTunnelE()
	if err != nil {
		c.t.Fatal(err)
	}

	return t
}

func (c *UniversalCluster) GetZoneIngressEnvoyTunnel() envoy_admin.Tunnel {
	t, err := c.GetZoneIngressEnvoyTunnelE()
	if err != nil {
		c.t.Fatal(err)
	}

	return t
}

func (c *UniversalCluster) GetZoneEgressEnvoyTunnelE() (envoy_admin.Tunnel, error) {
	t, ok := c.envoyTunnels[Config.ZoneEgressApp]
	if !ok {
		return nil, errors.Errorf("no tunnel with name %+q", Config.ZoneEgressApp)
	}

	return t, nil
}

func (c *UniversalCluster) GetZoneIngressEnvoyTunnelE() (envoy_admin.Tunnel, error) {
	t, ok := c.envoyTunnels[Config.ZoneIngressApp]
	if !ok {
		return nil, errors.Errorf("no tunnel with name %+q", Config.ZoneIngressApp)
	}

	return t, nil
}

func (c *UniversalCluster) Install(fn InstallFunc) error {
	return fn(c)
}

func (c *UniversalCluster) SetCp(cp *UniversalControlPlane) {
	c.controlplane = cp
}
