package framework

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	util_net "github.com/Kong/kuma/pkg/util/net"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sClusters struct {
	t        testing.T
	clusters map[string]*K8sCluster
	verbose  bool
}

type K8sCluster struct {
	testing.T
	kubeconfig          string
	kumactl             *KumactlOptions
	forwardedPortsChans map[uint32]chan struct{}
	localCPPort         uint32
	verbose             bool
}

// NewK8sTest gets the number of the clusters to use in the tests, and the pattern
// to locate the KUBECONFIG for them. The second argument can be empty
func NewK8sClusters(clusterNames []string, kubeConfigPathPattern string, verbose bool) (Clusters, error) {
	if len(clusterNames) < 1 || len(clusterNames) > maxClusters {
		return nil, fmt.Errorf("Invalid cluster number. Should be in the range [1,3], but it is %d", len(clusterNames))
	}
	if kubeConfigPathPattern == "" {
		kubeConfigPathPattern = defaultKubeConfigPathPattern
	}

	clusters := map[string]*K8sCluster{}
	for _, name := range clusterNames {
		options, err := NewKumactlOptions(name, "", "", verbose)
		if err != nil {
			return nil, err
		}
		clusters[name] = &K8sCluster{
			kubeconfig:          os.ExpandEnv(fmt.Sprintf(kubeConfigPathPattern, name)),
			kumactl:             options,
			verbose:             verbose,
			forwardedPortsChans: map[uint32]chan struct{}{},
		}
	}

	return &K8sClusters{
		t:        testing.T{},
		clusters: clusters,
		verbose:  verbose,
	}, nil
}

func (cs *K8sClusters) GetCluster(name string) Cluster {
	c, found := cs.clusters[name]
	if !found {
		return nil
	}
	return c
}

func (cs *K8sClusters) DeployKuma() error {
	for name, c := range cs.clusters {
		if err := c.DeployKuma(); err != nil {
			return fmt.Errorf("Deploy Kuma on %s failed: %v", name, err)
		}
	}
	return nil
}

func (cs *K8sClusters) VerifyKuma() error {
	for name, c := range cs.clusters {
		if err := c.VerifyKuma(); err != nil {
			return fmt.Errorf("Verify Kuma on %s failed: %v", name, err)
		}
	}
	return nil
}

func (cs *K8sClusters) GetKumaCPLogs() (string, error) {
	logs := ""
	for name, c := range cs.clusters {
		log, err := c.GetKumaCPLogs()
		if err != nil {
			return "", fmt.Errorf("Verify Kuma on %s failed: %v", name, err)
		}
		logs = logs + "========== " + name + " ==========\n" + log + "\n"
	}
	return logs, nil
}

func (cs *K8sClusters) DeleteKuma() error {
	for name, c := range cs.clusters {
		if err := c.DeleteKuma(); err != nil {
			return fmt.Errorf("Delete Kuma on %s failed: %v", name, err)
		}
	}
	return nil
}

func (c *K8sCluster) Apply(namespace string, yaml string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(c, options, c.kubeconfig)
	k8s.KubectlApply(c, options, yaml)
	return nil
}

func (c *K8sCluster) ApplyAndWaitServiceOnK8sCluster(namespace string, yaml string, service string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(c, options, c.kubeconfig)
	k8s.KubectlApply(c, options, yaml)
	k8s.WaitUntilServiceAvailable(c, options, service, defaultRetries, defaultTiemout)
	return nil
}

func (c *K8sCluster) DeployKuma() error {
	yaml, err := c.kumactl.KumactlInstallCP()
	if err != nil {
		return err
	}

	err = k8s.KubectlApplyFromStringE(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
		},
		yaml)
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailable(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
			Namespace:  kumaNamespace,
		},
		kumaServiceName, defaultRetries, defaultTiemout)

	k8s.WaitUntilNumPodsCreated(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
			Namespace:  kumaNamespace,
		},
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
		1, defaultRetries, defaultTiemout)

	kumacp_pods := c.GetKumaCPPods()
	Expect(len(kumacp_pods)).To(Equal(1))

	k8s.WaitUntilPodAvailable(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
			Namespace:  kumaNamespace,
		},
		kumacp_pods[0].Name,
		10, 3*time.Second)

	c.localCPPort = c.PortForwardKumaCP()

	return nil
}

func (c *K8sCluster) DeleteKuma() error {
	c.CleanupPortForwards()

	yaml, err := c.kumactl.KumactlInstallCP()
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteFromStringE(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
		},
		yaml)

	c.WaitKumaNamespaceDelete()

	return err
}

func (c *K8sCluster) DeleteKumaNamespace() error {
	return k8s.DeleteNamespaceE(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
		}, kumaNamespace)
}

func (c *K8sCluster) IsKumaNamespaceAcitve() bool {
	_, err := k8s.GetNamespaceE(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
		}, kumaNamespace)
	return err == nil
}

func (c *K8sCluster) WaitKumaNamespaceDelete() {
	fmt.Println("Waiting for the Kuma Namespace to be deleted.")
	for c.IsKumaNamespaceAcitve() {
		fmt.Println("Kuma Namespace still active, retrying.")
		time.Sleep(defaultTiemout)
	}
}

func (c *K8sCluster) PortForwardKumaCP() uint32 {
	pods := c.GetKumaCPPods()
	if len(pods) < 1 {
		fmt.Println("No kuma-cp pods found for port-forward.")
		return 0
	}
	name := pods[0].Name

	//find free local port
	localPort, err := util_net.PickTCPPort("127.0.0.1", kumaCPAPIPort, kumaCPAPIPort)
	if err != nil {
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
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c,
		"http://localhost:"+strconv.FormatUint(uint64(c.localCPPort), 10),
		&tls.Config{},
		defaultRetries/2,
		defaultTiemout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sCluster) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c,
		&k8s.KubectlOptions{
			ConfigPath: c.kubeconfig,
			Namespace:  kumaNamespace,
		},
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
	)
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
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c, &k8s.KubectlOptions{
		ConfigPath: c.kubeconfig,
	})
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

func IsK8sClustersStarted() bool {
	_, found := os.LookupEnv(envK8SCLUSTERS)
	return found
}
