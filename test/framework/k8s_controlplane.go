package framework

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework/kumactl"
)

var _ ControlPlane = &K8sControlPlane{}

type K8sControlPlane struct {
	t          testing.TestingT
	mode       core.CpMode
	name       string
	kubeconfig string
	kumactl    *kumactl.KumactlOptions
	cluster    *K8sCluster
	portFwd    PortFwd
	verbose    bool
	replicas   int
	apiHeaders []string
}

func NewK8sControlPlane(
	t testing.TestingT,
	mode core.CpMode,
	clusterName string,
	kubeconfig string,
	cluster *K8sCluster,
	verbose bool,
	replicas int,
	apiHeaders []string,
) *K8sControlPlane {
	name := clusterName + "-" + mode
	return &K8sControlPlane{
		t:          t,
		mode:       mode,
		name:       name,
		kubeconfig: kubeconfig,
		kumactl:    NewKumactlOptionsE2E(t, name, verbose),
		cluster:    cluster,
		verbose:    verbose,
		replicas:   replicas,
		apiHeaders: apiHeaders,
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
	kumaCpSvc := c.GetKumaCPSvc()
	if k8s.IsServiceAvailable(&kumaCpSvc) {
		c.portFwd.apiServerTunnel = k8s.NewTunnel(c.GetKubectlOptions(Config.KumaNamespace), k8s.ResourceTypeService, kumaCpSvc.Name, 0, 5681)
		c.portFwd.apiServerTunnel.ForwardPort(c.t)
		c.portFwd.ApiServerEndpoint = c.portFwd.apiServerTunnel.Endpoint()
	}

	c.t.Fatalf("Failed finding an available service, service: %v", kumaCpSvc)
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

func (c *K8sControlPlane) GetKumaCPSvc() v1.Service {
	return k8s.ListServices(c.t,
		c.GetKubectlOptions(Config.KumaNamespace),
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + Config.KumaServiceName,
		},
	)[0]
}

func (c *K8sControlPlane) GetKumaCPSyncSvc() v1.Service {
	return k8s.ListServices(c.t,
		c.GetKubectlOptions(Config.KumaNamespace),
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + Config.KumaGlobalZoneSyncServiceName,
		},
	)[0]
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
	headers := map[string]string{}
	for _, header := range c.apiHeaders {
		res := strings.Split(header, "=")
		headers[res[0]] = res[1]
	}
	_, err := http_helper.HTTPDoWithRetryE(
		c.t,
		"GET",
		c.GetGlobalStatusAPI(),
		nil,
		headers,
		http.StatusOK,
		DefaultRetries,
		DefaultTimeout,
		&tls.Config{MinVersion: tls.VersionTLS12},
	)
	return err
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
	if !c.cluster.opts.setupKumactl {
		return nil
	}

	var token string
	t, err := c.retrieveAdminToken()
	if err != nil {
		return err
	}
	token = t
	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, c.GetAPIServerAddress(), token, c.apiHeaders)
}

func (c *K8sControlPlane) retrieveAdminToken() (string, error) {
	if authnType, exist := c.cluster.opts.env["KUMA_API_SERVER_AUTHN_TYPE"]; exist && authnType != "tokens" {
		return "", nil
	}
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
	return c.kumactl.KumactlInstallCP(args...)
}

func (c *K8sControlPlane) GetKDSInsecureServerAddress() string {
	svc := c.GetKumaCPSyncSvc()
	return "grpc://" + c.getKumaCPAddress(svc, "global-zone-sync")
}

func (c *K8sControlPlane) GetKDSServerAddress() string {
	svc := c.GetKumaCPSyncSvc()
	return "grpcs://" + c.getKumaCPAddress(svc, "global-zone-sync")
}

func (c *K8sControlPlane) GetXDSServerAddress() string {
	svc := c.GetKumaCPSvc()
	return c.getKumaCPAddress(svc, "dp-server")
}

// A naive implementation to find the Host & Port where a Service is exposing a
// CP port.
func (c *K8sControlPlane) getKumaCPAddress(svc v1.Service, portName string) string {
	var svcPort v1.ServicePort
	for _, port := range svc.Spec.Ports {
		if port.Name == portName {
			svcPort = port
		}
	}

	var address string
	var portNumber int32

	// As EKS and AWS generally returns dns records of load balancers instead of
	// IP addresses, accessing this data (hostname) was only tested there,
	// so the env var was created for that purpose
	if Config.UseLoadBalancer {
		address = svc.Status.LoadBalancer.Ingress[0].IP

		if Config.UseHostnameInsteadOfIP {
			address = svc.Status.LoadBalancer.Ingress[0].Hostname
		}

		portNumber = svcPort.Port
	} else {
		pod := c.GetKumaCPPods()[0]
		address = pod.Status.HostIP

		portNumber = svcPort.NodePort
	}

	return net.JoinHostPort(
		address, strconv.FormatUint(uint64(portNumber), 10),
	)
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

func (c *K8sControlPlane) GenerateZoneToken(zone string, scope []string) (string, error) {
	scopeJson, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}

	data := fmt.Sprintf(`'{"zone": "%s", "scope": %s}'`, zone, scopeJson)

	return c.generateToken("/zone", data)
}
