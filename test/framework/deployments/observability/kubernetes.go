package observability

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	deploymentName  string
	namespace       string
	jaegerApiTunnel *k8s.Tunnel
	components      []string
}

var _ Deployment = &k8SDeployment{}

func (t *k8SDeployment) ZipkinCollectorURL() string {
	return fmt.Sprintf("http://jaeger-collector.%s:9411/api/v2/spans", t.namespace)
}

func (t *k8SDeployment) TracedServices() ([]string, error) {
	return tracedServices(fmt.Sprintf("http://%s", t.jaegerApiTunnel.Endpoint()))
}

func (t *k8SDeployment) Name() string {
	return t.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	kumactl := framework.NewKumactlOptions(cluster.GetTesting(), cluster.GetKuma().GetName(), true)
	yaml, err := kumactl.KumactlInstallObservability(t.namespace, t.components)
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
		cluster.GetKubectlOptions(t.namespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
		1,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	pods := k8s.ListPods(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.namespace),
		metav1.ListOptions{
			LabelSelector: "app=jaeger",
		},
	)
	if len(pods) != 1 {
		return errors.Errorf("counting Jaeger pods. Got: %d. Expected: 1", len(pods))
	}

	err = framework.WaitUntilPodAvailableE(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.namespace),
		pods[0].Name,
		framework.DefaultRetries,
		framework.DefaultTimeout)
	if err != nil {
		return err
	}

	t.jaegerApiTunnel = k8s.NewTunnel(cluster.GetKubectlOptions(t.namespace), k8s.ResourceTypePod, pods[0].Name, 0, 16686)
	t.jaegerApiTunnel.ForwardPort(cluster.GetTesting())
	return nil
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	return cluster.DeleteNamespace(t.namespace)
}
