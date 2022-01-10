package framework

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

type K8sCluster struct {
	t                   testing.TestingT
	name                string
	kubeconfig          string
	controlplane        *K8sControlPlane
	forwardedPortsChans []chan struct{}
	verbose             bool
	deployments         map[string]Deployment
	defaultTimeout      time.Duration
	defaultRetries      int
}

var _ Cluster = &K8sCluster{}

func NewK8sCluster(t *TestingT, clusterName string, verbose bool) *K8sCluster {
	return &K8sCluster{
		t:                   t,
		name:                clusterName,
		kubeconfig:          os.ExpandEnv(fmt.Sprintf(defaultKubeConfigPathPattern, clusterName)),
		forwardedPortsChans: []chan struct{}{},
		verbose:             verbose,
		deployments:         map[string]Deployment{},
		defaultRetries:      GetDefaultRetries(),
		defaultTimeout:      GetDefaultTimeout(),
	}
}

func (c *K8sCluster) WithTimeout(timeout time.Duration) Cluster {
	c.defaultTimeout = timeout

	return c
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

func (c *K8sCluster) Apply(namespace string, yamlPath string) error {
	options := c.GetKubectlOptions(namespace)

	return k8s.KubectlApplyE(c.t,
		options,
		yamlPath)
}

func (c *K8sCluster) Deployment(name string) Deployment {
	return c.deployments[name]
}

func (c *K8sCluster) ApplyAndWaitServiceOnK8sCluster(namespace string, service string, yamlPath string) error {
	options := c.GetKubectlOptions(namespace)

	err := k8s.KubectlApplyE(c.t,
		options,
		yamlPath)
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailable(c.t,
		options,
		service,
		c.defaultRetries,
		c.defaultTimeout)

	return nil
}
func (c *K8sCluster) WaitNamespaceCreate(namespace string) {
	retry.DoWithRetry(c.t,
		"Wait the Kuma Namespace to terminate.",
		c.defaultRetries,
		c.defaultTimeout,
		func() (string, error) {
			_, err := k8s.GetNamespaceE(c.t,
				c.GetKubectlOptions(),
				namespace)
			if err != nil {
				return "Namespace not available " + namespace, fmt.Errorf("Namespace %s still active", namespace)
			}

			return "Namespace " + namespace + " created", nil
		})
}

func (c *K8sCluster) WaitNamespaceDelete(namespace string) {
	retry.DoWithRetry(c.t,
		fmt.Sprintf("Wait for %s Namespace to terminate.", namespace),
		2*c.defaultRetries,
		2*c.defaultTimeout,
		func() (string, error) {
			_, err := k8s.GetNamespaceE(c.t,
				c.GetKubectlOptions(),
				namespace)
			if err != nil {
				return "Namespace " + namespace + " deleted", nil
			}
			return "Namespace available " + namespace, fmt.Errorf("Namespace %s still active", namespace)
		})
}

func (c *K8sCluster) GetPodLogs(pod v1.Pod) (string, error) {
	podLogOpts := v1.PodLogOptions{}
	// creates the clientset
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t, c.GetKubectlOptions())
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

type cniConf struct {
	binDir   string
	netDir   string
	confName string
}

func (c *K8sCluster) getCniConfig() (cfg cniConf) {
	nodes := k8s.GetNodes(c.t, c.GetKubectlOptions())
	Expect(nodes).ToNot(BeEmpty())
	// Check if it's k3d or kind
	if strings.HasPrefix(nodes[0].Spec.ProviderID, "k3s://") {
		cfg.confName = "10-flannel.conflist"
		cfg.netDir = "/var/lib/rancher/k3s/agent/etc/cni/net.d"
		cfg.binDir = "/bin"
	} else {
		cfg.confName = "10-kindnet.conflist"
		cfg.netDir = "/etc/cni/net.d"
		cfg.binDir = "/opt/cni/bin"
	}
	cName := GetCniConfName()
	if cName != "" {
		cfg.confName = cName
	}
	return cfg
}

// deployKumaViaKubectl uses kubectl to install kuma
// using the resources from the `kumactl install control-plane` command
func (c *K8sCluster) deployKumaViaKubectl(mode string, opts *kumaDeploymentOptions) error {
	yaml, err := c.yamlForKumaViaKubectl(mode, opts)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

func (c *K8sCluster) yamlForKumaViaKubectl(mode string, opts *kumaDeploymentOptions) (string, error) {
	argsMap := map[string]string{
		"--namespace":                 KumaNamespace,
		"--control-plane-registry":    KumaImageRegistry,
		"--control-plane-repository":  KumaCPImageRepo,
		"--dataplane-registry":        KumaImageRegistry,
		"--dataplane-repository":      KumaDPImageRepo,
		"--dataplane-init-registry":   KumaImageRegistry,
		"--dataplane-init-repository": KumaInitImageRepo,
	}

	if HasGlobalImageRegistry() {
		argsMap["--control-plane-registry"] = GetGlobalImageRegistry()
		argsMap["--dataplane-registry"] = GetGlobalImageRegistry()
		argsMap["--dataplane-init-registry"] = GetGlobalImageRegistry()
	}

	if HasGlobalImageTag() {
		argsMap["--control-plane-version"] = GetGlobalImageTag()
		argsMap["--dataplane-version"] = GetGlobalImageTag()
		argsMap["--dataplane-init-version"] = GetGlobalImageTag()
	}

	switch mode {
	case core.Zone:
		argsMap["--kds-global-address"] = opts.globalAddress
	}

	if opts.ingress {
		argsMap["--ingress-enabled"] = ""
		argsMap["--ingress-use-node-port"] = ""
	}

	if opts.cni {
		cniCfg := c.getCniConfig()
		argsMap["--cni-enabled"] = ""
		argsMap["--cni-chained"] = ""
		argsMap["--cni-net-dir"] = cniCfg.netDir
		argsMap["--cni-bin-dir"] = cniCfg.binDir
		argsMap["--cni-conf-name"] = cniCfg.confName
	}

	if HasApiVersion() {
		argsMap["--env-var"] = "KUMA_BOOTSTRAP_SERVER_API_VERSION=" + GetApiVersion()
	}

	if opts.isipv6 {
		argsMap["--env-var"] = fmt.Sprintf("KUMA_DNS_SERVER_CIDR=%s", cidrIPv6)
	}

	for opt, value := range opts.ctlOpts {
		argsMap[opt] = value
	}

	var args []string
	for k, v := range argsMap {
		args = append(args, k, v)
	}

	for k, v := range opts.env {
		args = append(args, "--env-var", fmt.Sprintf("%s=%s", k, v))
	}

	return c.controlplane.InstallCP(args...)
}

func (c *K8sCluster) genValues(mode string, opts *kumaDeploymentOptions) map[string]string {
	values := map[string]string{
		"controlPlane.mode":                      mode,
		"global.image.tag":                       kuma_version.Build.Version,
		"global.image.registry":                  KumaImageRegistry,
		"controlPlane.image.repository":          KumaCPImageRepo,
		"dataPlane.image.repository":             KumaDPImageRepo,
		"dataPlane.initImage.repository":         KumaInitImageRepo,
		"controlPlane.defaults.skipMeshCreation": strconv.FormatBool(opts.skipDefaultMesh),
	}

	if HasGlobalImageRegistry() {
		values["global.image.registry"] = GetGlobalImageRegistry()
	}

	if HasGlobalImageTag() {
		values["global.image.tag"] = GetGlobalImageTag()
	}

	if HasCpImageRegistry() {
		values["controlPlane.image.repository"] = GetCpImageRegistry()
	}

	if HasDpImageRegistry() {
		values["dataPlane.image.repository"] = GetDpImageRegistry()
	}

	if HasDpInitImageRegistry() {
		values["dataPlane.initImage.repository"] = GetDpInitImageRegistry()
	}

	if opts.cpReplicas != 0 {
		values["controlPlane.replicas"] = strconv.Itoa(opts.cpReplicas)
	}

	for opt, value := range opts.helmOpts {
		values[opt] = value
	}

	if HasApiVersion() {
		values["controlPlane.envVars.KUMA_BOOTSTRAP_SERVER_API_VERSION"] = GetApiVersion()
	}

	if opts.cni {
		cniCfg := c.getCniConfig()
		values["cni.enabled"] = "true"
		values["cni.chained"] = "true"
		values["cni.netDir"] = cniCfg.netDir
		values["cni.binDir"] = cniCfg.binDir
		values["cni.confName"] = cniCfg.confName
	}

	if opts.isipv6 {
		values["controlPlane.envVars.KUMA_DNS_SERVER_CIDR"] = cidrIPv6
	}

	switch mode {
	case core.Global:
		if !UseLoadBalancer() {
			values["controlPlane.globalZoneSyncService.type"] = "NodePort"
		}
	case core.Zone:
		values["controlPlane.zone"] = c.GetKumactlOptions().CPName
		values["controlPlane.kdsGlobalAddress"] = opts.globalAddress
	}

	for _, value := range opts.noHelmOpts {
		delete(values, value)
	}

	prefixedValues := map[string]string{}
	for k, v := range values {
		prefixedValues[HelmSubChartPrefix+k] = v
	}

	return prefixedValues
}

type helmFn func(testing.TestingT, *helm.Options, string, string) error

func (c *K8sCluster) processViaHelm(mode string, opts *kumaDeploymentOptions, fn helmFn) error {
	// run from test/e2e
	helmChart, err := filepath.Abs(HelmChartPath)
	if err != nil {
		return err
	}

	if HasHelmChartPath() {
		helmChart = GetHelmChartPath()
	}

	if opts.helmChartPath != nil {
		helmChart = *opts.helmChartPath
	}

	values := c.genValues(mode, opts)

	helmOpts := &helm.Options{
		SetValues:      values,
		KubectlOptions: c.GetKubectlOptions(KumaNamespace),
	}

	if opts.helmChartVersion != "" {
		helmOpts.Version = opts.helmChartVersion
	}

	releaseName := opts.helmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
	}

	// create the namespace if it does not exist
	if _, err = k8s.GetNamespaceE(c.t, c.GetKubectlOptions(), KumaNamespace); err != nil {
		if err = k8s.CreateNamespaceE(c.t, c.GetKubectlOptions(), KumaNamespace); err != nil {
			return err
		}
	}

	return fn(c.t, helmOpts, helmChart, releaseName)
}

// deployKumaViaHelm uses Helm to install kuma
// using the kuma helm chart
func (c *K8sCluster) deployKumaViaHelm(mode string, opts *kumaDeploymentOptions) error {
	return c.processViaHelm(mode, opts, helm.InstallE)
}

// upgradeKumaViaHelm uses Helm to upgrade kuma
// using the kuma helm chart
func (c *K8sCluster) upgradeKumaViaHelm(mode string, opts *kumaDeploymentOptions) error {
	return c.processViaHelm(mode, opts, helm.UpgradeE)
}

func (c *K8sCluster) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
	var opts kumaDeploymentOptions

	opts.apply(opt...)

	replicas := 1
	if opts.cpReplicas != 0 {
		replicas = opts.cpReplicas
	}

	// backwards compatibility, check for 1.3.x localhost is admin env variable.
	localhostIsAdmin := opts.env["KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST"] == "true"
	c.controlplane = NewK8sControlPlane(c.t, mode, c.name, c.kubeconfig, c, c.verbose, replicas, localhostIsAdmin)

	switch mode {
	case core.Zone:
		if opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for zone")
		}
	}

	var err error
	switch opts.installationMode {
	case KumactlInstallationMode:
		err = c.deployKumaViaKubectl(mode, &opts)
	case HelmInstallationMode:
		err = c.deployKumaViaHelm(mode, &opts)
	default:
		return errors.Errorf("invalid installation mode: %s", opts.installationMode)
	}

	if err != nil {
		return err
	}

	err = c.WaitApp(KumaServiceName, KumaNamespace, replicas)
	if err != nil {
		return err
	}

	if opts.cni {
		err = c.WaitApp(CNIApp, CNINamespace, 1)
		if err != nil {
			return err
		}
	}

	if !opts.skipDefaultMesh {
		// wait for the mesh
		_, err = retry.DoWithRetryE(c.t,
			"get default mesh",
			c.defaultRetries,
			c.defaultTimeout,
			func() (s string, err error) {
				return k8s.RunKubectlAndGetOutputE(c.t, c.GetKubectlOptions(), "get", "mesh", "default")
			})
		if err != nil {
			return err
		}
	}

	if err := c.controlplane.FinalizeAdd(); err != nil {
		return err
	}

	converter := resources_k8s.NewSimpleConverter()
	for name, updateFuncs := range opts.meshUpdateFuncs {
		for _, f := range updateFuncs {
			Logf("applying update function to mesh %q", name)
			err := c.controlplane.UpdateObject("mesh", name,
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

	return nil
}

func (c *K8sCluster) UpgradeKuma(mode string, opt ...KumaDeploymentOption) error {
	var opts kumaDeploymentOptions

	if c.controlplane == nil {
		return errors.New("To upgrade Kuma has to be installed first")
	}

	opts.apply(opt...)
	if opts.cpReplicas != 0 {
		c.controlplane.replicas = opts.cpReplicas
	}

	switch mode {
	case core.Zone:
		if opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for zone")
		}
	}

	if err := c.upgradeKumaViaHelm(mode, &opts); err != nil {
		return err
	}

	if err := c.WaitApp(KumaServiceName, KumaNamespace, c.controlplane.replicas); err != nil {
		return err
	}

	if opts.cni {
		if err := c.WaitApp(CNIApp, CNINamespace, 1); err != nil {
			return err
		}
	}

	if !opts.skipDefaultMesh {
		// wait for the mesh
		_, err := retry.DoWithRetryE(c.t,
			"get default mesh",
			c.defaultRetries,
			c.defaultTimeout,
			func() (s string, err error) {
				return k8s.RunKubectlAndGetOutputE(c.t, c.GetKubectlOptions(), "get", "mesh", "default")
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

// StopControlPlane scales the replicas of a control plane to 0 and wait for it to complete. Useful for testing restarts in combination with RestartControlPlane.
func (c *K8sCluster) StopControlPlane() error {
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(KumaNamespace), "scale", "--replicas=0", fmt.Sprintf("deployment/%s", KumaServiceName)); err != nil {
		return err
	}
	_, err := retry.DoWithRetryE(c.t,
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
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(KumaNamespace), "scale", fmt.Sprintf("--replicas=%d", c.controlplane.replicas), fmt.Sprintf("deployment/%s", KumaServiceName)); err != nil {
		return err
	}
	if err := c.WaitApp(KumaServiceName, KumaNamespace, c.controlplane.replicas); err != nil {
		return err
	}

	if err := c.controlplane.FinalizeAdd(); err != nil {
		return err
	}

	return c.VerifyKuma()
}

func (c *K8sCluster) GetKuma() ControlPlane {
	return c.controlplane
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

	return nil
}

func (c *K8sCluster) deleteCRDs() (errs error) {
	var crdsYAML bytes.Buffer
	cmd := exec.Command("kubectl", "get", "crds", "-o", "yaml", "--kubeconfig", c.kubeconfig)
	cmd.Stdout = &crdsYAML

	if err := cmd.Run(); err != nil {
		errs = multierr.Append(errs, err)
	} else if tmpfile, err := os.CreateTemp("", "crds.yaml"); err != nil {
		errs = multierr.Append(errs, err)
	} else {
		defer os.Remove(tmpfile.Name()) // clean up

		if _, err := tmpfile.Write(crdsYAML.Bytes()); err != nil {
			errs = multierr.Append(errs, err)
		} else if err := tmpfile.Close(); err != nil {
			errs = multierr.Append(errs, err)
		} else if err := k8s.KubectlDeleteE(c.t, c.GetKubectlOptions(), tmpfile.Name()); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func (c *K8sCluster) deleteKumaViaHelm(opts *kumaDeploymentOptions) (errs error) {
	if opts.helmReleaseName == "" {
		return errors.New("must supply a helm release name for cleanup")
	}

	helmOpts := &helm.Options{
		KubectlOptions: c.GetKubectlOptions(KumaNamespace),
	}

	if err := helm.DeleteE(c.t, helmOpts, opts.helmReleaseName, true); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err := c.DeleteNamespace(KumaNamespace); err != nil {
		errs = multierr.Append(errs, err)
	}

	// HELM does not remove CRDs therefore we need to do it manually.
	// It's important to remove CRDs to get rid of all "instances" of CRDs like default Mesh etc.
	if err := c.deleteCRDs(); err != nil {
		errs = multierr.Append(errs, err)
	}

	return errs
}

func (c *K8sCluster) deleteKumaViaKumactl(opts *kumaDeploymentOptions) error {
	yaml, err := c.yamlForKumaViaKubectl(c.controlplane.mode, opts)
	if err != nil {
		return err
	}

	_ = k8s.KubectlDeleteFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)

	c.WaitNamespaceDelete(KumaNamespace)

	return nil
}

func (c *K8sCluster) DeleteKuma(opt ...KumaDeploymentOption) error {
	var opts kumaDeploymentOptions

	opts.apply(opt...)

	if c.controlplane.portFwd.localAPITunnel != nil {
		c.controlplane.portFwd.localAPITunnel.Close()
	}
	var err error
	switch opts.installationMode {
	case HelmInstallationMode:
		err = c.deleteKumaViaHelm(&opts)
	case KumactlInstallationMode:
		err = c.deleteKumaViaKumactl(&opts)
	}

	return err
}

func (c *K8sCluster) GetKumactlOptions() *KumactlOptions {
	return c.controlplane.kumactl
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

func (c *K8sCluster) CreateNamespace(namespace string) error {
	err := k8s.CreateNamespaceE(c.GetTesting(), c.GetKubectlOptions(), namespace)
	if err != nil {
		return err
	}

	c.WaitNamespaceCreate(namespace)

	return nil
}

func (c *K8sCluster) DeleteNamespace(namespace string) error {
	err := k8s.DeleteNamespaceE(c.GetTesting(), c.GetKubectlOptions(), namespace)
	if err != nil {
		return err
	}

	c.WaitNamespaceDelete(namespace)

	return nil
}

func (c *K8sCluster) DeployApp(opt ...AppDeploymentOption) error {
	var opts appDeploymentOptions

	opts.apply(opt...)

	namespace := opts.namespace
	appname := opts.appname

	retry.DoWithRetry(c.GetTesting(), "apply "+appname+" svc", c.defaultRetries, c.defaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyE(c.GetTesting(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+"-svc.yaml"))
			return "", err
		})

	k8s.WaitUntilServiceAvailable(c.GetTesting(),
		c.GetKubectlOptions(namespace),
		appname, c.defaultRetries, c.defaultTimeout)

	retry.DoWithRetry(c.GetTesting(), "apply "+appname, c.defaultRetries, c.defaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyE(c.GetTesting(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+".yaml"))
			return "", err
		})

	k8s.WaitUntilNumPodsCreated(c.GetTesting(),
		c.GetKubectlOptions(),
		metav1.ListOptions{
			LabelSelector: "app=" + appname,
		},
		1, c.defaultRetries, c.defaultTimeout)

	return nil
}

func (c *K8sCluster) DeleteApp(namespace, appname string) error {
	err := k8s.KubectlDeleteE(c.GetTesting(),
		c.GetKubectlOptions(namespace),
		filepath.Join("testdata", appname+"-svc.yaml"))
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteE(c.GetTesting(),
		c.GetKubectlOptions(namespace),
		filepath.Join("testdata", appname+".yaml"))
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) InjectDNS(namespace ...string) error {
	args := []string{}
	if len(namespace) > 0 {
		args = append(args, "--namespace", namespace[0])
	}

	return c.controlplane.InjectDNS(args...)
}

func (c *K8sCluster) GetTesting() testing.TestingT {
	return c.t
}

func (c *K8sCluster) DismissCluster() (errs error) {
	for name, deployment := range c.deployments {
		if err := deployment.Delete(c); err != nil {
			errs = multierr.Append(errs, err)
		}
		delete(c.deployments, name)
	}
	return nil
}

func (c *K8sCluster) Deploy(deployment Deployment) error {
	c.deployments[deployment.Name()] = deployment
	return deployment.Deploy(c)
}

func (c *K8sCluster) DeleteDeployment(name string) error {
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

func (c *K8sCluster) WaitApp(name, namespace string, replicas int) error {
	k8s.WaitUntilNumPodsCreated(c.t,
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + name,
		},
		replicas,
		c.defaultRetries,
		c.defaultTimeout)

	pods := k8s.ListPods(c.t,
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + name,
		},
	)
	if len(pods) < replicas {
		return errors.Errorf("%s pods: %d. expected %d", name, len(pods), replicas)
	}

	for i := 0; i < replicas; i++ {
		k8s.WaitUntilPodAvailable(c.t,
			c.GetKubectlOptions(namespace),
			pods[i].Name,
			c.defaultRetries,
			c.defaultTimeout)
	}
	return nil
}
