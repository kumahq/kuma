package framework

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kumahq/kuma/pkg/config/core"
	bootstrap_k8s "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
)

type PortFwd struct {
	localAPITunnel *k8s.Tunnel
}

type K8sControlPlane struct {
	t                testing.TestingT
	mode             core.CpMode
	name             string
	kubeconfig       string
	kumactl          *KumactlOptions
	cluster          *K8sCluster
	portFwd          PortFwd
	verbose          bool
	replicas         int
	localhostIsAdmin bool
}

func NewK8sControlPlane(
	t testing.TestingT,
	mode core.CpMode,
	clusterName string,
	kubeconfig string,
	cluster *K8sCluster,
	verbose bool,
	replicas int,
	localhostIsAdmin bool,
) *K8sControlPlane {
	name := clusterName + "-" + mode
	kumactl, _ := NewKumactlOptions(t, name, verbose)
	return &K8sControlPlane{
		t:                t,
		mode:             mode,
		name:             name,
		kubeconfig:       kubeconfig,
		kumactl:          kumactl,
		cluster:          cluster,
		verbose:          verbose,
		replicas:         replicas,
		localhostIsAdmin: localhostIsAdmin,
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
			c.portFwd.localAPITunnel = k8s.NewTunnel(c.GetKubectlOptions(KumaNamespace), k8s.ResourceTypePod, kumaCpPods[i].Name, 0, kumaCPAPIPort)
			c.portFwd.localAPITunnel.ForwardPort(c.t)
			return
		}
	}
	c.t.Fatalf("Failed finding an available pod, allPods: %v", kumaCpPods)
}

func (c *K8sControlPlane) GetKumaCPPods() []v1.Pod {
	return k8s.ListPods(c.t,
		c.GetKubectlOptions(KumaNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + KumaServiceName,
		},
	)
}

func (c *K8sControlPlane) GetKumaCPSvcs() []v1.Service {
	return k8s.ListServices(c.t,
		c.GetKubectlOptions(KumaNamespace),
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + KumaGlobalZoneSyncServiceName,
		},
	)
}

func (c *K8sControlPlane) VerifyKumaCtl() error {
	if c.portFwd.localAPITunnel == nil {
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
		&tls.Config{},
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
		&tls.Config{},
		3,
		DefaultTimeout,
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

func (c *K8sControlPlane) FinalizeAdd() error {
	c.PortForwardKumaCP()
	var token string
	if !c.localhostIsAdmin {
		t, err := c.retrieveAdminToken()
		if err != nil {
			return err
		}
		token = t
	}
	return c.kumactl.KumactlConfigControlPlanesAdd(c.name, c.GetAPIServerAddress(), token)
}

func (c *K8sControlPlane) retrieveAdminToken() (string, error) {
	return retry.DoWithRetryE(c.t, "generating DP token", DefaultRetries, DefaultTimeout, func() (string, error) {
		sec, err := k8s.GetSecretE(c.t, c.GetKubectlOptions(KumaNamespace), "admin-user-token")
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

func (c *K8sControlPlane) InjectDNS(args ...string) error {
	// store the kumactl environment
	oldEnv := c.kumactl.Env
	c.kumactl.Env["KUBECONFIG"] = c.GetKubectlOptions().ConfigPath
	defer func() {
		c.kumactl.Env = oldEnv // restore kumactl environment
	}()

	yaml, err := c.kumactl.KumactlInstallDNS(args...)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(c.t,
		c.GetKubectlOptions(),
		yaml)
}

// A naive implementation to find the URL where Zone CP exposes its API
func (c *K8sControlPlane) GetKDSServerAddress() string {
	// As EKS and AWS generally returns dns records of load balancers instead of
	//  IP addresses, accessing this data (hostname) was only tested there,
	//  so the env var was created for that purpose
	if UseLoadBalancer() {
		svc := c.GetKumaCPSvcs()[0]

		address := svc.Status.LoadBalancer.Ingress[0].IP

		if UseHostnameInsteadOfIP() {
			address = svc.Status.LoadBalancer.Ingress[0].Hostname
		}

		return "grpcs://" + address + ":" + strconv.FormatUint(loadBalancerKdsPort, 10)
	}

	pod := c.GetKumaCPPods()[0]
	return "grpcs://" + net.JoinHostPort(
		pod.Status.HostIP, strconv.FormatUint(uint64(kdsPort), 10))
}

func (c *K8sControlPlane) GetAPIServerAddress() string {
	if c.portFwd.localAPITunnel == nil {
		panic("Port forward wasn't setup!")
	}
	return "http://" + c.portFwd.localAPITunnel.Endpoint()
}

func (c *K8sControlPlane) GetMetrics() (string, error) {
	panic("not implemented")
}

func (c *K8sControlPlane) GetGlobalStatusAPI() string {
	return c.GetAPIServerAddress() + "/status/zones"
}

func (c *K8sControlPlane) GenerateDpToken(mesh, service string) (string, error) {
	var token string
	if !c.localhostIsAdmin {
		t, err := c.retrieveAdminToken()
		if err != nil {
			return "", err
		}
		token = t
	}
	dpType := ""
	if service == "ingress" {
		dpType = "ingress"
	}
	return http_helper.HTTPDoWithRetryE(
		c.t,
		"POST",
		c.GetAPIServerAddress()+"/tokens",
		[]byte(fmt.Sprintf(`{"mesh": "%s", "type": "%s", "tags": {"kuma.io/service": ["%s"]}}`, mesh, dpType, service)),
		map[string]string{
			"content-type":  "application/json",
			"authorization": "Bearer " + token,
		},
		200,
		DefaultRetries,
		DefaultTimeout,
		&tls.Config{},
	)
}

func (c *K8sControlPlane) GenerateZoneIngressToken(zone string) (string, error) {
	var token string
	if !c.localhostIsAdmin {
		t, err := c.retrieveAdminToken()
		if err != nil {
			return "", err
		}
		token = t
	}
	return http_helper.HTTPDoWithRetryE(
		c.t,
		"POST",
		c.GetAPIServerAddress()+"/tokens/zone-ingress",
		[]byte(fmt.Sprintf(`{"zone": "%s"}`, zone)),
		map[string]string{
			"content-type":  "application/json",
			"authorization": "Bearer " + token,
		},
		200,
		DefaultRetries,
		DefaultTimeout,
		&tls.Config{},
	)
}

// UpdateObject fetches an object and updates it after the update function is applied to it.
func (c *K8sControlPlane) UpdateObject(
	typeName string,
	objectName string,
	update func(object runtime.Object) runtime.Object,
) error {
	scheme, err := bootstrap_k8s.NewScheme()
	if err != nil {
		return err
	}
	codecs := serializer.NewCodecFactory(scheme)
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	if !ok {
		return errors.Errorf("no serializer for %q", runtime.ContentTypeYAML)
	}

	_, err = retry.DoWithRetryableErrorsE(c.t, "update object", map[string]string{"Error from server \\(Conflict\\)": "object conflict"}, 5, time.Second, func() (string, error) {
		out, err := k8s.RunKubectlAndGetOutputE(c.t, c.GetKubectlOptions(), "get", typeName, objectName, "-o", "yaml")
		if err != nil {
			return "", err
		}

		decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(out)))
		into := map[string]interface{}{}

		if err := decoder.Decode(&into); err != nil {
			return "", err
		}

		u := unstructured.Unstructured{Object: into}
		obj, err := scheme.New(u.GroupVersionKind())
		if err != nil {
			return "", err
		}

		if err := scheme.Convert(&u, obj, nil); err != nil {
			return "", err
		}

		obj = update(obj)
		encoder := codecs.EncoderForVersion(info.Serializer, obj.GetObjectKind().GroupVersionKind().GroupVersion())
		yml, err := runtime.Encode(encoder, obj)
		if err != nil {
			return "", err
		}

		if err := k8s.KubectlApplyFromStringE(c.t, c.GetKubectlOptions(), string(yml)); err != nil {
			return "", err
		}
		return "", nil
	})

	if err != nil {
		return errors.Wrapf(err, "failed to update %s object %q", typeName, objectName)
	}
	return nil
}
