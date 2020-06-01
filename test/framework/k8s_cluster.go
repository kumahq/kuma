package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	util_net "github.com/Kong/kuma/pkg/util/net"
)

type K8sCluster struct {
	t                   testing.TestingT
	name                string
	kubeconfig          string
	kumactl             *KumactlOptions
	forwardedPortsChans map[uint32]chan struct{}
	localCPPort         uint32
	verbose             bool
}

func (c *K8sCluster) Apply(namespace string, yaml string) error {
	options := c.GetKubectlOptions(namespace)
	defer k8s.KubectlDelete(c.t, options, c.kubeconfig)
	k8s.KubectlApply(c.t, options, yaml)
	return nil
}

func (c *K8sCluster) ApplyAndWaitServiceOnK8sCluster(namespace string, yaml string, service string) error {
	options := c.GetKubectlOptions(namespace)
	defer k8s.KubectlDelete(c.t, options, c.kubeconfig)
	k8s.KubectlApply(c.t, options, yaml)
	k8s.WaitUntilServiceAvailable(c.t, options, service, defaultRetries, defaultTiemout)
	return nil
}

func (c *K8sCluster) DeployKuma() error {
	yaml, err := c.kumactl.KumactlInstallCP()
	if err != nil {
		return err
	}

	err = k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		kumaServiceName, defaultRetries, defaultTiemout)

	k8s.WaitUntilNumPodsCreated(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
		1, defaultRetries, defaultTiemout)

	kumacp_pods := c.GetKumaCPPods()
	Expect(len(kumacp_pods)).To(Equal(1))

	k8s.WaitUntilPodAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		kumacp_pods[0].Name,
		10, 3*time.Second)

	c.localCPPort = c.PortForwardKumaCP()

	err = c.kumactl.KumactlConfigControlPlanesAdd(c.name, "http://localhost:"+strconv.FormatUint(uint64(c.localCPPort), 10))
	Expect(err).ToNot(HaveOccurred())

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

	c.WaitKumaNamespaceDelete()

	return err
}

func (c *K8sCluster) DeleteKumaNamespace() error {
	return k8s.DeleteNamespaceE(c.t,
		c.GetKubectlOptions(),
		kumaNamespace)
}

func (c *K8sCluster) WaitKumaNamespaceDelete() {
	retry.DoWithRetry(c.t, "Wait the Kuma Namespace to terminate.", 2*defaultRetries, defaultTiemout,
		func() (string, error) {
			_, err := k8s.GetNamespaceE(c.t,
				c.GetKubectlOptions(),
				kumaNamespace)
			if err != nil {
				return "", nil
			}
			return "", fmt.Errorf("Kuma Namespace still not terminated.")
		})
}

func (c *K8sCluster) PortForwardKumaCP() uint32 {
	pods := c.GetKumaCPPods()
	if len(pods) < 1 {
		fmt.Println("No kuma-cp pods found for port-forward.")
		return 0
	}
	name := pods[0].Name

	//find free local port
	localPort, err := util_net.PickTCPPort("127.0.0.1", 32000+kumaCPAPIPort, 42000+kumaCPAPIPort)
	if err != nil {
		fmt.Println("No free port found in range: ", 32000+kumaCPAPIPort, " - ", 42000+kumaCPAPIPort)
		return 0
	}

	c.PortForwardPod(kumaNamespace, name, localPort, kumaCPAPIPort)
	return localPort
}

func (c *K8sCluster) PortForwardPod(namespace, name string, localPort, remotePort uint32) {
	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		fmt.Println("Error building config")
		return
	}
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, name)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	ports := []string{strconv.FormatUint(uint64(localPort), 10) + ":" + strconv.FormatUint(uint64(remotePort), 10)}

	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		panic(err)
	}

	go func() {
		for range readyChan { // Kubernetes will close this channel when it has something to tell us.
		}
		if len(errOut.String()) != 0 {
			panic(errOut.String())
		} else if len(out.String()) != 0 {
			fmt.Println(out.String())
		}
	}()

	go func() {
		if err = forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
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

func (c *K8sCluster) VerifyKuma() error {
	if err := c.VerifyKumaCtl(); err != nil {
		return err
	}
	if err := c.VerifyKumaREST(); err != nil {
		return err
	}
	return nil
}

func (c *K8sCluster) VerifyKumaCtl() error {
	output, err := c.kumactl.RunKumactlAndGetOutputV(Verbose, "get", "dataplanes")
	fmt.Println(output)
	return err
}

func (c *K8sCluster) VerifyKumaREST() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.localCPPort), 10),
		&tls.Config{},
		defaultRetries/2,
		defaultTiemout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
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
	return err
}

func (c *K8sCluster) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
	)
}

func (c *K8sCluster) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	options := &k8s.KubectlOptions{
		ConfigPath: c.kubeconfig,
	}
	for _, ns := range namespace {
		options.Namespace = ns
	}
	return options
}

func (c *K8sCluster) GetKumaCPLogs() (string, error) {
	logs := ""
	pods := c.GetKumaCPPods()
	if len(pods) < 1 {
		return "", fmt.Errorf("No kuma-cp pods found for logs.")
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

func (c *K8sCluster) GetPodLogs(pod v1.Pod) (string, error) {
	podLogOpts := v1.PodLogOptions{}
	// creates the clientset
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t, c.GetKubectlOptions())
	if err != nil {
		return "", fmt.Errorf("error in getting access to K8S")
	}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("error in opening stream")
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error in copy information from podLogs to buf")
	}
	str := buf.String()

	return str, nil
}

func (c *K8sCluster) GetTesting() testing.TestingT {
	return c.t
}
