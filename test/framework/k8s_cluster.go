package framework

import (
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	coordinationv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	resources_k8s "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	kuma_version "github.com/kumahq/kuma/v3/pkg/version"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/v3/test/framework/kumactl"
	"github.com/kumahq/kuma/v3/test/framework/portforward"
)

type K8sNetworkingState struct {
	ZoneEgress  portforward.Tunnel `json:"zoneEgress"`
	ZoneIngress portforward.Tunnel `json:"zoneIngress"`
	KumaCp      portforward.Tunnel `json:"kumaCp"`
	MADS        portforward.Tunnel `json:"mads"`
}

type K8sCluster struct {
	t                   testing.TestingT
	name                string
	kubeconfig          string
	controlplane        *K8sControlPlane
	forwardedPortsChans []chan struct{}
	verbose             bool
	deployments         map[string]Deployment
	mutex               sync.RWMutex // to protect deployments, portForwards, adminTunnels
	defaultTimeout      time.Duration
	defaultRetries      int
	opts                kumaDeploymentOptions
	portForwards        map[portforward.Spec]portforward.Tunnel
	adminTunnels        map[portforward.Spec]envoy_admin.Tunnel
}

var _ Cluster = &K8sCluster{}

func defaultKubeConfigPath(clusterName string) string {
	k8sType := K3dK8sType // default matches K8S_CLUSTER_TOOL in mk/e2e.new.mk
	if Config != nil {
		k8sType = Config.K8sType
	} else if value := os.Getenv("KUMA_K8S_TYPE"); value != "" {
		k8sType = K8sType(value)
	}

	var primary string
	switch k8sType {
	case KindK8sType:
		primary = os.ExpandEnv(fmt.Sprintf(defaultToolKubeConfigPathPattern, "kind", clusterName))
	case K3dK8sType, K3dCalicoK8sType:
		primary = os.ExpandEnv(fmt.Sprintf(defaultToolKubeConfigPathPattern, "k3d", clusterName))
	default:
		return os.ExpandEnv(fmt.Sprintf(legacyKubeConfigPathPattern, clusterName))
	}

	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	// Fall back to the pre-refactor filename so that clusters created before
	// the tool-prefixed naming scheme are still discovered.
	legacy := os.ExpandEnv(fmt.Sprintf(oldKindKubeConfigPathPattern, clusterName))
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return primary
}

func NewK8sCluster(t testing.TestingT, clusterName string, verbose bool) *K8sCluster {
	return &K8sCluster{
		t:                   t,
		name:                clusterName,
		kubeconfig:          defaultKubeConfigPath(clusterName),
		forwardedPortsChans: []chan struct{}{},
		verbose:             verbose,
		deployments:         map[string]Deployment{},
		defaultRetries:      Config.DefaultClusterStartupRetries,
		defaultTimeout:      Config.DefaultClusterStartupTimeout,
		portForwards:        map[portforward.Spec]portforward.Tunnel{},
		adminTunnels:        map[portforward.Spec]envoy_admin.Tunnel{},
	}
}

func (c *K8sCluster) WithTimeout(timeout time.Duration) Cluster {
	c.defaultTimeout = timeout

	return c
}

func (c *K8sCluster) WithKubeConfig(kubeConfigPath string) Cluster {
	c.kubeconfig = kubeConfigPath
	return c
}

func (c *K8sCluster) PortForwardApp(spec portforward.Spec) (portforward.Tunnel, error) {
	if err := spec.ValidateFullSpec(); err != nil {
		return portforward.Tunnel{}, err
	}

	c.mutex.RLock()
	existing := c.portForwards[spec]
	c.mutex.RUnlock()
	if existing.Endpoint != "" {
		return existing, nil
	}

	podName, err := PodNameOfApp(c, spec.AppName, spec.Namespace)
	if err != nil {
		return portforward.Tunnel{}, errors.Wrapf(
			err,
			"resolving target for port-forward failed: app %q in namespace %q could not be mapped to a Pod",
			spec.AppName,
			spec.Namespace,
		)
	}

	fwd, err := c.PortForward(k8s.ResourceTypePod, podName, spec.Namespace, spec.RemotePort)
	if err != nil {
		return portforward.Tunnel{}, errors.Wrapf(
			err,
			"failed to start port-forward to %s/%s in namespace %q (port %d)",
			k8s.ResourceTypePod.String(),
			podName,
			spec.Namespace,
			spec.RemotePort,
		)
	}

	c.mutex.Lock()
	existing = c.portForwards[spec]
	if existing.Endpoint != "" {
		c.mutex.Unlock()
		fwd.Close()
		return existing, nil
	}
	c.portForwards[spec] = fwd
	c.mutex.Unlock()

	return fwd, nil
}

func (c *K8sCluster) PortForward(
	resourceType k8s.KubeResourceType,
	resourceName string,
	namespace string,
	remotePort int,
) (portforward.Tunnel, error) {
	localPort, err := k8s.GetAvailablePortContextE(c.t, context.Background())
	if err != nil {
		return portforward.Tunnel{}, errors.Wrapf(
			err,
			"failed to allocate a local port for port-forward to %s/%s in namespace %q (remote port: %d)",
			resourceType,
			resourceName,
			namespace,
			remotePort,
		)
	}

	tnl := k8s.NewTunnel(c.GetKubectlOptions(namespace), resourceType, resourceName, localPort, remotePort)
	if err := tnl.ForwardPortE(c.t); err != nil {
		return portforward.Tunnel{}, errors.Wrapf(
			err,
			"failed to start port-forward to %s/%s in namespace %q (mapping %d -> %d)",
			resourceType,
			resourceName,
			namespace,
			localPort,
			remotePort,
		)
	}

	if tnl.Endpoint() == "" {
		return portforward.Tunnel{}, errors.Errorf(
			"empty endpoint after port-forward to %s/%s in %q (local %d -> remote %d); verify the target and port-forward",
			resourceType,
			resourceName,
			namespace,
			localPort,
			remotePort,
		)
	}

	return portforward.NewTunnel(tnl, tnl.Endpoint()), nil
}

func (c *K8sCluster) AddPortForward(portFwd portforward.Tunnel, spec portforward.Spec) {
	if err := spec.ValidateFullSpec(); err != nil {
		c.t.Fatalf("invalid port-forward spec: %s", err)
	}

	c.mutex.Lock()
	c.portForwards[spec] = portFwd
	c.mutex.Unlock()
}

func (c *K8sCluster) GetPortForward(spec portforward.Spec) portforward.Tunnel {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.portForwards[spec]
}

func (c *K8sCluster) ClosePortForwards(specs ...portforward.Spec) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, spec := range specs {
		for fwdSpec, tnl := range c.portForwards {
			if !fwdSpec.Matches(spec) {
				continue
			}

			tnl.Close()

			delete(c.portForwards, fwdSpec)
		}

		for tnlSpec := range c.adminTunnels {
			if tnlSpec.Matches(spec) {
				delete(c.adminTunnels, tnlSpec)
			}
		}
	}
}

func (c *K8sCluster) GetZoneEgressEnvoyTunnel() envoy_admin.Tunnel {
	return c.GetEnvoyAdminTunnel(Config.ZoneEgressApp, Config.KumaNamespace)
}

func (c *K8sCluster) GetZoneIngressEnvoyTunnel() envoy_admin.Tunnel {
	return c.GetEnvoyAdminTunnel(Config.ZoneIngressApp, Config.KumaNamespace)
}

// GetEnvoyAdminTunnel creates or returns an Envoy admin tunnel for any named
// app in a given namespace.
func (c *K8sCluster) GetEnvoyAdminTunnel(appName, namespace string) envoy_admin.Tunnel {
	tnl, err := c.GetOrCreateAdminTunnel(portforward.Spec{
		AppName:   appName,
		Namespace: namespace,
	})
	if err != nil {
		c.t.Fatal(err)
	}

	return tnl
}

func (c *K8sCluster) Verbose() bool {
	return c.verbose
}

func (c *K8sCluster) WithRetries(retries int) Cluster {
	c.defaultRetries = retries

	return c
}

func (c *K8sCluster) Name() string {
	return c.name
}

func (c *K8sCluster) Deployment(name string) Deployment {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.deployments[name]
}

func (c *K8sCluster) ApplyAndWaitServiceOnK8sCluster(namespace string, service string, yamlPath string) error {
	options := c.GetKubectlOptions(namespace)

	err := k8s.KubectlApplyContextE(c.t, context.Background(),
		options,
		yamlPath)
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailableContext(c.t, context.Background(),
		options,
		service,
		c.defaultRetries,
		c.defaultTimeout)

	return nil
}

func (c *K8sCluster) WaitNamespaceCreate(namespace string) {
	retry.DoWithRetryContext(c.t, context.Background(),
		"Wait the Kuma Namespace to terminate.",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			_, err := k8s.GetNamespaceContextE(c.t, context.Background(),
				c.GetKubectlOptions(),
				namespace)
			if err != nil {
				return "Namespace not available " + namespace, fmt.Errorf("Namespace %s still active", namespace)
			}

			return "Namespace " + namespace + " created", nil
		})
}

func WaitNamespaceDelete(cluster Cluster, namespace string) error {
	c, ok := cluster.(*K8sCluster)
	if !ok {
		return errors.New("cluster is not a K8sCluster")
	}

	_, err := retry.DoWithRetryContextE(c.t, context.Background(),
		fmt.Sprintf("Wait for %s Namespace to terminate.", namespace),
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			nsObject, err := k8s.GetNamespaceContextE(c.t, context.Background(),
				c.GetKubectlOptions(),
				namespace)
			if err != nil {
				if k8s_errors.IsNotFound(err) {
					return "Namespace " + namespace + " deleted", nil
				}
				return "Failed to get Namespace " + namespace, err
			}

			var conditions []string
			for _, condition := range nsObject.Status.Conditions {
				conditions = append(conditions, condition.String())
			}
			return "Namespace available " + namespace, fmt.Errorf("namespace %s still active, conditions: %s", namespace, strings.Join(conditions, ","))
		})
	if err != nil {
		var namespaceStr string
		nsObject, err := k8s.GetNamespaceContextE(c.t, context.Background(), c.GetKubectlOptions(), namespace)
		if err == nil {
			namespaceStr = "namespace object: " + nsObject.String()
		}
		all, _ := k8s.RunKubectlAndGetOutputContextE(c.t, context.Background(), c.GetKubectlOptions(namespace), "get", "all")
		return errors.Wrapf(err, "debug data: %s, all in namespace: %s", namespaceStr, all)
	}
	return err
}

func (c *K8sCluster) WaitNodeDelete(node string) (string, error) {
	return retry.DoWithRetryContextE(c.t, context.Background(),
		fmt.Sprintf("Wait for %s node to terminate.", node),
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			nodes, err := k8s.GetNodesContextE(c.t, context.Background(), c.GetKubectlOptions())
			if err != nil {
				return "Error getting node " + node, err
			}

			nodeAvailable := false
			for _, _node := range nodes {
				if _node.Name == node {
					nodeAvailable = true
				}
			}

			if nodeAvailable {
				return "Node available " + node, fmt.Errorf("node %s still active", node)
			}
			return "Node deleted " + node, nil
		})
}

func (c *K8sCluster) GetPodLogs(pod v1.Pod, podLogOpts v1.PodLogOptions) (string, error) {
	// creates the clientset
	clientset, err := k8s.GetKubernetesClientFromOptionsContextE(c.t, context.Background(), c.GetKubectlOptions())
	if err != nil {
		return "", errors.Wrapf(err, "error in getting access to K8S")
	}

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)

	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", errors.Wrapf(err, "error in opening stream")
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", errors.Wrap(err, "error in copy information from podLogs to buf")
	}

	str := buf.String()

	return str, nil
}

// deployKumaViaKubectl uses kubectl to install kuma
// using the resources from the `kumactl install control-plane` command
func (c *K8sCluster) deployKumaViaKubectl(mode string) error {
	if err := c.installCRDs(); err != nil {
		return err
	}
	yaml, err := c.yamlForKumaViaKubectl(mode)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringContextE(c.t, context.Background(),
		c.GetKubectlOptions(),
		yaml)
}

// installCRDs installs Kuma CRDs and waits until it's ready
// Usually it's immediately, but when we were installing CRDs and Kuma CP at the same time, sometimes we hit
// a problem when CP could not recognize CRDs in Kubernetes and CP was restarted.
func (c *K8sCluster) installCRDs() error {
	crds, err := c.GetKumactlOptions().RunKumactlAndGetOutputV(false, "install", "crds")
	if err != nil {
		return err
	}
	crdFile, err := k8s.StoreConfigToTempFileE(c.t, crds)
	if err != nil {
		return err
	}
	defer os.Remove(crdFile)
	// --server-side --force-conflicts handles CRDs left over from a previous
	// suite that were created without the last-applied-configuration annotation.
	if err := k8s.RunKubectlContextE(c.t, context.Background(), c.GetKubectlOptions(),
		"apply", "--server-side", "--force-conflicts", "-f", crdFile); err != nil {
		return err
	}

	regexPattern := `(?m)^\s*name:\s*([^\s]+\.kuma\.io)\b`
	re := regexp.MustCompile(regexPattern)
	matches := re.FindAllStringSubmatch(crds, -1)
	if matches == nil {
		return fmt.Errorf("no matches found")
	}

	for _, match := range matches {
		if len(match) > 1 {
			crdName := match[1]
			err := k8s.RunKubectlContextE(c.t, context.Background(), c.GetKubectlOptions(), "wait", "--for", "condition=established", "--timeout=60s", "crd/"+crdName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *K8sCluster) yamlForKumaViaKubectl(mode string) (string, error) {
	argsMap := map[string]string{
		"--mode":                      mode,
		"--namespace":                 Config.KumaNamespace,
		"--control-plane-repository":  Config.KumaCPImageRepo,
		"--dataplane-repository":      Config.KumaDPImageRepo,
		"--dataplane-init-repository": Config.KumaInitImageRepo,
	}
	var args []string
	if Config.KumaImageRegistry != "" {
		argsMap["--control-plane-registry"] = Config.KumaImageRegistry
		argsMap["--dataplane-registry"] = Config.KumaImageRegistry
		argsMap["--dataplane-init-registry"] = Config.KumaImageRegistry
	}

	if Config.KumaImageTag != "" && Config.KumaImageTag != "unknown" {
		argsMap["--control-plane-version"] = Config.KumaImageTag
		argsMap["--dataplane-version"] = Config.KumaImageTag
		argsMap["--dataplane-init-version"] = Config.KumaImageTag
	}

	if mode == core.Zone && c.opts.globalAddress != "" {
		argsMap["--zone"] = c.ZoneName()
		argsMap["--kds-global-address"] = c.opts.globalAddress
	}
	if !Config.UseLoadBalancer {
		argsMap["--use-node-port"] = ""
	}

	if c.opts.zoneIngress {
		argsMap["--ingress-enabled"] = ""
		argsMap["--ingress-use-node-port"] = ""
		args = append(args, "--set", fmt.Sprintf("%singress.resources.limits.cpu=null", Config.HelmSubChartPrefix))
	}

	if c.opts.zoneEgress {
		argsMap["--egress-enabled"] = ""
		args = append(args, "--set", fmt.Sprintf("%segress.resources.limits.cpu=null", Config.HelmSubChartPrefix))
		if Config.Debug {
			args = append(args, "--set", fmt.Sprintf("%segress.logLevel=debug", Config.HelmSubChartPrefix))
		}
	}

	if c.opts.cni {
		argsMap["--cni-enabled"] = ""
		argsMap["--cni-chained"] = ""
		argsMap["--cni-net-dir"] = Config.CNIConf.NetDir
		argsMap["--cni-bin-dir"] = Config.CNIConf.BinDir
		argsMap["--cni-conf-name"] = Config.CNIConf.ConfName
	}

	if Config.XDSApiVersion != "" {
		argsMap["--env-var"] = "KUMA_BOOTSTRAP_SERVER_API_VERSION=" + Config.XDSApiVersion
	}

	maps.Copy(argsMap, c.opts.ctlOpts)

	for k, v := range argsMap {
		args = append(args, k, v)
	}

	if Config.IPV6 {
		args = append(args,
			"--env-var", "KUMA_DNS_SERVER_CIDR=fd00:fd00::/64",
			"--env-var", "KUMA_IPAM_MESH_SERVICE_CIDR=fd00:fd01::/64",
			"--env-var", "KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR=fd00:fd02::/64",
			"--env-var", "KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR=fd00:fd03::/64",
		)
	}

	if Config.Debug {
		args = append(args, "--set", fmt.Sprintf("%scontrolPlane.logLevel=debug", Config.HelmSubChartPrefix))
	}

	for k, v := range c.opts.env {
		args = append(args, "--env-var", fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range c.opts.helmOpts {
		args = append(args, "--set", fmt.Sprintf("%s%s=%s", Config.HelmSubChartPrefix, k, v))
	}

	if c.opts.yamlConfig != "" {
		args = append(args, "--set", fmt.Sprintf("%s%s=%s", Config.HelmSubChartPrefix, "controlPlane.config", c.opts.yamlConfig))
	}

	if c.opts.memory != "" {
		args = append(args,
			"--set", fmt.Sprintf("%scontrolPlane.resources.limits.memory=%s", Config.HelmSubChartPrefix, c.opts.memory),
			"--set", fmt.Sprintf("%scontrolPlane.resources.requests.memory=%s", Config.HelmSubChartPrefix, c.opts.memory),
		)
	}

	return c.controlplane.InstallCP(args...)
}

func (c *K8sCluster) genValues(mode string) map[string]string {
	values := map[string]string{
		"controlPlane.mode":                      mode,
		"controlPlane.image.repository":          Config.KumaCPImageRepo,
		"dataPlane.image.repository":             Config.KumaDPImageRepo,
		"dataPlane.initImage.repository":         Config.KumaInitImageRepo,
		"controlPlane.defaults.skipMeshCreation": strconv.FormatBool(c.opts.skipDefaultMesh),
		"ingress.resources.limits.cpu":           "null",
		"egress.resources.limits.cpu":            "null",
	}
	if Config.KumaImageRegistry != "" {
		values["global.image.registry"] = Config.KumaImageRegistry
	}
	if Config.KumaImageTag != "" {
		values["global.image.tag"] = Config.KumaImageTag
	}

	if c.opts.cpReplicas != 0 {
		values["controlPlane.replicas"] = strconv.Itoa(c.opts.cpReplicas)
	}

	maps.Copy(values, c.opts.helmOpts)

	if Config.XDSApiVersion != "" {
		values["controlPlane.envVars.KUMA_BOOTSTRAP_SERVER_API_VERSION"] = Config.XDSApiVersion
	}

	if c.opts.cni {
		values["cni.image.repository"] = Config.KumaCNIImageRepo
		values["cni.enabled"] = "true"
		values["cni.chained"] = "true"
		values["cni.netDir"] = Config.CNIConf.NetDir
		values["cni.binDir"] = Config.CNIConf.BinDir
		values["cni.confName"] = Config.CNIConf.ConfName
	}

	if Config.IPV6 {
		values["controlPlane.envVars.KUMA_DNS_SERVER_CIDR"] = "fd00:fd00::/64"
		values["controlPlane.envVars.KUMA_IPAM_MESH_SERVICE_CIDR"] = "fd00:fd01::/64"
		values["controlPlane.envVars.KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR"] = "fd00:fd02::/64"
		values["controlPlane.envVars.KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR"] = "fd00:fd03::/64"
	}

	if Config.Debug {
		values["controlPlane.logLevel"] = "debug"
	}

	for key, value := range c.opts.env {
		values[fmt.Sprintf("controlPlane.envVars.%s", key)] = value
	}

	switch mode {
	case core.Global:
		if !Config.UseLoadBalancer {
			values["controlPlane.globalZoneSyncService.type"] = "NodePort"
		}
	case core.Zone:
		if c.opts.globalAddress != "" {
			values["controlPlane.zone"] = c.ZoneName()
			values["controlPlane.kdsGlobalAddress"] = c.opts.globalAddress
			values["controlPlane.tls.kdsZoneClient.skipVerify"] = "true"
		}
	}

	for _, value := range c.opts.helmOptsExcluded {
		delete(values, value)
	}

	prefixedValues := map[string]string{}
	for k, v := range values {
		prefixedValues[Config.HelmSubChartPrefix+k] = v
	}

	return prefixedValues
}

type helmFn func(testing.TestingT, context.Context, *helm.Options, string, string) error

func (c *K8sCluster) processViaHelm(mode string, fn helmFn) error {
	// run from test/e2e
	helmChart, err := filepath.Abs(Config.HelmChartPath)
	if err != nil {
		return err
	}

	if c.opts.helmChartPath != nil {
		helmChart = *c.opts.helmChartPath
	}

	values := c.genValues(mode)

	helmOpts := &helm.Options{
		SetValues:      values,
		KubectlOptions: c.GetKubectlOptions(Config.KumaNamespace),
	}

	if c.opts.helmChartVersion != "" && helmChart == Config.HelmChartName {
		helmChart, err = HelmChartFromRepoE(
			c.t,
			Config.HelmRepoUrl,
			filepath.Base(Config.HelmChartName),
			c.opts.helmChartVersion,
		)
		if err != nil {
			return err
		}
	} else if c.opts.helmChartVersion != "" {
		helmOpts.Version = c.opts.helmChartVersion
	}

	releaseName := c.opts.helmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueID()),
		)
	}

	// create the namespace if it does not exist
	if _, err = k8s.GetNamespaceContextE(c.t, context.Background(), c.GetKubectlOptions(), Config.KumaNamespace); err != nil {
		if err := k8s.CreateNamespaceContextE(c.t, context.Background(), c.GetKubectlOptions(), Config.KumaNamespace); err != nil {
			return err
		}
	}

	return fn(c.t, context.Background(), helmOpts, helmChart, releaseName)
}

// deployKumaViaHelm uses Helm to install kuma
// using the kuma helm chart
func (c *K8sCluster) deployKumaViaHelm(mode string) error {
	return c.processViaHelm(mode, helm.InstallContextE)
}

// upgradeKumaViaHelm uses Helm to upgrade kuma
// using the kuma helm chart
func (c *K8sCluster) upgradeKumaViaHelm(mode string) error {
	return c.processViaHelm(mode, helm.UpgradeContextE)
}

func (c *K8sCluster) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
	if mode == core.Zone {
		c.opts.apply(WithCtlOpts(Config.KumaZoneK8sCtlFlags))
	} else {
		c.opts.apply(WithCtlOpts(Config.KumaK8sCtlFlags))
	}
	c.opts.apply(opt...)

	replicas := 1
	if c.opts.cpReplicas != 0 {
		replicas = c.opts.cpReplicas
	}

	c.controlplane = NewK8sControlPlane(c.t, mode, c.name, c.kubeconfig, c, c.verbose, replicas, c.opts.apiHeaders)

	if mode == core.Zone {
		c.opts.env["KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY"] = "true"
	}

	if Config.Debug {
		dpEnvVarKey := "KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ENV_VARS"
		debugEnv := "KUMA_DATAPLANE_RUNTIME_ENVOY_LOG_LEVEL:debug"
		if envVars, ok := c.opts.env[dpEnvVarKey]; ok {
			c.opts.env[dpEnvVarKey] = envVars + "," + debugEnv
		} else {
			c.opts.env[dpEnvVarKey] = debugEnv
		}
	}

	var err error
	switch c.opts.installationMode {
	case KumactlInstallationMode:
		err = c.deployKumaViaKubectl(mode)
	case HelmInstallationMode:
		if mode == core.Global && Config.HelmGlobalExtraYaml != "" {
			if err := k8s.KubectlApplyFromStringContextE(c.t, context.Background(), c.GetKubectlOptions(), Config.HelmGlobalExtraYaml); err != nil {
				return nil
			}
		}
		err = c.deployKumaViaHelm(mode)
	default:
		err = errors.Errorf("invalid installation mode: %s", c.opts.installationMode)
	}
	if err != nil {
		return err
	}

	// First wait for kuma cp to start, then wait for the other components (they all need the CP anyway)
	if err := c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, replicas); err != nil {
		return errors.Wrap(err, "Kuma control-plane failed to start")
	}
	if err := c.WaitControlPlaneLeader(); err != nil {
		return errors.Wrap(err, "Kuma control-plane failed to become leader")
	}

	var wg sync.WaitGroup
	var appsToInstall []appInstallation
	if c.opts.cni {
		namespace := ""
		if c.opts.cniNamespace != "" {
			namespace = c.opts.cniNamespace
		} else {
			namespace = Config.CNINamespace
		}
		appsToInstall = append(appsToInstall, appInstallation{Config.CNIApp, namespace, 1, nil})
	}
	if c.opts.zoneIngress {
		appsToInstall = append(appsToInstall, appInstallation{Config.ZoneIngressApp, Config.KumaNamespace, 1, nil})
	}
	if c.opts.zoneEgress {
		appsToInstall = append(appsToInstall, appInstallation{Config.ZoneEgressApp, Config.KumaNamespace, 1, nil})
	}

	for i := range appsToInstall {
		idx := i
		wg.Go(func() {
			appsToInstall[idx].Outcome = c.WaitApp(appsToInstall[idx].Name, appsToInstall[idx].Namespace, appsToInstall[idx].Replicas)
		})
	}

	wg.Wait() // Because of the wait group we have a memory barrier which allows us to read Outcome in a thread safe manner.
	for _, appInstall := range appsToInstall {
		if appInstall.Outcome != nil {
			err = multierr.Append(err, errors.Wrapf(appInstall.Outcome, "%s failed to start", appInstall.Name))
		}
	}
	if err != nil {
		return err
	}

	if c.opts.zoneEgressEnvoyAdminTunnel {
		if !c.opts.zoneEgress {
			return errors.New("cannot create tunnel to zone egress's envoy admin without egress")
		}

		if _, err := c.GetOrCreateAdminTunnel(portforward.Spec{
			AppName:   Config.ZoneEgressApp,
			Namespace: Config.KumaNamespace,
		}); err != nil {
			return err
		}
	}

	if c.opts.zoneIngressEnvoyAdminTunnel {
		if !c.opts.zoneIngress {
			return errors.New("cannot create tunnel to zone ingress' envoy admin without ingress")
		}

		if _, err := c.GetOrCreateAdminTunnel(portforward.Spec{
			AppName:   Config.ZoneIngressApp,
			Namespace: Config.KumaNamespace,
		}); err != nil {
			return err
		}
	}

	if !c.opts.skipDefaultMesh {
		// wait for the mesh
		_, err = retry.DoWithRetryContextE(c.t, context.Background(),
			"get default mesh",
			c.defaultRetries,
			c.defaultTimeout,
			func() (string, error) {
				return k8s.RunKubectlAndGetOutputContextE(c.t, context.Background(), c.GetKubectlOptions(), "get", "mesh", "default")
			})
		if err != nil {
			deploymentDetails := ExtractDeploymentDetails(c.t, c.GetKubectlOptions(Config.KumaNamespace), Config.KumaServiceName)
			return &K8sDecoratedError{Err: err, Details: deploymentDetails}
		}
	}

	if err := c.controlplane.FinalizeAdd(); err != nil {
		return err
	}

	converter := resources_k8s.NewSimpleConverter()
	for name, updateFuncs := range c.opts.meshUpdateFuncs {
		for _, f := range updateFuncs {
			Logf("applying update function to mesh %q", name)
			err := UpdateKubeObject(c.GetTesting(), c.GetKubectlOptions(), "mesh", name,
				func(obj runtime.Object) runtime.Object {
					mesh := core_mesh.NewMeshResource()

					// The kubectl updater should have already converted the Kubernetes object
					// to a concrete type, so we can safely cast here.
					if err := converter.ToCoreResource(obj.(*mesh_k8s.Mesh), mesh); err != nil {
						panic(err.Error())
					}

					// Apply the conversion function.
					mesh.Spec = f(mesh.Spec)

					// Convert back to a Kubernetes resource.
					meshObj, err := converter.ToKubernetesObject(mesh)
					if err != nil {
						panic(err.Error())
					}

					// Note that at this point we might have lost some Kubernetes
					// resource metadata. That probably doesn't matter for Mesh objects
					// though.

					return meshObj
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

func (c *K8sCluster) UpgradeKuma(mode string, opt ...KumaDeploymentOption) error {
	if c.controlplane == nil {
		return errors.New("To upgrade Kuma has to be installed first")
	}

	c.opts.apply(opt...)
	if c.opts.cpReplicas != 0 {
		c.controlplane.replicas = c.opts.cpReplicas
	}

	if err := c.upgradeKumaViaHelm(mode); err != nil {
		return err
	}

	if err := c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, c.controlplane.replicas); err != nil {
		return err
	}
	if err := c.WaitControlPlaneLeader(); err != nil {
		return err
	}

	if c.opts.cni {
		if err := c.WaitApp(Config.CNIApp, Config.CNINamespace, 1); err != nil {
			return err
		}
	}

	if !c.opts.skipDefaultMesh {
		// wait for the mesh
		_, err := retry.DoWithRetryContextE(c.t, context.Background(),
			"get default mesh",
			c.defaultRetries,
			c.defaultTimeout,
			func() (string, error) {
				return k8s.RunKubectlAndGetOutputContextE(c.t, context.Background(), c.GetKubectlOptions(), "get", "mesh", "default")
			})
		if err != nil {
			return err
		}
	}

	if err := c.controlplane.FinalizeAdd(); err != nil {
		return err
	}

	return nil
}

// StartZoneIngress scales the replicas of a zone ingress to 1 and wait for it to complete.
func (c *K8sCluster) StartZoneIngress() error {
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=1", fmt.Sprintf("deployment/%s", Config.ZoneIngressApp)); err != nil {
		return err
	}
	if err := c.WaitApp(Config.ZoneIngressApp, Config.KumaNamespace, 1); err != nil {
		return err
	}

	if _, err := c.GetOrCreateAdminTunnel(portforward.Spec{
		AppName:   Config.ZoneIngressApp,
		Namespace: Config.KumaNamespace,
	}); err != nil {
		return err
	}

	return nil
}

// StopZoneIngress scales the replicas of a zone ingress to 0 and wait for it to complete. Useful for testing behavior when traffic goes through ingress but there is no instance.
func (c *K8sCluster) StopZoneIngress() error {
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=0", fmt.Sprintf("deployment/%s", Config.ZoneIngressApp)); err != nil {
		return err
	}

	c.ClosePortForwards(portforward.Spec{
		AppName:   Config.ZoneIngressApp,
		Namespace: Config.KumaNamespace,
	})

	_, err := retry.DoWithRetryContextE(c.t, context.Background(),
		"wait for zone ingress to be down",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			pods := c.getPods(Config.KumaNamespace, Config.ZoneIngressApp)
			if len(pods) == 0 {
				return "Done", nil
			}
			var names []string
			for _, p := range pods {
				names = append(names, p.Name)
			}
			return "", fmt.Errorf("some pods are still present count: '%s'", strings.Join(names, ","))
		},
	)
	return err
}

// StartZoneEngress scales the replicas of a zone engress to 1 and wait for it to complete.
func (c *K8sCluster) StartZoneEgress() error {
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=1", fmt.Sprintf("deployment/%s", Config.ZoneEgressApp)); err != nil {
		return err
	}
	if err := c.WaitApp(Config.ZoneEgressApp, Config.KumaNamespace, 1); err != nil {
		return err
	}

	if _, err := c.GetOrCreateAdminTunnel(portforward.Spec{
		AppName:   Config.ZoneEgressApp,
		Namespace: Config.KumaNamespace,
	}); err != nil {
		return err
	}

	return nil
}

// StopZoneEgress scales the replicas of a zone egress to 0 and wait for it to complete. Useful for testing behavior when traffic goes through egress but there is no instance.
func (c *K8sCluster) StopZoneEgress() error {
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=0", fmt.Sprintf("deployment/%s", Config.ZoneEgressApp)); err != nil {
		return err
	}

	c.ClosePortForwards(portforward.Spec{
		AppName:   Config.ZoneEgressApp,
		Namespace: Config.KumaNamespace,
	})

	_, err := retry.DoWithRetryContextE(c.t, context.Background(),
		"wait for zone egress to be down",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			pods := c.getPods(Config.KumaNamespace, Config.ZoneEgressApp)
			if len(pods) == 0 {
				return "Done", nil
			}
			names := []string{}
			for _, p := range pods {
				names = append(names, p.Name)
			}
			return "", fmt.Errorf("some pods are still present count: '%s'", strings.Join(names, ","))
		},
	)
	return err
}

// StopControlPlane scales the replicas of a control plane to 0 and wait for it to complete. Useful for testing restarts in combination with RestartControlPlane.
func (c *K8sCluster) StopControlPlane() error {
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=0", fmt.Sprintf("deployment/%s", Config.KumaServiceName)); err != nil {
		return err
	}
	_, err := retry.DoWithRetryContextE(c.t, context.Background(),
		"wait for control-plane to be down",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			pods := c.controlplane.GetKumaCPPods()
			if len(pods) == 0 {
				return "Done", nil
			}
			names := []string{}
			for _, p := range pods {
				names = append(names, p.Name)
			}
			return "", fmt.Errorf("some pods are still present count: '%s'", strings.Join(names, ","))
		},
	)
	return err
}

// RestartControlPlane scales the replicas of a control plane back to `c.controlplane.replicas` and waits for it to be running. Useful for testing restarts in combination with StopControlPlane.
func (c *K8sCluster) RestartControlPlane() error {
	if c.controlplane.replicas == 0 {
		return errors.New("replica count is 0, can't restart the control-plane")
	}
	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), "scale", fmt.Sprintf("--replicas=%d", c.controlplane.replicas), fmt.Sprintf("deployment/%s", Config.KumaServiceName)); err != nil {
		return err
	}
	if err := c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, c.controlplane.replicas); err != nil {
		return err
	}
	if err := c.WaitControlPlaneLeader(); err != nil {
		return err
	}

	if err := c.controlplane.FinalizeAdd(); err != nil {
		return err
	}

	if c.opts.verifyKuma {
		return c.VerifyKuma()
	}

	return nil
}

func (c *K8sCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *K8sCluster) GetKumaCPLogs() map[string]string {
	if c.controlplane == nil { // This is required if the cp never succeeded to start
		return map[string]string{}
	}
	pods := c.controlplane.GetKumaCPPods()
	if len(pods) < 1 {
		return map[string]string{"failed": "no kuma-cp pods found for logs"}
	}

	out := make(map[string]string)
	for _, p := range pods {
		stdoutKey := fmt.Sprintf("%s/stdout", p.Name)
		log, err := c.GetPodLogs(p, v1.PodLogOptions{})
		if err != nil {
			out[stdoutKey] = fmt.Sprintf("failed to retrieve logs %v", err)
			continue
		}
		out[stdoutKey] = log

		log, err = c.GetPodLogs(p, v1.PodLogOptions{
			Previous: true,
		})
		if err == nil {
			out[fmt.Sprintf("%s/stdout_previous", p.Name)] = log
		}
	}

	return out
}

func (c *K8sCluster) VerifyKuma() error {
	if err := c.controlplane.VerifyKumaGUI(); err != nil {
		return err
	}

	if err := c.controlplane.VerifyKumaREST(); err != nil {
		return err
	}

	if err := c.controlplane.VerifyKumaCtl(); err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailableContext(c.GetTesting(), context.Background(), c.GetKubectlOptions(Config.KumaNamespace), Config.KumaServiceName, DefaultRetries, DefaultTimeout)

	return nil
}

func (c *K8sCluster) deleteCRDs() error {
	out, err := k8s.RunKubectlAndGetOutputContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(), "get", "crds", "-o", "name")
	if err != nil {
		return err
	}

	var kumaCRDs []string
	for l := range strings.SplitSeq(out, "\n") {
		if strings.Contains(l, "kuma.io") {
			kumaCRDs = append(kumaCRDs, l)
		}
	}
	if len(kumaCRDs) == 0 {
		return nil
	}

	if err := k8s.RunKubectlContextE(
		c.GetTesting(), context.Background(),
		c.GetKubectlOptions(),
		slices.Concat([]string{"delete"}, kumaCRDs)...,
	); err != nil {
		return err
	}

	// Wait for CRDs to be fully removed so the next suite can install cleanly
	return k8s.RunKubectlContextE(
		c.GetTesting(), context.Background(),
		c.GetKubectlOptions(),
		slices.Concat([]string{"wait", "--for=delete", "--timeout=60s"}, kumaCRDs)...,
	)
}

func (c *K8sCluster) deleteKumaViaHelm() error {
	if c.opts.helmReleaseName == "" {
		return errors.New("must supply a helm release name for cleanup")
	}

	helmOpts := &helm.Options{
		KubectlOptions: c.GetKubectlOptions(Config.KumaNamespace),
	}

	var errs error
	if err := helm.DeleteContextE(c.t, context.Background(), helmOpts, c.opts.helmReleaseName, true); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err := c.DeleteNamespace(Config.KumaNamespace); err != nil {
		errs = multierr.Append(errs, err)
	}

	// there is no CRDs in universal env
	if c.opts.helmOpts["controlPlane.environment"] != "universal" {
		// HELM does not remove CRDs therefore we need to do it manually.
		// It's important to remove CRDs to get rid of all "instances" of CRDs like default Mesh etc.
		if err := c.deleteCRDs(); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func (c *K8sCluster) getPods(namespace string, appName string) []v1.Pod {
	return k8s.ListPodsContext(c.t, context.Background(),
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + appName,
		},
	)
}

func (c *K8sCluster) deleteKumaViaKumactl() error {
	yaml, err := c.yamlForKumaViaKubectl(c.controlplane.mode)
	if err != nil {
		return err
	}

	_ = k8s.KubectlDeleteFromStringContextE(c.t, context.Background(), c.GetKubectlOptions(), yaml)

	if err := WaitNamespaceDelete(c, Config.KumaNamespace); err != nil {
		return err
	}

	return c.deleteCRDs()
}

func (c *K8sCluster) DeleteKuma() error {
	if c.controlplane == nil {
		return nil
	}

	c.controlplane.ClosePortForwards()
	var err error
	switch c.opts.installationMode {
	case HelmInstallationMode:
		err = c.deleteKumaViaHelm()
	case KumactlInstallationMode:
		err = c.deleteKumaViaKumactl()
	}

	return err
}

func (c *K8sCluster) GetKumactlOptions() *kumactl.KumactlOptions {
	return c.controlplane.kumactl
}

func (c *K8sCluster) RefreshKumaCPPortForwards() error {
	if c.controlplane == nil {
		return nil
	}

	return c.controlplane.RefreshPortForwards()
}

func (c *K8sCluster) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	options := &k8s.KubectlOptions{
		ConfigPath: c.kubeconfig,
	}
	for _, ns := range namespace {
		options.Namespace = ns
		break
	}

	return options
}

func (c *K8sCluster) GetK8sVersion() (*semver.Version, error) {
	v, err := k8s.GetKubernetesClusterVersionWithOptionsContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions())
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(`^v?(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+).*$`)
	match := r.FindStringSubmatch(v)

	paramsMap := make(map[string]uint64)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			u64, err := strconv.ParseUint(match[i], 10, 32)
			if err != nil {
				return nil, errors.Wrapf(err, "parsing version %s failed", name)
			}

			paramsMap[name] = u64
		}
	}

	return semver.New(paramsMap["Major"], paramsMap["Minor"], paramsMap["Patch"], "", ""), nil
}

func (c *K8sCluster) CreateNamespace(namespace string) error {
	err := k8s.CreateNamespaceContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(), namespace)
	if err != nil {
		return err
	}

	c.WaitNamespaceCreate(namespace)

	return nil
}

func DeleteAllResources(kinds string, flags ...string) NamespaceDeleteHookFunc {
	return func(c Cluster, namespace string) error {
		baseArgs := []string{"delete", "--all", kinds}

		return k8s.RunKubectlContextE(
			c.GetTesting(), context.Background(),
			c.GetKubectlOptions(namespace),
			slices.Concat(baseArgs, flags)...,
		)
	}
}

// DeleteNamespace deletes a namespace and waits for it to be fully removed. It uses the
// default hook that force deletes services and pods for faster deletion and appends a wait
// hook to ensure the namespace is gone before returning.
func (c *K8sCluster) DeleteNamespace(namespace string, hooks ...NamespaceDeleteHookFunc) error {
	return c.TriggerDeleteNamespace(namespace, append(hooks, WaitNamespaceDelete)...)
}

// TriggerDeleteNamespace deletes a namespace with a default hook that force deletes all
// services and pods, making the namespace removal significantly faster. Additional custom
// hooks can be provided to run after deletion.
func (c *K8sCluster) TriggerDeleteNamespace(namespace string, hooks ...NamespaceDeleteHookFunc) error {
	baseHooks := []NamespaceDeleteHookFunc{
		DeleteAllResources("services,pods", "--grace-period=0", "--force"),
	}

	return c.TriggerDeleteNamespaceCustomHooks(namespace, slices.Concat(baseHooks, hooks)...)
}

// TriggerDeleteNamespaceCustomHooks deletes a namespace without the default hook that force
// deletes all services and pods. This means the namespace deletion might take longer compared
// to TriggerDeleteNamespace, which removes resources aggressively to speed up the process.
// Custom hooks can be provided to run additional actions after deletion.
func (c *K8sCluster) TriggerDeleteNamespaceCustomHooks(namespace string, hooks ...NamespaceDeleteHookFunc) error {
	if err := k8s.DeleteNamespaceContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(), namespace); err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	var errs []error
	for _, fn := range hooks {
		errs = append(errs, fn(c, namespace))
	}

	return std_errors.Join(errs...)
}

func (c *K8sCluster) DeleteMesh(mesh string) error {
	now := time.Now()
	_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "remove mesh", c.defaultRetries, c.defaultTimeout,
		func() (string, error) {
			return "", k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(), "delete", "mesh", mesh)
		})
	Logf("%s", "mesh: "+mesh+" deleted in: "+time.Since(now).String())
	return err
}

func (c *K8sCluster) GetClusterIP(serviceName, namespace string) (string, error) {
	service, err := k8s.GetServiceContextE(
		c.t, context.Background(),
		c.GetKubectlOptions(namespace),
		serviceName,
	)
	if err != nil {
		return "", err
	}
	return service.Spec.ClusterIP, nil
}

func (c *K8sCluster) GetLBIngressIP(serviceName, namespace string) (string, error) {
	service, err := k8s.GetServiceContextE(
		c.t, context.Background(),
		c.GetKubectlOptions(namespace),
		serviceName,
	)
	if err != nil {
		return "", err
	}
	ingress := service.Status.LoadBalancer.Ingress
	if len(ingress) == 0 {
		return "", errors.Errorf("ingress information not found on the load balancer service '%s'", serviceName)
	}
	return ingress[0].IP, nil
}

func (c *K8sCluster) DeployApp(opt ...AppDeploymentOption) error {
	var opts appDeploymentOptions

	opts.apply(opt...)

	namespace := opts.namespace
	appname := opts.appname

	retry.DoWithRetryContext(c.GetTesting(), context.Background(), "apply "+appname+" svc", c.defaultRetries, c.defaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyContextE(c.GetTesting(), context.Background(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+"-svc.yaml"))
			return "", err
		})

	k8s.WaitUntilServiceAvailableContext(c.GetTesting(), context.Background(),
		c.GetKubectlOptions(namespace),
		appname, c.defaultRetries, c.defaultTimeout)

	retry.DoWithRetryContext(c.GetTesting(), context.Background(), "apply "+appname, c.defaultRetries, c.defaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyContextE(c.GetTesting(), context.Background(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+".yaml"))
			return "", err
		})

	k8s.WaitUntilNumPodsCreatedContext(c.GetTesting(), context.Background(),
		c.GetKubectlOptions(),
		metav1.ListOptions{
			LabelSelector: "app=" + appname,
		},
		1, c.defaultRetries, c.defaultTimeout)

	return nil
}

func (c *K8sCluster) DeleteApp(namespace, appname string) error {
	err := k8s.KubectlDeleteContextE(c.GetTesting(), context.Background(),
		c.GetKubectlOptions(namespace),
		filepath.Join("testdata", appname+"-svc.yaml"))
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteContextE(c.GetTesting(), context.Background(),
		c.GetKubectlOptions(namespace),
		filepath.Join("testdata", appname+".yaml"))
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) GetTesting() testing.TestingT {
	return c.t
}

func (c *K8sCluster) DismissCluster() error {
	var errs error
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for name, deployment := range c.deployments {
		if err := deployment.Delete(c); err != nil {
			errs = multierr.Append(errs, err)
		}
		delete(c.deployments, name)
	}
	return errs
}

func (c *K8sCluster) Deploy(deployment Deployment) error {
	c.mutex.Lock()
	c.deployments[deployment.Name()] = deployment
	c.mutex.Unlock()
	return deployment.Deploy(c)
}

func (c *K8sCluster) DeleteDeployment(name string) error {
	c.mutex.RLock()
	deployment, ok := c.deployments[name]
	c.mutex.RUnlock()
	if !ok {
		return errors.Errorf("deployment %s not found", name)
	}
	if err := deployment.Delete(c); err != nil {
		return err
	}
	c.mutex.Lock()
	delete(c.deployments, name)
	c.mutex.Unlock()
	return nil
}

func (c *K8sCluster) WaitApp(name, namespace string, replicas int) error {
	err := k8s.WaitUntilNumPodsCreatedContextE(c.t, context.Background(),
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + name,
		},
		replicas,
		c.defaultRetries,
		c.defaultTimeout)
	if err != nil {
		deploymentDetails := ExtractDeploymentDetails(c.t, c.GetKubectlOptions(namespace), name)
		return &K8sDecoratedError{Err: err, Details: deploymentDetails}
	}

	pods := k8s.ListPodsContext(c.t, context.Background(),
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + name,
		},
	)
	if len(pods) < replicas {
		return errors.Errorf("%s pods: %d. expected %d", name, len(pods), replicas)
	}

	for i := range replicas {
		podError := k8s.WaitUntilPodAvailableContextE(c.t, context.Background(),
			c.GetKubectlOptions(namespace),
			pods[i].Name,
			c.defaultRetries,
			c.defaultTimeout)

		if podError != nil {
			podDetails := ExtractPodDetails(c.t, c.GetKubectlOptions(namespace), pods[i].Name)
			return &K8sDecoratedError{Err: podError, Details: podDetails}
		}
	}
	return nil
}

func (c *K8sCluster) WaitControlPlaneLeader() error {
	if c.usesUniversalLeaderElection() {
		return c.waitUniversalControlPlaneLeader()
	}
	return c.waitKubernetesControlPlaneLeader()
}

func (c *K8sCluster) usesUniversalLeaderElection() bool {
	return c.opts.helmOpts["controlPlane.environment"] == "universal"
}

func (c *K8sCluster) waitUniversalControlPlaneLeader() error {
	_, err := retry.DoWithRetryContextE(c.t,
		context.Background(),
		"wait for universal control-plane leader",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			pods, err := k8s.ListPodsContextE(c.t,
				context.Background(),
				c.GetKubectlOptions(Config.KumaNamespace),
				metav1.ListOptions{
					LabelSelector: "app=" + Config.KumaServiceName,
				},
			)
			if err != nil {
				return "", err
			}
			if len(pods) == 0 {
				return "", errors.New("no control-plane pods found")
			}

			var checkedPods []string
			for _, pod := range pods {
				fwd, err := c.PortForward(k8s.ResourceTypePod, pod.Name, pod.Namespace, 5680)
				if err != nil {
					return "", err
				}

				metrics, metricsErr := getControlPlaneMetrics(fwd.Endpoint)
				fwd.Close()
				if metricsErr != nil {
					return "", metricsErr
				}
				if hasLeaderMetric(metrics) {
					return pod.Name, nil
				}
				checkedPods = append(checkedPods, pod.Name)
			}

			return "", errors.Errorf("no universal control-plane leader found in metrics for pods: %s", strings.Join(checkedPods, ","))
		})
	if err != nil {
		deploymentDetails := ExtractDeploymentDetails(c.t, c.GetKubectlOptions(Config.KumaNamespace), Config.KumaServiceName)
		return &K8sDecoratedError{Err: err, Details: deploymentDetails}
	}
	return nil
}

func getControlPlaneMetrics(endpoint string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+endpoint+"/metrics", http.NoBody)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected metrics status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func hasLeaderMetric(metrics string) bool {
	for line := range strings.SplitSeq(metrics, "\n") {
		if strings.HasPrefix(line, "leader{") && strings.HasSuffix(line, "} 1") {
			return true
		}
	}
	return false
}

func (c *K8sCluster) waitKubernetesControlPlaneLeader() error {
	clientset, err := k8s.GetKubernetesClientFromOptionsContextE(c.t, context.Background(), c.GetKubectlOptions())
	if err != nil {
		return errors.Wrapf(err, "error in getting access to K8S")
	}

	_, err = retry.DoWithRetryContextE(c.t,
		context.Background(),
		"wait for control-plane leader",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			pods, err := k8s.ListPodsContextE(c.t,
				context.Background(),
				c.GetKubectlOptions(Config.KumaNamespace),
				metav1.ListOptions{
					LabelSelector: "app=" + Config.KumaServiceName,
				},
			)
			if err != nil {
				return "", err
			}
			if len(pods) == 0 {
				return "", errors.New("no control-plane pods found")
			}

			podNames := make([]string, 0, len(pods))
			for _, pod := range pods {
				podNames = append(podNames, pod.Name)
			}

			lease, err := clientset.CoordinationV1().Leases(Config.KumaNamespace).Get(
				context.Background(),
				"cp-leader-lease",
				metav1.GetOptions{},
			)
			if err != nil {
				return "", err
			}

			holder := lease.Spec.HolderIdentity
			if leaseHeldByCurrentPod(holder, podNames) {
				return *holder, nil
			}
			if holder == nil || *holder == "" {
				return "", errors.Errorf("control-plane lease has no holder, want one of current pods: %s", strings.Join(podNames, ","))
			}

			if controlPlaneLeaseExpired(lease, time.Now()) {
				if err := clientset.CoordinationV1().Leases(Config.KumaNamespace).Delete(
					context.Background(),
					lease.Name,
					controlPlaneLeaseDeleteOptions(lease),
				); err != nil {
					return "", err
				}
				return "", errors.Errorf("deleted stale control-plane lease held by %q, want one of current pods: %s", *holder, strings.Join(podNames, ","))
			}
			return "", errors.Errorf("control-plane lease held by %q is not expired yet, want one of current pods: %s", *holder, strings.Join(podNames, ","))
		})
	if err != nil {
		deploymentDetails := ExtractDeploymentDetails(c.t, c.GetKubectlOptions(Config.KumaNamespace), Config.KumaServiceName)
		return &K8sDecoratedError{Err: err, Details: deploymentDetails}
	}
	return nil
}

func leaseHeldByCurrentPod(holder *string, podNames []string) bool {
	if holder == nil {
		return false
	}
	for _, podName := range podNames {
		if strings.HasPrefix(*holder, podName+"_") {
			return true
		}
	}
	return false
}

func controlPlaneLeaseExpired(lease *coordinationv1.Lease, now time.Time) bool {
	if lease.Spec.RenewTime == nil || lease.Spec.LeaseDurationSeconds == nil {
		return false
	}
	deadline := lease.Spec.RenewTime.Add(time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second)
	return !deadline.After(now)
}

func controlPlaneLeaseDeleteOptions(lease *coordinationv1.Lease) metav1.DeleteOptions {
	resourceVersion := lease.ResourceVersion
	return metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{
			ResourceVersion: &resourceVersion,
		},
	}
}

func (c *K8sCluster) Install(fn InstallFunc) error {
	return fn(c)
}

func (c *K8sCluster) SetCP(cp *K8sControlPlane) {
	c.controlplane = cp
}

// CreateNode creates a new node
// warning: there seems to be a bug in k3s1 v1.19.16 so that each tests needs a unique node name
func (c *K8sCluster) CreateNode(name string, label string) error {
	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType:
		container := c.name
		createCmd := exec.CommandContext(context.Background(), "k3d", "node", "create", name, "-c", container, "--k3s-node-label", label)
		createCmd.Stdout = os.Stdout
		return createCmd.Run()
	case KindK8sType, AwsK8sType, AzureK8sType:
		return errors.New("creating new node not available for " + string(Config.K8sType))
	default:
		return errors.New("unknown kubernetes type")
	}
}

func (c *K8sCluster) LoadImages(names ...string) error {
	// 3 retries with 0 backoff was too tight: a single transient docker
	// daemon hiccup blew through all attempts before recovery. Bumped to
	// 5 attempts with 5s backoff so a brief image-import failure does
	// not fail the whole test suite.
	_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "load images", 5, 5*time.Second, func() (string, error) {
		err := c.loadImages(names...)
		return "Loaded images " + strings.Join(names, ", "), err
	})
	return err
}

func (c *K8sCluster) loadImages(names ...string) error {
	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType:
		version := kuma_version.Build.Version

		var fullImageNames []string
		for _, image := range names {
			fullImageNames = append(fullImageNames, Config.KumaImageRegistry+"/"+image+":"+version)
		}
		defaultArgs := []string{"image", "import", "-m", "direct", "-c", c.name}
		allArgs := append(defaultArgs, fullImageNames...)

		// Put a timeout of 1 minute to this command, because for some reason the command can be stuck with
		// ERRO[0004] Failed to copy read stream. write unix @->/run/docker.sock: use of closed network connection
		ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancelFn()
		importCmd := exec.CommandContext(ctx, "k3d", allArgs...)
		importCmd.Stdout = os.Stdout
		importCmd.Stderr = os.Stderr
		return importCmd.Run()
	case KindK8sType, AwsK8sType, AzureK8sType:
		return errors.New("loading new images to a node not available for " + string(Config.K8sType))
	default:
		return errors.New("unknown kubernetes type")
	}
}

// PreloadImages pulls each fully-qualified image with `docker pull` (in
// parallel) and imports them into the cluster's container runtime so pods
// never need to fetch them at runtime. Use this before installing Helm charts
// whose pods reference external images — it removes the dominant flake mode
// on CI runners where ghcr.io / docker.io return slow or cancel the request,
// surfacing as ImagePullBackOff after the test's wait budget is exhausted.
//
// Only k3d (incl. calico variant) and kind are supported because
// `k3d image import` / `kind load docker-image` are local-runtime operations.
// On other cluster types (AWS/Azure, used in downstream smoke tests like
// kong-mesh-smoke / mink-charts) this is a no-op: those nodes pull from the
// registry directly and the helper has nothing to shortcut.
//
// docker pull is cheap when the image is already present locally. Images
// referenced by `<repo>:<tag>@sha256:<digest>` are also retagged locally to
// the bare `<repo>:<tag>` form, because `docker pull` of the digest form
// stores the image with empty RepoTags, which `k3d image import` would then
// import as `<none>:<none>` and kubelet would still re-pull at runtime.
func (c *K8sCluster) PreloadImages(images ...string) error {
	if len(images) == 0 {
		return nil
	}
	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType, KindK8sType:
		// supported, continue
	default:
		return nil
	}

	pullCtx, cancelPull := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelPull()
	var wg sync.WaitGroup
	pullErrs := make([]error, len(images))
	for i, img := range images {
		wg.Add(1)
		go func(i int, img string) {
			defer wg.Done()
			localRef := img
			if tagged, _, ok := splitDigestRef(img); ok {
				localRef = tagged
			}
			if exec.CommandContext(pullCtx, "docker", "image", "inspect", localRef).Run() == nil {
				return
			}
			_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "pull image "+img, 5, 5*time.Second, func() (string, error) {
				pullCmd := exec.CommandContext(pullCtx, "docker", "pull", "--quiet", img)
				out, err := pullCmd.CombinedOutput()
				return strings.TrimSpace(string(out)), err
			})
			if err != nil {
				pullErrs[i] = errors.Wrapf(err, "docker pull %s", img)
				return
			}
			// If the image is digest-pinned, also tag it locally as the bare
			// `<repo>:<tag>` form so that `docker save` / `k3d image import`
			// preserves the tag in the cluster's containerd image store.
			tagged, digestRef, ok := splitDigestRef(img)
			if !ok {
				return
			}
			tagCmd := exec.CommandContext(pullCtx, "docker", "tag", digestRef, tagged)
			if out, err := tagCmd.CombinedOutput(); err != nil {
				pullErrs[i] = errors.Wrapf(err, "docker tag %s %s: %s", digestRef, tagged, strings.TrimSpace(string(out)))
			}
		}(i, img)
	}
	wg.Wait()
	if err := std_errors.Join(pullErrs...); err != nil {
		return errors.Wrap(err, "preload images: failed to pull on host")
	}

	// Use bare `<repo>:<tag>` form for import; the cluster's containerd will
	// then have the tagged entry, and kubelet's digest-aware lookup against
	// it succeeds because content addressability matches.
	importImages := make([]string, len(images))
	for i, img := range images {
		if tagged, _, ok := splitDigestRef(img); ok {
			importImages[i] = tagged
		} else {
			importImages[i] = img
		}
	}

	if manifest := os.Getenv("KUMA_E2E_PRELOAD_MANIFEST"); manifest != "" {
		recordPreloadedImages(manifest, importImages)
	}

	// Match `LoadImages` retry strategy: a single transient docker / k3d
	// hiccup shouldn't fail the whole preload. 5 attempts with 5s backoff.
	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType:
		_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "k3d image import", 5, 5*time.Second, func() (string, error) {
			args := append([]string{"image", "import", "-m", "direct", "-c", c.name}, importImages...)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "k3d", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return "", errors.Wrapf(err, "k3d image import (images=%v): %s", importImages, strings.TrimSpace(string(out)))
			}
			return "imported " + strings.Join(importImages, ", "), nil
		})
		return err
	case KindK8sType:
		_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "kind load docker-image", 5, 5*time.Second, func() (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			for _, img := range importImages {
				cmd := exec.CommandContext(ctx, "kind", "load", "docker-image", img, "--name", c.name)
				out, err := cmd.CombinedOutput()
				if err != nil {
					return "", errors.Wrapf(err, "kind load %s: %s", img, strings.TrimSpace(string(out)))
				}
			}
			return "loaded " + strings.Join(importImages, ", "), nil
		})
		return err
	default:
		return nil
	}
}

// splitDigestRef parses an image reference of the form `<repo>:<tag>@sha256:<digest>`
// into the bare `<repo>:<tag>` form and the `<repo>@sha256:<digest>` form. The
// third return value is false when the input has no digest component or no
// tag component.
func splitDigestRef(image string) (string, string, bool) {
	atIdx := strings.LastIndex(image, "@")
	if atIdx <= 0 {
		return "", "", false
	}
	base, digestPart := image[:atIdx], image[atIdx:]
	// `<repo>` may contain a port (e.g. `host:5000/repo`) which uses a
	// colon, so look for the colon to the right of the last slash.
	slashIdx := strings.LastIndex(base, "/")
	colonIdx := strings.LastIndex(base[slashIdx+1:], ":")
	if colonIdx == -1 {
		return "", "", false
	}
	colonIdx += slashIdx + 1
	repo := base[:colonIdx]
	return base, repo + digestPart, true
}

var preloadManifestMu sync.Mutex

func recordPreloadedImages(path string, images []string) {
	preloadManifestMu.Lock()
	defer preloadManifestMu.Unlock()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	for _, img := range images {
		fmt.Fprintln(f, img)
	}
}

func (c *K8sCluster) DeleteNode(name string) error {
	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType:
		err := c.DeleteNodeViaApi(name)
		if err != nil {
			return err
		}
		_, err = c.WaitNodeDelete(name)
		if err != nil {
			return err
		}
		return exec.CommandContext(context.Background(), "k3d", "node", "delete", name).Run()
	case KindK8sType, AwsK8sType, AzureK8sType:
		return errors.New("deleting new node not available for " + string(Config.K8sType))
	default:
		return errors.New("unknown kubernetes type")
	}
}

func (c *K8sCluster) DeleteNodeViaApi(node string) error {
	clientset, err := k8s.GetKubernetesClientFromOptionsContextE(c.t, context.Background(), c.GetKubectlOptions())
	if err != nil {
		return errors.Wrapf(err, "error in getting access to K8S")
	}

	foreground := metav1.DeletePropagationForeground
	return clientset.CoreV1().Nodes().Delete(context.Background(), node, metav1.DeleteOptions{PropagationPolicy: &foreground})
}

func (c *K8sCluster) KillAppPod(app, namespace string) error {
	pod, err := PodNameOfApp(c, app, namespace)
	if err != nil {
		return err
	}

	if err := k8s.RunKubectlContextE(c.GetTesting(), context.Background(), c.GetKubectlOptions(namespace), "delete", "pod", pod); err != nil {
		return err
	}

	return c.WaitApp(app, namespace, 1)
}

// K8sVersionCompare compares the cluster's version with another version
func (c *K8sCluster) K8sVersionCompare(otherVersion string, baseMessage string) (int, string) {
	version, err := c.GetK8sVersion()
	if err != nil {
		c.t.Fatal(err)
	}
	return version.Compare(semver.MustParse(otherVersion)), fmt.Sprintf("%s with k8s version %s", baseMessage, version)
}

func (c *K8sCluster) ZoneName() string {
	if c.opts.zoneName != "" {
		return c.opts.zoneName
	}
	return c.Name()
}

func (c *K8sCluster) GetOrCreateAdminTunnel(args portforward.Spec) (envoy_admin.Tunnel, error) {
	args = args.WithDefaults(portforward.EnvoyAdminDefaultSpec)

	if err := args.ValidateFullSpec(); err != nil {
		return nil, errors.Wrap(err, "invalid port-forward spec")
	}

	c.mutex.RLock()
	tnl := c.adminTunnels[args]
	c.mutex.RUnlock()
	if tnl != nil {
		return tnl, nil
	}

	fwd, err := c.PortForwardApp(args)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to start port-forward to %s in namespace %q (port %d)",
			args.AppName,
			args.Namespace,
			args.RemotePort,
		)
	}

	tnl, err = tunnel.NewK8sEnvoyAdminTunnel(c.t, fwd.Endpoint)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create admin tunnel to %s in namespace %q (port %d)",
			args.AppName,
			args.Namespace,
			args.RemotePort,
		)
	}

	c.mutex.Lock()
	if existing := c.adminTunnels[args]; existing != nil {
		c.mutex.Unlock()
		return existing, nil
	}
	c.adminTunnels[args] = tnl
	c.mutex.Unlock()

	return tnl, nil
}

type appInstallation struct {
	Name      string
	Namespace string
	Replicas  int
	Outcome   error
}
