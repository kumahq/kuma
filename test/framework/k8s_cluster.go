package framework

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	kuma_version "github.com/kumahq/kuma/pkg/version"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/pkg/errors"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type K8sCluster struct {
	t                   testing.TestingT
	name                string
	kubeconfig          string
	controlplane        *K8sControlPlane
	loPort              uint32
	hiPort              uint32
	forwardedPortsChans map[uint32]chan struct{}
	verbose             bool
	clientset           *kubernetes.Clientset
	deployments         map[string]Deployment
}

func NewK8SCluster(t *TestingT, clusterName string, verbose bool) (Cluster, error) {
	cluster := &K8sCluster{
		t:                   t,
		name:                clusterName,
		kubeconfig:          os.ExpandEnv(fmt.Sprintf(defaultKubeConfigPathPattern, clusterName)),
		loPort:              uint32(kumaCPAPIPortFwdBase + 1000),
		hiPort:              uint32(kumaCPAPIPortFwdBase + 1999),
		forwardedPortsChans: map[uint32]chan struct{}{},
		verbose:             verbose,
		deployments:         map[string]Deployment{},
	}

	var err error
	cluster.clientset, err = k8s.GetKubernetesClientFromOptionsE(t, cluster.GetKubectlOptions())
	if err != nil {
		return nil, errors.Wrapf(err, "error in getting access to K8S")
	}
	return cluster, nil
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
		DefaultRetries,
		DefaultTimeout)

	return nil
}
func (c *K8sCluster) WaitNamespaceCreate(namespace string) {
	retry.DoWithRetry(c.t,
		"Wait the Kuma Namespace to terminate.",
		DefaultRetries,
		DefaultTimeout,
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
		DefaultRetries,
		DefaultTimeout,
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

func (c *K8sCluster) PortForwardPod(namespace string, podName string, localPort, remotePort uint32) {
	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		fmt.Printf("Error building config %v", err)
		return
	}

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		fmt.Printf("Error port forwarding %v", err)
		return
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{
		Scheme: "https",
		Path:   path,
		Host:   hostIP,
	}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	stdOut, stdErr := new(bytes.Buffer), new(bytes.Buffer)

	dialer := spdy.NewDialer(upgrader,
		&http.Client{
			Transport: roundTripper,
		},
		http.MethodPost, &serverURL)

	localPortStr := strconv.FormatUint(uint64(localPort), 10)
	remotePortStr := strconv.FormatUint(uint64(remotePort), 10)
	ports := []string{localPortStr + ":" + remotePortStr}

	forwarder, err := portforward.New(dialer, ports,
		stopChan, readyChan,
		stdOut, stdErr)
	if err != nil {
		panic(err)
	}

	go func() {
		// Kubernetes will close this channel when it has something to tell us.
		for range readyChan {
		}

		if len(stdErr.String()) != 0 {
			panic(stdErr.String())
		} else if len(stdOut.String()) != 0 {
			fmt.Println(stdOut.String())
		}
	}()

	go func() {
		err := forwarder.ForwardPorts()
		if err != nil {
			panic(err)
		}
	}()

	c.forwardedPortsChans[localPort] = stopChan
}

func (c *K8sCluster) CleanupPortForwards() {
	for _, stop := range c.forwardedPortsChans {
		close(stop)
	}

	c.forwardedPortsChans = map[uint32]chan struct{}{}
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
func (c *K8sCluster) deployKumaViaKubectl(mode string, opts *deployOptions) error {
	yaml, err := c.yamlForKumaViaKubectl(mode, opts)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

func (c *K8sCluster) yamlForKumaViaKubectl(mode string, opts *deployOptions) (string, error) {
	argsMap := map[string]string{
		"--namespace":               KumaNamespace,
		"--control-plane-registry":  KumaImageRegistry,
		"--dataplane-registry":      KumaImageRegistry,
		"--dataplane-init-registry": KumaImageRegistry,
	}

	globalImageRegistry := os.Getenv("KUMA_GLOBAL_IMAGE_REGISTRY")
	if globalImageRegistry != "" {
		argsMap["--control-plane-registry"] = globalImageRegistry
		argsMap["--dataplane-registry"] = globalImageRegistry
		argsMap["--dataplane-init-registry"] = globalImageRegistry
	}

	globalImageTag := os.Getenv("KUMA_GLOBAL_IMAGE_TAG")
	if globalImageTag != "" {
		argsMap["--control-plane-version"] = globalImageTag
		argsMap["--dataplane-version"] = globalImageTag
		argsMap["--dataplane-init-version"] = globalImageTag
	}

	switch mode {
	case core.Remote:
		argsMap["--kds-global-address"] = opts.globalAddress
	}

	if opts.ingress {
		argsMap["--ingress-enabled"] = ""
		argsMap["--ingress-use-node-port"] = ""
	}

	if opts.cni {
		argsMap["--cni-enabled"] = ""
		argsMap["--cni-chained"] = ""
		argsMap["--cni-net-dir"] = "/etc/cni/net.d"
		argsMap["--cni-bin-dir"] = "/opt/cni/bin"
		argsMap["--cni-conf-name"] = "10-kindnet.conflist"

		cniConfName := os.Getenv("KUMA_CNI_CONF_NAME")
		if cniConfName != "" {
			argsMap["--cni-conf-name"] = cniConfName
		}
	}

	apiVersion := os.Getenv(envAPIVersion)
	if apiVersion != "" {
		argsMap["--env-var"] = "KUMA_BOOTSTRAP_SERVER_API_VERSION=" + apiVersion
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

func genValues(mode string, opts *deployOptions, kumactlOpts *KumactlOptions) map[string]string {
	values := map[string]string{
		"controlPlane.mode":                      mode,
		"global.image.tag":                       kuma_version.Build.Version,
		"global.image.registry":                  KumaImageRegistry,
		"controlPlane.image.repository":          KumaCPImageRepo,
		"dataPlane.image.repository":             KumaDPImageRepo,
		"dataPlane.initImage.repository":         KumaInitImageRepo,
		"controlPlane.defaults.skipMeshCreation": strconv.FormatBool(opts.skipDefaultMesh),
	}

	globalImageTag := os.Getenv("KUMA_GLOBAL_IMAGE_TAG")
	if globalImageTag != "" {
		values["global.image.tag"] = globalImageTag
	}

	globalImageRegistry := os.Getenv("KUMA_GLOBAL_IMAGE_REGISTRY")
	if globalImageRegistry != "" {
		values["global.image.registry"] = globalImageRegistry
	}

	cpImageRepository := os.Getenv("KUMA_CP_IMAGE_REPOSITORY")
	if cpImageRepository != "" {
		values["controlPlane.image.repository"] = cpImageRepository
	}

	dpImageRepository := os.Getenv("KUMA_DP_IMAGE_REPOSITORY")
	if dpImageRepository != "" {
		values["dataPlane.image.repository"] = dpImageRepository
	}

	dpInitImageRepository := os.Getenv("KUMA_DP_INIT_IMAGE_REPOSITORY")
	if dpInitImageRepository != "" {
		values["dataPlane.initImage.repository"] = dpInitImageRepository
	}

	if opts.cpReplicas != 0 {
		values["controlPlane.replicas"] = strconv.Itoa(opts.cpReplicas)
	}

	for opt, value := range opts.helmOpts {
		values[opt] = value
	}

	apiVersion := os.Getenv(envAPIVersion)
	if apiVersion != "" {
		values["controlPlane.envVars.KUMA_BOOTSTRAP_SERVER_API_VERSION"] = apiVersion
	}

	if opts.cni {
		values["cni.enabled"] = "true"
		values["cni.chained"] = "true"
		values["cni.netDir"] = "/etc/cni/net.d"
		values["cni.binDir"] = "/opt/cni/bin"
		values["cni.confName"] = "10-aws.conflist"

		cniConfName := os.Getenv("KUMA_CNI_CONF_NAME")
		if cniConfName != "" {
			values["cni.confName"] = cniConfName
		}
	}

	switch mode {
	case core.Global:
		if useLoadBalancer := os.Getenv("KUMA_USE_LOAD_BALANCER"); useLoadBalancer == "" {
			values["controlPlane.globalRemoteSyncService.type"] = "NodePort"
		}
	case core.Remote:
		values["controlPlane.zone"] = kumactlOpts.CPName
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

func (c *K8sCluster) processViaHelm(mode string, opts *deployOptions, fn helmFn) error {
	// run from test/e2e
	helmChart, err := filepath.Abs(HelmChartPath)
	if err != nil {
		return err
	}

	helmChartPath := os.Getenv("HELM_CHART_PATH")
	if helmChartPath != "" {
		helmChart = helmChartPath
	}

	if opts.helmChartPath != nil {
		helmChart = *opts.helmChartPath
	}

	values := genValues(mode, opts, c.GetKumactlOptions())

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
func (c *K8sCluster) deployKumaViaHelm(mode string, opts *deployOptions) error {
	return c.processViaHelm(mode, opts, helm.InstallE)
}

// upgradeKumaViaHelm uses Helm to upgrade kuma
// using the kuma helm chart
func (c *K8sCluster) upgradeKumaViaHelm(mode string, opts *deployOptions) error {
	return c.processViaHelm(mode, opts, helm.UpgradeE)
}

func (c *K8sCluster) DeployKuma(mode string, fs ...DeployOptionsFunc) error {
	opts := newDeployOpt(fs...)
	replicas := 1
	if opts.cpReplicas != 0 {
		replicas = opts.cpReplicas
	}
	c.controlplane = NewK8sControlPlane(c.t, mode, c.name, c.kubeconfig, c, c.loPort, c.hiPort, c.verbose, replicas)

	switch mode {
	case core.Remote:
		if opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for remote")
		}
	}

	var err error
	switch opts.installationMode {
	case KumactlInstallationMode:
		err = c.deployKumaViaKubectl(mode, opts)
	case HelmInstallationMode:
		err = c.deployKumaViaHelm(mode, opts)
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
			DefaultRetries,
			DefaultTimeout,
			func() (s string, err error) {
				return k8s.RunKubectlAndGetOutputE(c.t, c.GetKubectlOptions(), "get", "mesh", "default")
			})
		if err != nil {
			return err
		}
	}

	err = c.controlplane.FinalizeAdd()
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) UpgradeKuma(mode string, fs ...DeployOptionsFunc) error {
	if c.controlplane == nil {
		return errors.New("To upgrade Kuma has to be installed first")
	}

	opts := newDeployOpt(fs...)
	if opts.cpReplicas != 0 {
		c.controlplane.replicas = opts.cpReplicas
	}

	switch mode {
	case core.Remote:
		if opts.globalAddress == "" {
			return errors.Errorf("GlobalAddress expected for remote")
		}
	}

	if err := c.upgradeKumaViaHelm(mode, opts); err != nil {
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
			DefaultRetries,
			DefaultTimeout,
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

func (c *K8sCluster) deleteKumaViaHelm(opts *deployOptions) (errs error) {
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
	crdsYAML, err := k8s.RunKubectlAndGetOutputE(c.t, c.GetKubectlOptions(), "get", "crds", "-o", "yaml")
	if err != nil {
		errs = multierr.Append(errs, err)
	} else if err := k8s.KubectlDeleteFromStringE(c.t, c.GetKubectlOptions(), crdsYAML); err != nil {
		errs = multierr.Append(errs, err)
	}

	return errs
}

func (c *K8sCluster) deleteKumaViaKumactl(opts *deployOptions) error {
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

func (c *K8sCluster) DeleteKuma(fs ...DeployOptionsFunc) error {
	c.CleanupPortForwards()

	opts := newDeployOpt(fs...)

	var err error
	switch opts.installationMode {
	case HelmInstallationMode:
		err = c.deleteKumaViaHelm(opts)
	case KumactlInstallationMode:
		err = c.deleteKumaViaKumactl(opts)
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

func (c *K8sCluster) DeployApp(fs ...DeployOptionsFunc) error {
	opts := newDeployOpt(fs...)
	namespace := opts.namespace
	appname := opts.appname

	retry.DoWithRetry(c.GetTesting(), "apply "+appname+" svc", DefaultRetries, DefaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyE(c.GetTesting(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+"-svc.yaml"))
			return "", err
		})

	k8s.WaitUntilServiceAvailable(c.GetTesting(),
		c.GetKubectlOptions(namespace),
		appname, DefaultRetries, DefaultTimeout)

	retry.DoWithRetry(c.GetTesting(), "apply "+appname, DefaultRetries, DefaultTimeout,
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
		1, DefaultRetries, DefaultTimeout)

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
	for _, deployment := range c.deployments {
		if err := deployment.Delete(c); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return nil
}

func (c *K8sCluster) Deploy(deployment Deployment) error {
	c.deployments[deployment.Name()] = deployment
	return deployment.Deploy(c)
}

func (c *K8sCluster) WaitApp(name, namespace string, replicas int) error {
	k8s.WaitUntilNumPodsCreated(c.t,
		c.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: "app=" + name,
		},
		replicas,
		DefaultRetries,
		DefaultTimeout)

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
			DefaultRetries,
			DefaultTimeout)
	}
	return nil
}
