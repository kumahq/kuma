package tracing

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sort"
)

const tracingNamespace = "kuma-tracing"

type k8SDeployment struct {
	port uint32
}

func (t *k8SDeployment) ZipkinCollectorURL() string {
	return "http://jaeger-collector.kuma-tracing:9411/api/v2/spans"
}

func (t *k8SDeployment) TracedServices() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/services", t.port))
	if err != nil {
		return nil, err
	}
	output := &jaegerServicesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	sort.Strings(output.Data)
	return output.Data, nil
}

var _ Deployment = &k8SDeployment{}

func (t *k8SDeployment) Name() string {
	return DeploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	kumactl, _ := framework.NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), true)
	yaml, err := kumactl.KumactlInstallTracing()
	if err != nil {
		return err
	}
	err = k8s.KubectlApplyFromStringE(cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		yaml)
	if err != nil {
		return err
	}

	k8s.WaitUntilNumPodsCreated(cluster.GetTesting(),
		cluster.GetKubectlOptions(tracingNamespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
		1,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	pods := k8s.ListPods(cluster.GetTesting(),
		cluster.GetKubectlOptions(tracingNamespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
	)
	if len(pods) != 1 {
		return errors.Errorf("counting Jaeger pods. Got: %d. Expected: 1", len(pods))
	}

	k8s.WaitUntilPodAvailable(cluster.GetTesting(),
		cluster.GetKubectlOptions(tracingNamespace),
		pods[0].Name,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	port, err := test.FindFreePort("")
	if err != nil {
		return err
	}
	t.port = port

	cluster.(*framework.K8sCluster).PortForwardPod(tracingNamespace, pods[0].Name, port, 16686)
	return nil
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	kumactl, _ := framework.NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), true)
	yaml, err := kumactl.KumactlInstallTracing()
	if err != nil {
		return err
	}

	err = k8s.KubectlDeleteFromStringE(cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		yaml)

	cluster.(*framework.K8sCluster).WaitNamespaceDelete(tracingNamespace)
	return nil
}
