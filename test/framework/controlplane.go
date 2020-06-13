package framework

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ghodss/yaml"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/core"
	util_net "github.com/Kong/kuma/pkg/util/net"
)

type ControlPlane interface {
	AddLocalCP(name, url string) error
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
}

type PortFwd struct {
	lowFwdPort     uint32
	hiFwdPort      uint32
	localCPAPIPort uint32
	localCPGUIPort uint32
}

type K8sControlPlane struct {
	t          testing.TestingT
	mode       core.CpMode
	name       string
	kubeconfig string
	kumactl    *KumactlOptions
	cluster    *K8sCluster
	portFwd    PortFwd
	verbose    bool
}

func NewK8sControlPlane(t testing.TestingT, mode core.CpMode, name string,
	kubeconfig string, cluster *K8sCluster,
	loPort, hiPort uint32,
	verbose bool) *K8sControlPlane {
	kumactl, _ := NewKumactlOptions(t, name, verbose)
	return &K8sControlPlane{
		t:          t,
		mode:       mode,
		name:       name,
		kubeconfig: kubeconfig,
		kumactl:    kumactl,
		cluster:    cluster,
		portFwd: PortFwd{
			localCPAPIPort: loPort,
			localCPGUIPort: hiPort,
		},
		verbose: verbose,
	}
}

func (c *K8sControlPlane) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	options := &k8s.KubectlOptions{
		ConfigPath: c.kubeconfig,
	}
	for _, ns := range namespace {
		options.Namespace = ns
		break
	}

	return options
}

func (c *K8sControlPlane) AddLocalCP(name, url string) error {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t,
		c.GetKubectlOptions())
	if err != nil {
		return err
	}

	kumaCM, err := clientset.CoreV1().ConfigMaps("kuma-system").Get(context.TODO(), "kuma-control-plane-config", metav1.GetOptions{})
	if err != nil {
		return err
	}

	cfg := kuma_cp.Config{}
	err = yaml.Unmarshal([]byte(kumaCM.Data["config.yaml"]), cfg)
	if err != nil {
		return err
	}

	cfg.GlobalCP.LocalCPs[name] = url

	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	kumaCM.Data["config.yaml"] = string(yamlBytes)

	_, err = clientset.CoreV1().ConfigMaps("kube-system").Update(context.TODO(), kumaCM, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c *K8sControlPlane) PortForwardKumaCP() (uint32, uint32) {
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

	c.cluster.PortForwardPod(kumaNamespace, kumacpPodName, localAPIPort, kumaCPAPIPort)

	//find a free local port
	localGUIPort, err := util_net.PickTCPPort("", localAPIPort+1, c.portFwd.hiFwdPort)
	if err != nil {
		fmt.Println("No free port found in range: ", kumaCPAPIPortFwdLow, " - ", kumaCPAPIPortFwdHi)
		return 0, 0
	}

	c.cluster.PortForwardPod(kumaNamespace, kumacpPodName, localGUIPort, kumaCPGUIPort)

	return localAPIPort, localGUIPort
}

func (c *K8sControlPlane) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c.t,
		c.GetKubectlOptions(kumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + kumaServiceName,
		},
	)
}

func (c *K8sControlPlane) VerifyKumaCtl() error {
	output, err := c.kumactl.RunKumactlAndGetOutputV(c.verbose, "get", "dataplanes")
	fmt.Println(output)

	return err
}

func (c *K8sControlPlane) VerifyKumaREST() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.portFwd.localCPAPIPort), 10),
		&tls.Config{},
		defaultRetries,
		defaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sControlPlane) VerifyKumaGUI() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		"http://localhost:"+strconv.FormatUint(uint64(c.portFwd.localCPGUIPort), 10),
		&tls.Config{},
		3,
		defaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sControlPlane) GetKumaCPLogs() (string, error) {
	logs := ""

	pods := c.GetKumaCPPods()
	if len(pods) < 1 {
		return "", errors.Errorf("no kuma-cp pods found for logs")
	}

	for _, p := range pods {
		log, err := c.cluster.GetPodLogs(p)
		if err != nil {
			return "", err
		}

		logs = logs + "\n >>> " + p.Name + "\n" + log
	}

	return logs, nil
}

func (c *K8sControlPlane) FinalizeAdd(apiPort uint32, guiPort uint32) error {
	if c.mode == core.Local {
		return nil
	}

	c.portFwd.localCPAPIPort = apiPort
	c.portFwd.localCPGUIPort = guiPort

	kumacpURL := "http://localhost:" + strconv.FormatUint(uint64(c.portFwd.localCPAPIPort), 10)

	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, kumacpURL)
}

func (c *K8sControlPlane) InstallCP() (string, error) {
	return c.kumactl.KumactlInstallCP(c.mode)
}

func (c *K8sControlPlane) InjectDNS() error {
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
