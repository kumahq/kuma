package framework

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
)

type K8sControlPlane struct {
	t          testing.TestingT
	mode       core.CpMode
	name       string
	kubeconfig string
	kumactl    *KumactlOptions
	cluster    *K8sCluster
	portFwd    PortFwd
	verbose    bool
	replicas   int
}

func NewK8sControlPlane(
	t testing.TestingT,
	mode core.CpMode,
	clusterName string,
	kubeconfig string,
	cluster *K8sCluster,
	verbose bool,
	replicas int,
) *K8sControlPlane {
	name := clusterName + "-" + mode
	return &K8sControlPlane{
		t:          t,
		mode:       mode,
		name:       name,
		kubeconfig: kubeconfig,
		kumactl:    NewKumactlOptions(t, name, verbose),
		cluster:    cluster,
		verbose:    verbose,
		replicas:   replicas,
	}
}

func (c *K8sControlPlane) GetName() string {
	return c.name
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

func (c *K8sControlPlane) PortForwardKumaCP() {
	kumaCpPods := c.GetKumaCPPods()
	// There could be multiple pods still starting so pick one that's available already
	for i := range kumaCpPods {
		if k8s.IsPodAvailable(&kumaCpPods[i]) {
			c.portFwd.apiServerTunnel = k8s.NewTunnel(c.GetKubectlOptions(Config.KumaNamespace), k8s.ResourceTypePod, kumaCpPods[i].Name, 0, 5681)
			c.portFwd.apiServerTunnel.ForwardPort(c.t)
			c.portFwd.ApiServerEndpoint = c.portFwd.apiServerTunnel.Endpoint()
			return
		}
	}
	c.t.Fatalf("Failed finding an available pod, allPods: %v", kumaCpPods)
}

func (c *K8sControlPlane) ClosePortForwards() {
	if c.portFwd.apiServerTunnel != nil {
		c.portFwd.apiServerTunnel.Close()
	}
}

func (c *K8sControlPlane) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c.t,
		c.GetKubectlOptions(Config.KumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + Config.KumaServiceName,
		},
	)
}

func (c *K8sControlPlane) GetKumaCPSvcs() []v1.Service {
	return k8s.ListServices(c.t,
		c.GetKubectlOptions(Config.KumaNamespace),
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + Config.KumaGlobalZoneSyncServiceName,
		},
	)
}

func (c *K8sControlPlane) VerifyKumaCtl() error {
	if c.portFwd.ApiServerEndpoint == "" {
		return errors.Errorf("API port not forwarded")
	}

	output, err := c.kumactl.RunKumactlAndGetOutputV(c.verbose, "get", "dataplanes")
	fmt.Println(output)

	return err
}

func (c *K8sControlPlane) VerifyKumaREST() error {
	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		c.GetGlobalStatusAPI(),
		&tls.Config{MinVersion: tls.VersionTLS12},
		DefaultRetries,
		DefaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sControlPlane) VerifyKumaGUI() error {
	if c.mode == core.Zone {
		return nil
	}

	return http_helper.HttpGetWithRetryWithCustomValidationE(
		c.t,
		c.GetAPIServerAddress()+"/gui",
		&tls.Config{MinVersion: tls.VersionTLS12},
		3,
		DefaultTimeout,
		func(statusCode int, body string) bool {
			return statusCode == http.StatusOK
		},
	)
}

func (c *K8sControlPlane) PortFwd() PortFwd {
	return c.portFwd
}

func (c *K8sControlPlane) FinalizeAdd() error {
	c.PortForwardKumaCP()
	return c.FinalizeAddWithPortFwd(c.portFwd)
}

func (c *K8sControlPlane) FinalizeAddWithPortFwd(portFwd PortFwd) error {
	c.portFwd = portFwd
	var token string
	t, err := c.retrieveAdminToken()
	if err != nil {
		return err
	}
	token = t
	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, c.GetAPIServerAddress(), token)
}

func (c *K8sControlPlane) retrieveAdminToken() (string, error) {
	if c.cluster.opts.helmOpts["controlPlane.environment"] == "universal" {
		body, err := http_helper.HTTPDoWithRetryWithOptionsE(c.t, http_helper.HttpDoOptions{
			Method:    "GET",
			Url:       c.GetAPIServerAddress() + "/global-secrets/admin-user-token",
			TlsConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			Body:      bytes.NewReader([]byte{}),
		}, http.StatusOK, DefaultRetries, DefaultTimeout)
		if err != nil {
			return "", err
		}
		return ExtractSecretDataFromResponse(body)
	}

	return retry.DoWithRetryE(c.t, "generating DP token", DefaultRetries, DefaultTimeout, func() (string, error) {
		sec, err := k8s.GetSecretE(c.t, c.GetKubectlOptions(Config.KumaNamespace), "admin-user-token")
		if err != nil {
			return "", err
		}
		return string(sec.Data["value"]), nil
	})
}

func (c *K8sControlPlane) InstallCP(args ...string) (string, error) {
	// store the kumactl environment
	oldEnv := c.kumactl.Env
	c.kumactl.Env["KUBECONFIG"] = c.GetKubectlOptions().ConfigPath
	defer func() {
		c.kumactl.Env = oldEnv // restore kumactl environment
	}()
	return c.kumactl.KumactlInstallCP(c.mode, args...)
}

func (c *K8sControlPlane) GetKDSInsecureServerAddress() string {
	return c.getKDSServerAddress(false)
}

func (c *K8sControlPlane) GetKDSServerAddress() string {
	return c.getKDSServerAddress(true)
}

// A naive implementation to find the URL where Zone CP exposes its API
func (c *K8sControlPlane) getKDSServerAddress(secure bool) string {
	protocol := "grpcs"
	if !secure {
		protocol = "grpc"
	}

	// As EKS and AWS generally returns dns records of load balancers instead of
	//  IP addresses, accessing this data (hostname) was only tested there,
	//  so the env var was created for that purpose
	if Config.UseLoadBalancer {
		svc := c.GetKumaCPSvcs()[0]

		address := svc.Status.LoadBalancer.Ingress[0].IP

		if Config.UseHostnameInsteadOfIP {
			address = svc.Status.LoadBalancer.Ingress[0].Hostname
		}

		return protocol + "://" + address + ":" + strconv.FormatUint(loadBalancerKdsPort, 10)
	}

	pod := c.GetKumaCPPods()[0]
	return protocol + "://" + net.JoinHostPort(
		pod.Status.HostIP, strconv.FormatUint(uint64(kdsPort), 10))
}

func (c *K8sControlPlane) GetAPIServerAddress() string {
	if c.portFwd.ApiServerEndpoint == "" {
		panic("Port forward wasn't setup!")
	}
	return "http://" + c.portFwd.ApiServerEndpoint
}

func (c *K8sControlPlane) GetMetrics() (string, error) {
	panic("not implemented")
}

func (c *K8sControlPlane) Exec(cmd ...string) (string, string, error) {
	pod := c.GetKumaCPPods()[0]
	return c.cluster.Exec(pod.Namespace, pod.Name, "", cmd...)
}

func (c *K8sControlPlane) GetGlobalStatusAPI() string {
	return c.GetAPIServerAddress() + "/status/zones"
}

func (c *K8sControlPlane) generateToken(
	tokenPath string,
	data string,
) (string, error) {
	token, err := c.retrieveAdminToken()
	if err != nil {
		return "", err
	}

	return http_helper.HTTPDoWithRetryE(
		c.t,
		"POST",
		c.GetAPIServerAddress()+"/tokens"+tokenPath,
		[]byte(data),
		map[string]string{
			"content-type":  "application/json",
			"authorization": "Bearer " + token,
		},
		200,
		DefaultRetries,
		DefaultTimeout,
		&tls.Config{MinVersion: tls.VersionTLS12},
	)
}

func (c *K8sControlPlane) GenerateDpToken(mesh, service string) (string, error) {
	var dpType string
	if service == "ingress" {
		dpType = "ingress"
	}

	data := fmt.Sprintf(
		`{"mesh": "%s", "type": "%s", "tags": {"kuma.io/service": ["%s"]}}`,
		mesh,
		dpType,
		service,
	)

	return c.generateToken("", data)
}

func (c *K8sControlPlane) GenerateZoneIngressToken(zone string) (string, error) {
	data := fmt.Sprintf(`{"zone": "%s", "scope": ["ingress"]}`, zone)

	return c.generateToken("/zone", data)
}

func (c *K8sControlPlane) GenerateZoneIngressLegacyToken(zone string) (string, error) {
	data := fmt.Sprintf(`{"zone": "%s"}`, zone)

	return c.generateToken("/zone-ingress", data)
}

func (c *K8sControlPlane) GenerateZoneEgressToken(zone string) (string, error) {
	data := fmt.Sprintf(`{"zone": "%s", "scope": ["egress"]}`, zone)

	return c.generateToken("/zone", data)
}
