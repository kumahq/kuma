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
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
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
	opts                kumaDeploymentOptions
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
		defaultRetries:      Config.DefaultClusterStartupRetries,
		defaultTimeout:      Config.DefaultClusterStartupTimeout,
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

// deployKumaViaKubectl uses kubectl to install kuma
// using the resources from the `kumactl install control-plane` command
func (c *K8sCluster) deployKumaViaKubectl(mode string) error {
	yaml, err := c.yamlForKumaViaKubectl(mode)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

func (c *K8sCluster) yamlForKumaViaKubectl(mode string) (string, error) {
	argsMap := map[string]string{
		"--namespace":                 Config.KumaNamespace,
		"--control-plane-repository":  Config.KumaCPImageRepo,
		"--dataplane-repository":      Config.KumaDPImageRepo,
		"--dataplane-init-repository": Config.KumaInitImageRepo,
	}
	if Config.KumaImageRegistry != "" {
		argsMap["--control-plane-registry"] = Config.KumaImageRegistry
		argsMap["--dataplane-registry"] = Config.KumaImageRegistry
		argsMap["--dataplane-init-registry"] = Config.KumaImageRegistry
	}

	if Config.KumaImageTag != "" {
		argsMap["--control-plane-version"] = Config.KumaImageTag
		argsMap["--dataplane-version"] = Config.KumaImageTag
		argsMap["--dataplane-init-version"] = Config.KumaImageTag
	}

	switch mode {
	case core.Zone:
		argsMap["--kds-global-address"] = c.opts.globalAddress
	}

	if c.opts.ingress {
		argsMap["--ingress-enabled"] = ""
		argsMap["--ingress-use-node-port"] = ""
	}

	if c.opts.egress {
		argsMap["--egress-enabled"] = ""
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

	if Config.CIDR != "" {
		argsMap["--env-var"] = fmt.Sprintf("KUMA_DNS_SERVER_CIDR=%s", Config.CIDR)
	}

	for opt, value := range c.opts.ctlOpts {
		argsMap[opt] = value
	}

	var args []string
	for k, v := range argsMap {
		args = append(args, k, v)
	}

	for k, v := range c.opts.env {
		args = append(args, "--env-var", fmt.Sprintf("%s=%s", k, v))
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

	for opt, value := range c.opts.helmOpts {
		values[opt] = value
	}

	if Config.XDSApiVersion != "" {
		values["controlPlane.envVars.KUMA_BOOTSTRAP_SERVER_API_VERSION"] = Config.XDSApiVersion
	}

	if c.opts.cni {
		values["cni.enabled"] = "true"
		values["cni.chained"] = "true"
		values["cni.netDir"] = Config.CNIConf.NetDir
		values["cni.binDir"] = Config.CNIConf.BinDir
		values["cni.confName"] = Config.CNIConf.ConfName
	}

	if Config.CIDR != "" {
		values["controlPlane.envVars.KUMA_DNS_SERVER_CIDR"] = Config.CIDR
	}

	switch mode {
	case core.Global:
		if !Config.UseLoadBalancer {
			values["controlPlane.globalZoneSyncService.type"] = "NodePort"
		}
	case core.Zone:
		values["controlPlane.zone"] = c.GetKumactlOptions().CPName
		values["controlPlane.kdsGlobalAddress"] = c.opts.globalAddress
	}

	for _, value := range c.opts.noHelmOpts {
		delete(values, value)
	}

	prefixedValues := map[string]string{}
	for k, v := range values {
		prefixedValues[Config.HelmSubChartPrefix+k] = v
	}

	return prefixedValues
}

type helmFn func(testing.TestingT, *helm.Options, string, string) error

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

	if c.opts.helmChartVersion != "" {
		helmOpts.Version = c.opts.helmChartVersion
	}

	releaseName := c.opts.helmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
	}

	// create the namespace if it does not exist
	if _, err = k8s.GetNamespaceE(c.t, c.GetKubectlOptions(), Config.KumaNamespace); err != nil {
		if err = k8s.CreateNamespaceE(c.t, c.GetKubectlOptions(), Config.KumaNamespace); err != nil {
			return err
		}
	}

	return fn(c.t, helmOpts, helmChart, releaseName)
}

// deployKumaViaHelm uses Helm to install kuma
// using the kuma helm chart
func (c *K8sCluster) deployKumaViaHelm(mode string) error {
	return c.processViaHelm(mode, helm.InstallE)
}

// upgradeKumaViaHelm uses Helm to upgrade kuma
// using the kuma helm chart
func (c *K8sCluster) upgradeKumaViaHelm(mode string) error {
	return c.processViaHelm(mode, helm.UpgradeE)
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

	// backwards compatibility, check for 1.3.x localhost is admin env variable.
	localhostIsAdmin := c.opts.env["KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST"] == "true"
	c.controlplane = NewK8sControlPlane(c.t, mode, c.name, c.kubeconfig, c, c.verbose, replicas, localhostIsAdmin)

	switch mode {
	case core.Zone:
		if c.opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for zone")
		}
	}

	var err error
	switch c.opts.installationMode {
	case KumactlInstallationMode:
		err = c.deployKumaViaKubectl(mode)
	case HelmInstallationMode:
		err = c.deployKumaViaHelm(mode)
	default:
		return errors.Errorf("invalid installation mode: %s", c.opts.installationMode)
	}

	if err != nil {
		return err
	}

	err = c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, replicas)
	if err != nil {
		return err
	}

	if c.opts.cni {
		err = c.WaitApp(Config.CNIApp, Config.CNINamespace, 1)
		if err != nil {
			return err
		}
	}

	if !c.opts.skipDefaultMesh {
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
	for name, updateFuncs := range c.opts.meshUpdateFuncs {
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

	return c.VerifyKuma()
}

func (c *K8sCluster) UpgradeKuma(mode string, opt ...KumaDeploymentOption) error {
	if c.controlplane == nil {
		return errors.New("To upgrade Kuma has to be installed first")
	}

	c.opts.apply(opt...)
	if c.opts.cpReplicas != 0 {
		c.controlplane.replicas = c.opts.cpReplicas
	}

	switch mode {
	case core.Zone:
		if c.opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for zone")
		}
	}

	if err := c.upgradeKumaViaHelm(mode); err != nil {
		return err
	}

	if err := c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, c.controlplane.replicas); err != nil {
		return err
	}

	if c.opts.cni {
		if err := c.WaitApp(Config.CNIApp, Config.CNINamespace, 1); err != nil {
			return err
		}
	}

	if !c.opts.skipDefaultMesh {
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
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(Config.KumaNamespace), "scale", "--replicas=0", fmt.Sprintf("deployment/%s", Config.KumaServiceName)); err != nil {
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
	if err := k8s.RunKubectlE(c.GetTesting(), c.GetKubectlOptions(Config.KumaNamespace), "scale", fmt.Sprintf("--replicas=%d", c.controlplane.replicas), fmt.Sprintf("deployment/%s", Config.KumaServiceName)); err != nil {
		return err
	}
	if err := c.WaitApp(Config.KumaServiceName, Config.KumaNamespace, c.controlplane.replicas); err != nil {
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

func (c *K8sCluster) deleteKumaViaHelm() (errs error) {
	if c.opts.helmReleaseName == "" {
		return errors.New("must supply a helm release name for cleanup")
	}

	helmOpts := &helm.Options{
		KubectlOptions: c.GetKubectlOptions(Config.KumaNamespace),
	}

	if err := helm.DeleteE(c.t, helmOpts, c.opts.helmReleaseName, true); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err := c.DeleteNamespace(Config.KumaNamespace); err != nil {
		errs = multierr.Append(errs, err)
	}

	// HELM does not remove CRDs therefore we need to do it manually.
	// It's important to remove CRDs to get rid of all "instances" of CRDs like default Mesh etc.
	if err := c.deleteCRDs(); err != nil {
		errs = multierr.Append(errs, err)
	}

	return errs
}

func (c *K8sCluster) deleteKumaViaKumactl() error {
	yaml, err := c.yamlForKumaViaKubectl(c.controlplane.mode)
	if err != nil {
		return err
	}

	_ = k8s.KubectlDeleteFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)

	c.WaitNamespaceDelete(Config.KumaNamespace)

	return nil
}

func (c *K8sCluster) DeleteKuma() error {
	if c.controlplane.portFwd.localAPITunnel != nil {
		c.controlplane.portFwd.localAPITunnel.Close()
	}
	var err error
	switch c.opts.installationMode {
	case HelmInstallationMode:
		err = c.deleteKumaViaHelm()
	case KumactlInstallationMode:
		err = c.deleteKumaViaKumactl()
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
