package observability

import (
	"fmt"
	"slices"

	"github.com/gruntwork-io/terratest/modules/k8s"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	deploymentName  string
	namespace       string
	jaegerApiTunnel *k8s.Tunnel
	components      []Component
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
	kumactl := framework.NewKumactlOptionsE2E(cluster.GetTesting(), cluster.Name(), true)
	var strComponents []string
	for _, component := range t.components {
		strComponents = append(strComponents, string(component))
	}
	yaml, err := kumactl.KumactlInstallObservability(t.namespace, strComponents)
	if err != nil {
		return err
	}
	err = k8s.KubectlApplyFromStringE(cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		yaml)
	if err != nil {
		return err
	}

	expectedPods := map[Component]int{
		JaegerComponent:     1,
		GrafanaComponent:    1,
		LokiComponent:       1,
		PrometheusComponent: 2,
	}

	for _, component := range t.components {
		err := framework.NewClusterSetup().
			Install(framework.WaitNumPods(t.namespace, expectedPods[component], string(component))).
			Install(framework.WaitPodsAvailable(t.namespace, string(component))).
			Setup(cluster)
		if err != nil {
			return err
		}
	}

	if slices.Contains(t.components, JaegerComponent) {
		podName, err := framework.PodNameOfApp(cluster, string(JaegerComponent), t.namespace)
		if err != nil {
			return err
		}
		t.jaegerApiTunnel = k8s.NewTunnel(cluster.GetKubectlOptions(t.namespace), k8s.ResourceTypePod, podName, 0, 16686)
		t.jaegerApiTunnel.ForwardPort(cluster.GetTesting())
	}
	return nil
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.namespace)
}
