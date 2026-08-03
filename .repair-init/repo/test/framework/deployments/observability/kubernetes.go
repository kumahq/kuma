package observability

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/gruntwork-io/terratest/modules/k8s"

	"github.com/kumahq/kuma/v3/test/framework"
)

//go:embed jaeger.yaml
var jaegerManifest string

const jaegerComponent = "jaeger"

type k8SDeployment struct {
	deploymentName  string
	namespace       string
	jaegerApiTunnel *k8s.Tunnel
}

var _ Deployment = &k8SDeployment{}

func (t *k8SDeployment) ZipkinCollectorURL() string {
	return fmt.Sprintf("http://jaeger-collector.%s:9411/api/v2/spans", t.namespace)
}

func (t *k8SDeployment) OTelCollectorTraceURL() string {
	return fmt.Sprintf("http://jaeger-collector.%s:4318/v1/traces", t.namespace)
}

func (t *k8SDeployment) TracedServices() ([]string, error) {
	return tracedServices(fmt.Sprintf("http://%s", t.jaegerApiTunnel.Endpoint()))
}

func (t *k8SDeployment) TracesForService(service string, limit int) ([]Trace, error) {
	return tracesForService(fmt.Sprintf("http://%s", t.jaegerApiTunnel.Endpoint()), service, limit)
}

func (t *k8SDeployment) Name() string {
	return t.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	yaml, err := renderJaegerManifest(t.namespace)
	if err != nil {
		return err
	}
	err = k8s.KubectlApplyFromStringContextE(cluster.GetTesting(), context.Background(),
		cluster.GetKubectlOptions(),
		yaml)
	if err != nil {
		return err
	}

	if err := framework.NewClusterSetup().
		Install(framework.WaitNumPods(t.namespace, 1, jaegerComponent)).
		Install(framework.WaitPodsAvailable(t.namespace, jaegerComponent)).
		Setup(cluster); err != nil {
		return err
	}

	podName, err := framework.PodNameOfApp(cluster, jaegerComponent, t.namespace)
	if err != nil {
		return err
	}
	t.jaegerApiTunnel = k8s.NewTunnel(cluster.GetKubectlOptions(t.namespace), k8s.ResourceTypePod, podName, 0, 16686)
	t.jaegerApiTunnel.ForwardPort(cluster.GetTesting())
	return nil
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.namespace)
}

func renderJaegerManifest(namespace string) (string, error) {
	tmpl, err := template.New("jaeger.yaml").Parse(jaegerManifest)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Namespace string
	}{
		Namespace: namespace,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
