package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Kong/kuma/pkg/config/core"

	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	k8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	util_net "github.com/Kong/kuma/pkg/util/net"
)

type PortFwd struct {
	lowFwdPort          uint32
	hiFwdPort           uint32
	localCPAPIPort      uint32
	localCPGUIPort      uint32
	forwardedPortsChans map[uint32]chan struct{}
}

type K8sCluster struct {
	t          testing.TestingT
	name       string
	kubeconfig string
	kumactl    *KumactlOptions
	portFwd    PortFwd
	verbose    bool
	clientset  *kubernetes.Clientset
}

func (c *K8sCluster) Apply(namespace string, yamlPath string) error {
	options := c.GetKubectlOptions(namespace)

	return k8s.KubectlApplyE(c.t,
		options,
		yamlPath)
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

func (c *K8sCluster) PortForwardKumaCP() (uint32, uint32) {
	kumacpPods := c.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		fmt.Printf("Kuma CP pods: %d", len(kumacpPods))
		return 0, 0
	}

	kumacpPodName := kumacpPods[0].Name

	//find a free local port
	localAPIPort, err := util_net.PickTCPPort("", c.portFwd.lowFwdPort, c.portFwd.hiFwdPort)
	if err != nil {
		fmt.Println("No free port found in range: ", kumaCPAPIPortFwdLow, " - ", kumaCPAPIPortFwdHi)
		return 0, 0
	}

	c.PortForwardPod(kumaNamespace, kumacpPodName, localAPIPort, kumaCPAPIPort)

	//find a free local port
	localGUIPort, err := util_net.PickTCPPort("", localAPIPort+1, c.portFwd.hiFwdPort)
	if err != nil {
		fmt.Println("No free port found in range: ", kumaCPAPIPortFwdLow, " - ", kumaCPAPIPortFwdHi)
		return 0, 0
	}

	c.PortForwardPod(kumaNamespace, kumacpPodName, localGUIPort, kumaCPGUIPort)

	return localAPIPort, localGUIPort
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

	c.portFwd.forwardedPortsChans[localPort] = stopChan
}

func (c *K8sCluster) CleanupPortForwards() {
	for _, stop := range c.portFwd.forwardedPortsChans {
		close(stop)
	}

	c.portFwd.forwardedPortsChans = map[uint32]chan struct{}{}
}

func (c *K8sCluster) VerifyKumaCtl() error {
	output, err := c.kumactl.RunKumactlAndGetOutputV(c.verbose, "get", "dataplanes")
	fmt.Println(output)

	return err
}

func (c *K8sCluster) VerifyKumaREST() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.portFwd.localCPAPIPort), 10),
		&tls.Config{},
		DefaultRetries,
		DefaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sCluster) VerifyKumaGUI() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.portFwd.localCPGUIPort), 10),
		&tls.Config{},
		3,
		DefaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sCluster) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
	)
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

func (c *K8sCluster) DeployKuma(mode ...string) error {
	yaml, err := c.kumactl.KumactlInstallCP(mode...)
	if err != nil {
		return err
	}

	err = k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
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

	kumacpPods := c.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		return errors.Wrapf(err, "Kuma CP pods: %d", len(kumacpPods))
	}

	k8s.WaitUntilPodAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		kumacpPods[0].Name,
		DefaultRetries,
		DefaultTimeout)

	c.portFwd.localCPAPIPort, c.portFwd.localCPGUIPort = c.PortForwardKumaCP()

	// Figure out a better way to handle Local
	if len(mode) > 0 && mode[0] == core.Local {
		return nil
	}

	kumacpURL := "http://localhost:" + strconv.FormatUint(uint64(c.portFwd.localCPAPIPort), 10)

	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
}

func (c *K8sCluster) VerifyKuma() error {
	if err := c.VerifyKumaGUI(); err != nil {
		return err
	}

	if err := c.VerifyKumaREST(); err != nil {
		return err
	}

	if err := c.VerifyKumaCtl(); err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) DeleteKuma() error {
	c.CleanupPortForwards()

	yaml, err := c.kumactl.KumactlInstallCP()
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)

	c.WaitNamespaceDelete(kumaNamespace)

	return err
}

func (c *K8sCluster) GetKumactlOptions() *KumactlOptions {
	return c.kumactl
}

func (c *K8sCluster) GetKumaCPLogs() (string, error) {
	logs := ""

	pods := c.GetKumaCPPods()
	if len(pods) < 1 {
		return "", errors.Errorf("no kuma-cp pods found for logs")
	}

	for _, p := range pods {
		log, err := c.GetPodLogs(p)
		if err != nil {
			return "", err
		}

		logs = logs + "\n >>> " + p.Name + "\n" + log
	}

	return logs, nil
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

func (c *K8sCluster) LabelNamespaceForSidecarInjection(namespace string) error {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t, c.GetKubectlOptions())
	if err != nil {
		return err
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"kuma.io/sidecar-injection": "enabled",
			},
		},
	}
	_, err = clientset.CoreV1().Namespaces().Update(context.Background(), ns, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	return err
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
	// store the kumactl environment
	oldEnv := c.kumactl.Env
	c.kumactl.Env["KUBECONFIG"] = c.GetKubectlOptions().ConfigPath

	yaml, err := c.kumactl.RunKumactlAndGetOutput("install", "dns")
	if err != nil {
		return err
	}

	// restore kumactl environment
	c.kumactl.Env = oldEnv

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

func (c *K8sCluster) GetTesting() testing.TestingT {
	return c.t
}
