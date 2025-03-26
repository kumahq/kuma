package framework

import (
	"fmt"
	"net"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/crypto/ssh"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/template"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/test/framework/kumactl"
	kssh "github.com/kumahq/kuma/test/framework/ssh"
	"github.com/kumahq/kuma/test/framework/universal"
	"github.com/kumahq/kuma/test/framework/utils"
)

type VmClusterHost struct {
	kssh.Host
	InternalIPAddress string
	AppMode           AppMode
}

type VmPortForward struct {
	StopChannel   chan struct{}
	LocalPort     int
	RemoteAddress string
}

type VmUniversalCluster struct {
	t            testing.TestingT
	name         string
	controlplane *VmUniversalControlPlane

	hosts          map[string]*VmClusterHost
	apps           map[string]*VmUniversalApp
	verbose        bool
	deployments    map[string]Deployment
	dataplanes     []string
	portForwards   map[string]*VmPortForward
	defaultTimeout time.Duration
	defaultRetries int
	opts           KumaDeploymentOptions
	mutex          sync.RWMutex

	envoyTunnels map[string]envoy_admin.Tunnel
	networking   map[string]*universal.Networking
}

var _ Cluster = &VmUniversalCluster{}

func NewVmUniversalCluster(t *TestingT, name string, hosts []*VmClusterHost, verbose bool) *VmUniversalCluster {
	hostMap := make(map[string]*VmClusterHost)
	for _, host := range hosts {
		hostMap[string(host.AppMode)] = host
	}

	return &VmUniversalCluster{
		t:              t,
		name:           name,
		apps:           map[string]*VmUniversalApp{},
		hosts:          hostMap,
		verbose:        verbose,
		deployments:    map[string]Deployment{},
		portForwards:   map[string]*VmPortForward{},
		defaultRetries: Config.DefaultClusterStartupRetries,
		defaultTimeout: Config.DefaultClusterStartupTimeout,
		envoyTunnels:   map[string]envoy_admin.Tunnel{},
		networking:     map[string]*universal.Networking{},
	}
}

func (c *VmUniversalCluster) WithTimeout(timeout time.Duration) Cluster {
	c.defaultTimeout = timeout

	return c
}

func (c *VmUniversalCluster) WithRetries(retries int) Cluster {
	c.defaultRetries = retries

	return c
}

func (c *VmUniversalCluster) Name() string {
	return c.name
}

func (c *VmUniversalCluster) DismissCluster() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var errs error
	for _, app := range c.apps {
		c.closePortForward(app.appName)
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
	return errs
}

func (c *VmUniversalCluster) Verbose() bool {
	return c.verbose
}

func (c *VmUniversalCluster) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
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

	if Config.IPV6 {
		env["KUMA_DNS_SERVER_CIDR"] = "fd00:fd00::/64"
		env["KUMA_IPAM_MESH_SERVICE_CIDR"] = "fd00:fd01::/64"
		env["KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR"] = "fd00:fd02::/64"
		env["KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR"] = "fd00:fd03::/64"
	}

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
	if mode == core.Zone {
		env["KUMA_MULTIZONE_ZONE_NAME"] = c.ZoneName()
		env["KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY"] = "true"
	}

	cmd := &strings.Builder{}
	_, _ = cmd.WriteString("#!/bin/sh\n")
	_, _ = fmt.Fprintf(cmd, `
curl -L --no-progress-bar --fail '%s' | VERSION=%s sh
PRODUCT_DIR=$(find -maxdepth 1 -type d -name '*-%s')
export PATH=$PATH:$(realpath ${PRODUCT_DIR}/bin/)
		`, Config.KumaInstallerUrl, Config.KumaImageTag, Config.KumaImageTag)

	for k, v := range env {
		_, _ = fmt.Fprintf(cmd, "export %s=%s\n", k, utils.ShellEscape(v))
	}
	if c.opts.runPostgresMigration {
		_, _ = fmt.Fprintf(cmd, "kuma-cp migrate up\n")
	}
	// _, _ = fmt.Fprintf(cmd, "cat /kuma/kuma-cp.conf\n")

	runCp := "kuma-cp run" // --config-file /kuma/kuma-cp.conf
	if Config.Debug {
		runCp += " --log-level debug"
	}
	_, _ = fmt.Fprintf(cmd, "%s\n", runCp)

	app, err := NewVmUniversalApp(c.t, AppModeCP, "", c.hosts[AppModeCP], false)
	if err != nil {
		return err
	}
	c.apps[AppModeCP] = app

	localApiServerPort, err := c.PortForward(AppModeCP, "api-server", "localhost:5681")
	if err != nil {
		return fmt.Errorf("could not establish port-forward for the control plane API Server: %w", err)
	}
	app.universalNetworking.ApiServerPort = strconv.Itoa(localApiServerPort)

	app.CreateMainApp(cmd.String())

	if err := app.mainApp.Start(); err != nil {
		return err
	}

	c.controlplane, err = NewVmUniversalControlPlane(c.t, mode, c.name, c.verbose, app.universalNetworking, c.opts.apiHeaders, c.opts.setupKumactl)
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

	if c.opts.verifyKuma {
		return c.VerifyKuma()
	}

	return nil
}

func (c *VmUniversalCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *VmUniversalCluster) GetKumaCPLogs() map[string]string {
	if c.controlplane == nil { // This is required if the cp never succeeded to start
		return map[string]string{}
	}
	net := c.controlplane.Networking()
	if net.IP == "" {
		return map[string]string{
			"failed": "control plane app not found",
		}
	}
	out := make(map[string]string)

	bytes, err := os.ReadFile(net.StdErrFile)
	if err != nil {
		out["stderr"] = fmt.Sprintf("error reading kuma stderr: %s", err)
	} else {
		out["stderr"] = string(bytes)
	}

	bytes, err = os.ReadFile(net.StdOutFile)
	if err != nil {
		out["stdout"] = fmt.Sprintf("error reading kuma stdout: %v", err)
	} else {
		out["stdout"] = string(bytes)
	}
	return out
}

func (c *VmUniversalCluster) VerifyKuma() error {
	return c.controlplane.kumactl.RunKumactl("get", "meshes")
}

func (c *VmUniversalCluster) DeleteKuma() error {
	Logf("Deleting Kuma cluster %q", c.Name())
	if err := c.DeleteApp(AppModeCP); err != nil {
		return err
	}
	c.controlplane = nil
	return nil
}

func (c *VmUniversalCluster) GetKumactlOptions() *kumactl.KumactlOptions {
	return c.controlplane.kumactl
}

// K8s
func (c *VmUniversalCluster) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	return nil
}

func (c *VmUniversalCluster) CreateNamespace(namespace string) error {
	return nil
}

func (c *VmUniversalCluster) DeleteNamespace(string, ...NamespaceDeleteHookFunc) error {
	return nil
}

func (c *VmUniversalCluster) CreateDP(app *VmUniversalApp, name string, mesh string, ip string, dpyaml string, envs map[string]string, token string, builtindns bool, concurrency int, transparent bool) error {
	cpIp := c.controlplane.Networking().IP
	cpAddress := "https://" + net.JoinHostPort(cpIp, "5678")
	err := app.CreateDP(token, cpAddress, name, mesh, ip, dpyaml, builtindns, "", concurrency, envs, transparent)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.dataplanes = append(c.dataplanes, name)
	c.mutex.Unlock()
	return app.dpApp.Start()
}

func (c *VmUniversalCluster) CreateZoneIngress(app *VmUniversalApp, name, ip, dpyaml, token string, builtindns bool) error {
	err := app.CreateDP(token, c.controlplane.Networking().BootstrapAddress(), name, "", ip, dpyaml, builtindns, "ingress", 0, nil, false)
	if err != nil {
		return err
	}

	c.networking[Config.ZoneIngressApp] = app.universalNetworking

	c.createEnvoyTunnel(Config.ZoneIngressApp)
	return app.dpApp.Start()
}

func (c *VmUniversalCluster) CreateZoneEgress(
	app *UniversalApp,
	name, ip, dpYAML, token string,
	builtinDNS bool,
) error {
	err := app.CreateDP(token, c.controlplane.Networking().BootstrapAddress(), name, "", ip, dpYAML, builtinDNS, "egress", app.concurrency, nil, false, "")
	if err != nil {
		return err
	}

	c.networking[Config.ZoneEgressApp] = app.universalNetworking

	c.createEnvoyTunnel(Config.ZoneEgressApp)
	return app.dpApp.Start()
}

func (c *VmUniversalCluster) DeployApp(opt ...AppDeploymentOption) error {
	var opts AppDeploymentOptions
	opts.apply(opt...)
	token := opts.token
	transparent := pointer.Deref(opts.transparent)
	args := opts.appArgs

	if opts.verbose == nil {
		opts.verbose = &c.verbose
	}

	var caps []string
	if transparent {
		caps = append(caps, "NET_ADMIN", "NET_RAW")
	}

	Logf("IPV6 is %v", opts.isipv6)

	if app := c.GetApp(opts.name); app != nil {
		return errors.Errorf("app %q already exists", opts.name)
	}

	host, hostExists := c.hosts[opts.name]
	if !hostExists {
		return errors.Errorf("there is no host to deploy app %q", opts.name)
	}

	app, err := NewVmUniversalApp(c.t, opts.name, opts.mesh, host, false)
	if err != nil {
		return err
	}

	// We need to record the app before running any other options,
	// since those options might fail. If they do, we have a running
	// container that isn't fully configured, and we need it to be
	// recorded so that DismissCluster can clean it up.
	Logf("Started universal app %q on host %q (internal: %s)", opts.name, host.Address, host.InternalIPAddress)
	c.mutex.Lock()
	c.apps[opts.name] = app
	c.mutex.Unlock()

	if !opts.omitDataplane {
		if opts.kumactlFlow {
			dataplaneResource := template.Render(opts.appYaml, map[string]string{
				"name":    opts.name,
				"address": app.GetIP(),
			})
			err := c.GetKumactlOptions().KumactlApplyFromString(string(dataplaneResource))
			if err != nil {
				return err
			}
		}

		builtindns := pointer.DerefOr(opts.builtindns, true)

		ip := app.GetIP()

		var dataplaneResource string
		if opts.kumactlFlow {
			dataplaneResource = ""
		} else {
			dataplaneResource = opts.appYaml
		}

		if opts.mesh == "" {
			opts.mesh = "default"
		}
		if err := c.CreateDP(app, opts.name, opts.mesh, ip, dataplaneResource, opts.dpEnvs, token, builtindns, opts.concurrency, transparent); err != nil {
			return err
		}
	}

	if !opts.proxyOnly {
		app.CreateMainApp(strings.Join(args, " "))
		err = app.mainApp.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *VmUniversalCluster) GetApp(appName string) *VmUniversalApp {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.apps[appName]
}

func (c *VmUniversalCluster) GetDataplanes() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.dataplanes
}

func (c *VmUniversalCluster) DeleteApp(appname string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	app, ok := c.apps[appname]
	if !ok {
		return errors.Errorf("App %s not found for deletion", appname)
	}

	c.closePortForward(appname)

	if err := app.Stop(); err != nil {
		return err
	}
	delete(c.apps, appname)
	c.dataplanes = slices.DeleteFunc(c.dataplanes, func(s string) bool {
		return s == string(appname)
	})
	return nil
}

func (c *VmUniversalCluster) DeleteMesh(mesh string) error {
	now := time.Now()
	_, err := retry.DoWithRetryE(c.t, "remove mesh", DefaultRetries, 1*time.Second,
		func() (string, error) {
			return "", c.GetKumactlOptions().KumactlDelete("mesh", mesh, "")
		})
	Logf("mesh: " + mesh + " deleted in: " + time.Since(now).String())
	return err
}

func (c *VmUniversalCluster) DeleteMeshApps(mesh string) error {
	c.mutex.RLock()
	apps := util_maps.AllKeys(c.apps)
	c.mutex.RUnlock()
	for _, name := range apps {
		if c.GetApp(name).mesh == mesh {
			if err := c.DeleteApp(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *VmUniversalCluster) Exec(namespace, podName, appname string, cmd ...string) (string, string, error) {
	app := c.GetApp(appname)
	if app == nil {
		return "", "", errors.Errorf("App %s not found", appname)
	}
	runCmd := strings.Join(cmd, " ")
	Logf("Running on app %q command %q", appname, runCmd)
	return app.universalNetworking.RunCommand(runCmd)
}

// Kill a process running in this app
func (c *VmUniversalCluster) Kill(appname, cmd string) error {
	_, _, err := c.Exec("", "", appname, fmt.Sprintf("pkill -9 -f %q", cmd))
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		out, _, err := c.Exec("", "", appname, fmt.Sprintf("pgrep -f %q", cmd))
		var exitError *ssh.ExitError
		if errors.As(err, &exitError) {
			if exitError.ExitStatus() == 1 {
				return nil
			}
		}
		if err != nil {
			Logf("Failed to check for process %q: %v", appname, err)
			continue
		}
		Logf("Process %q still running output %q", appname, out)
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("process killed timed out")
}

func (c *VmUniversalCluster) GetTesting() testing.TestingT {
	return c.t
}

func (c *VmUniversalCluster) Deployment(name string) Deployment {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.deployments[name]
}

func (c *VmUniversalCluster) Deploy(deployment Deployment) error {
	c.mutex.Lock()
	c.deployments[deployment.Name()] = deployment
	c.mutex.Unlock()
	return deployment.Deploy(c)
}

func (c *VmUniversalCluster) DeleteDeployment(name string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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

func (c *VmUniversalCluster) GetUniversalNetworkingState() universal.NetworkingState {
	out := universal.NetworkingState{
		KumaCp: *c.controlplane.Networking(),
	}
	if ingressState := c.networking[Config.ZoneIngressApp]; ingressState != nil {
		out.ZoneIngress = *ingressState // nolint:govet
	}
	if egressState := c.networking[Config.ZoneEgressApp]; egressState != nil {
		out.ZoneEgress = *egressState // nolint:govet
	}
	return out // nolint:govet
}

func (c *VmUniversalCluster) AddNetworking(networking *universal.Networking, name string) error {
	c.networking[name] = networking
	c.createEnvoyTunnel(name)
	return nil
}

func (c *VmUniversalCluster) createEnvoyTunnel(name string) {
	c.envoyTunnels[name] = tunnel.NewUniversalEnvoyAdminTunnel(func(cmdName, cmd string) (string, error) {
		s, err := c.networking[name].NewSession(path.Join(c.name, "universal", "envoytunnel", name), cmdName, c.verbose, cmd)
		if err != nil {
			return "", err
		}
		err = s.Run()
		if err != nil {
			return "", err
		}
		b, err := os.ReadFile(s.StdOutFile())
		return string(b), err
	})
}

func (c *VmUniversalCluster) GetZoneEgressEnvoyTunnel() envoy_admin.Tunnel {
	t, err := c.GetZoneEgressEnvoyTunnelE()
	if err != nil {
		c.t.Fatal(err)
	}

	return t
}

func (c *VmUniversalCluster) GetZoneIngressEnvoyTunnel() envoy_admin.Tunnel {
	t, err := c.GetZoneIngressEnvoyTunnelE()
	if err != nil {
		c.t.Fatal(err)
	}

	return t
}

func (c *VmUniversalCluster) GetZoneEgressEnvoyTunnelE() (envoy_admin.Tunnel, error) {
	t, ok := c.envoyTunnels[Config.ZoneEgressApp]
	if !ok {
		return nil, errors.Errorf("no tunnel with name %+q", Config.ZoneEgressApp)
	}

	return t, nil
}

func (c *VmUniversalCluster) GetZoneIngressEnvoyTunnelE() (envoy_admin.Tunnel, error) {
	t, ok := c.envoyTunnels[Config.ZoneIngressApp]
	if !ok {
		return nil, errors.Errorf("no tunnel with name %+q", Config.ZoneIngressApp)
	}

	return t, nil
}

func (c *VmUniversalCluster) Install(fn InstallFunc) error {
	return fn(c)
}

func (c *VmUniversalCluster) SetCp(cp *VmUniversalControlPlane) {
	c.controlplane = cp
}

func (c *VmUniversalCluster) ZoneName() string {
	if c.opts.zoneName != "" {
		return c.opts.zoneName
	}
	return c.Name()
}

func (c *VmUniversalCluster) PortForward(appName, serviceName, remoteAddress string) (int, error) {
	app := c.apps[appName]
	stopChan := make(chan struct{})
	addr, err := app.universalNetworking.PortForward("localhost:0", remoteAddress, stopChan)
	if err != nil {
		return 0, err
	}

	pf := &VmPortForward{
		StopChannel:   stopChan,
		LocalPort:     addr.(*net.TCPAddr).Port,
		RemoteAddress: remoteAddress,
	}

	pfKey := fmt.Sprintf("%s/%s", appName, serviceName)
	c.portForwards[pfKey] = pf
	return pf.LocalPort, err
}

func (c *VmUniversalCluster) GetPortForward(appName, serviceName string) *VmPortForward {
	pfKey := fmt.Sprintf("%s/%s", appName, serviceName)
	return c.portForwards[pfKey]
}

func (c *VmUniversalCluster) closePortForward(appName string) {
	var pfsToDelete []string
	for key, _ := range c.portForwards {
		parts := strings.Split(key, "/")
		if parts[0] == appName {
			pfsToDelete = append(pfsToDelete, key)
		}
	}

	for _, key := range pfsToDelete {
		if pf, ok := c.portForwards[key]; ok {
			close(pf.StopChannel)
			delete(c.portForwards, key)
		}
	}
}
