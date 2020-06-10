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

	"github.com/pkg/errors"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
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
		defaultRetries,
		defaultTimeout)

	return nil
}
func (c *K8sCluster) WaitNamespaceCreate(namespace string) {
	retry.DoWithRetry(c.t,
		"Wait the Kuma Namespace to terminate.",
		defaultRetries,
		defaultTimeout,
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
		defaultRetries,
		defaultTimeout,
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

func (c *K8sCluster) PortForwardKumaCP() uint32 {
	kumacpPods := c.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		fmt.Printf("Kuma CP pods: %d", len(kumacpPods))
		return 0
	}

	kumacpPodName := kumacpPods[0].Name

	//find a free local port
	localPort, err := util_net.PickTCPPort("127.0.0.1", kumaCPAPIPortFwdLow, kumaCPAPIPortFwdHi)
	if err != nil {
		fmt.Println("No free port found in range: ", kumaCPAPIPortFwdLow, " - ", kumaCPAPIPortFwdHi)
		return 0
	}

	c.PortForwardPod(kumaNamespace, kumacpPodName, localPort, kumaCPAPIPort)

	return localPort
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

func (c *K8sCluster) VerifyKumaCtl() error {
	output, err := c.kumactl.RunKumactlAndGetOutputV(c.verbose, "get", "dataplanes")
	fmt.Println(output)

	return err
}

func (c *K8sCluster) VerifyKumaREST() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.localCPPort), 10),
		&tls.Config{},
		defaultRetries,
		defaultTimeout,
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
		defaultRetries,
		defaultTimeout)

	kumacpPods := c.GetKumaCPPods()
	if len(kumacpPods) != 1 {
		return errors.Wrapf(err, "Kuma CP pods: %d", len(kumacpPods))
	}

	k8s.WaitUntilPodAvailable(c.t,
		c.GetKubectlOptions(kumaNamespace),
		kumacpPods[0].Name,
		defaultRetries,
		defaultTimeout)

	c.localCPPort = c.PortForwardKumaCP()
	kumacpURL := "http://localhost:" + strconv.FormatUint(uint64(c.localCPPort), 10)

	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
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
	retry.DoWithRetry(c.GetTesting(), "apply "+appname+" svc", defaultRetries, defaultTimeout,
		func() (string, error) {
			err := k8s.KubectlApplyE(c.GetTesting(),
				c.GetKubectlOptions(namespace),
				filepath.Join("testdata", appname+"-svc.yaml"))
			return "", err
		})

	k8s.WaitUntilServiceAvailable(c.GetTesting(),
		c.GetKubectlOptions(namespace),
		appname, defaultRetries, defaultTimeout)

	retry.DoWithRetry(c.GetTesting(), "apply "+appname, defaultRetries, defaultTimeout,
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
		1, defaultRetries, defaultTimeout)

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
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t,
		c.GetKubectlOptions())
	if err != nil {
		return err
	}

	kumaCPSVC, err := k8s.GetServiceE(c.t,
		c.GetKubectlOptions("kuma-system"),
		"kuma-control-plane")
	if err != nil {
		return err
	}

	cpaddress := kumaCPSVC.Spec.ClusterIP

	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if strings.Contains(corednsConfigMap.Data["Corefile"], "kuma:53") {
		return nil
	}

	toappend := fmt.Sprintf(`kuma:53 {
        errors
        cache 30
        forward . %s:5653
    }`, cpaddress)

	corednsConfigMap.Data["Corefile"] += toappend

	_, err = clientset.CoreV1().ConfigMaps("kube-system").Update(context.TODO(), corednsConfigMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sCluster) GetTesting() testing.TestingT {
	return c.t
}
