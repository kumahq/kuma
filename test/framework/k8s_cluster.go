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
	"github.com/onsi/gomega"
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
		"Wait the Kuma Namespace to terminate.",
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
	argsMap := map[string]string{
		"--control-plane-image":  kumaCPImage,
		"--dataplane-image":      kumaDPImage,
		"--dataplane-init-image": kumaInitImage,
	}
	switch mode {
	case core.Remote:
		argsMap["--kds-global-address"] = opts.globalAddress
	}
	for opt, value := range opts.ctlOpts {
		argsMap[opt] = value
	}

	var args []string
	for k, v := range argsMap {
		args = append(args, k, v)
	}
	yaml, err := c.controlplane.InstallCP(args...)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

// deployKumaViaHelm uses Helm to install kuma
// using the kuma helm chart
func (c *K8sCluster) deployKumaViaHelm(mode string, opts *deployOptions) error {
	// run from test/e2e
	helmChartPath, err := filepath.Abs(helmChartPath)
	if err != nil {
		return err
	}

	values := map[string]string{
		"controlPlane.mode":              mode,
		"global.image.tag":               kuma_version.Build.Version,
		"global.image.registry":          kumaImageRegistry,
		"controlPlane.image.repository":  kumaCPImageRepo,
		"dataPlane.image.repository":     kumaDPImageRepo,
		"dataPlane.initImage.repository": kumaInitImageRepo,
	}
	for opt, value := range opts.helmOpts {
		values[opt] = value
	}

	switch mode {
	case core.Global:
		values["controlPlane.globalRemoteSyncService.type"] = "NodePort"
	case core.Remote:
		values["controlPlane.zone"] = c.GetKumactlOptions().CPName
		values["controlPlane.kdsGlobalAddress"] = opts.globalAddress
	}

	helmOpts := &helm.Options{
		SetValues:      values,
		KubectlOptions: c.GetKubectlOptions(kumaNamespace),
	}

	releaseName := opts.helmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
	}

	// first create the namespace
	if err := k8s.CreateNamespaceE(c.t, c.GetKubectlOptions(), kumaNamespace); err != nil {
		return err
	}

	return helm.InstallE(c.t, helmOpts, helmChartPath, releaseName)
}

func (c *K8sCluster) DeployKuma(mode string, fs ...DeployOptionsFunc) error {
	c.controlplane = NewK8sControlPlane(c.t, mode, c.name, c.kubeconfig,
		c, c.loPort, c.hiPort, c.verbose)

	opts := newDeployOpt(fs...)
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

	k8s.WaitUntilNumPodsCreated(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
		1,
		DefaultRetries,
		DefaultTimeout)

	kumacpPods := c.controlplane.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		return errors.Errorf("Kuma CP pods: %d", len(kumacpPods))
	}

	k8s.WaitUntilPodAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		kumacpPods[0].Name,
		DefaultRetries,
		DefaultTimeout)

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

	err = c.controlplane.FinalizeAdd()
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) GetKuma() ControlPlane {
	return c.controlplane
}

func (c *K8sCluster) RestartKuma() error {
	c.CleanupPortForwards()

	kumacpPods := c.controlplane.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		return errors.Errorf("Kuma CP pods: %d", len(kumacpPods))
	}
	oldPod := kumacpPods[0]

	// creates the clientset
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t, c.GetKubectlOptions())
	if err != nil {
		return errors.Wrapf(err, "error in getting access to K8S")
	}

	// delete the pod
	err = clientset.CoreV1().Pods(oldPod.Namespace).Delete(context.TODO(), oldPod.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// wait for pod to terminate
	retry.DoWithRetry(c.t,
		"Wait the Kuma CP pod to terminate.",
		DefaultRetries,
		DefaultTimeout,
		func() (string, error) {
			_, err := k8s.GetPodE(c.t,
				c.GetKubectlOptions(oldPod.Namespace),
				oldPod.Name)
			if err != nil {
				return "Pod " + oldPod.Name + " deleted", nil
			}
			return "Pod available " + oldPod.Name, fmt.Errorf("Pod %s still active", oldPod.Name)
		})

	k8s.WaitUntilNumPodsCreated(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
		1,
		DefaultRetries,
		DefaultTimeout)

	kumacpPods = c.controlplane.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		return errors.Errorf("Kuma CP pods: %d", len(kumacpPods))
	}

	newPod := kumacpPods[0]
	gomega.Expect(oldPod.Name).ToNot(gomega.Equal(newPod.Name))

	k8s.WaitUntilPodAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		newPod.Name,
		DefaultRetries,
		DefaultTimeout)

	err = c.controlplane.FinalizeAdd()
	if err != nil {
		return err
	}

	return nil
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
		KubectlOptions: c.GetKubectlOptions(kumaNamespace),
	}

	if err := helm.DeleteE(c.t, helmOpts, opts.helmReleaseName, true); err != nil {
		errs = multierr.Append(errs, err)
	}

	if err := c.DeleteNamespace(kumaNamespace); err != nil {
		errs = multierr.Append(errs, err)
	}

	return errs
}

func (c *K8sCluster) deleteKumaViaKumactl(opts *deployOptions) error {
	argsMap := map[string]string{}
	switch c.controlplane.mode {
	case core.Remote:
		// kumactl remote deployment will fail if GlobalAddress is not specified
		argsMap["--kds-global-address"] = "grpcs://0.0.0.0:5685"
	}
	for opt, value := range opts.ctlOpts {
		argsMap[opt] = value
	}

	var args []string
	for k, v := range argsMap {
		args = append(args, k, v)
	}
	yaml, err := c.controlplane.InstallCP(args...)
	if err != nil {
		return err
	}

	_ = k8s.KubectlDeleteFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)

	c.WaitNamespaceDelete(kumaNamespace)

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

func (c *K8sCluster) DeployApp(namespace, appname string) error {
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

func (c *K8sCluster) InjectDNS() error {
	return c.controlplane.InjectDNS()
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
