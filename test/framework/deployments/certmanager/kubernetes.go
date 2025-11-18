package certmanager

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v2/test/framework"
)

type k8sDeployment struct {
	namespace string
	version   string
}

var _ Deployment = &k8sDeployment{}

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(t.namespace),
	}

	// Install cert-manager via Helm
	_, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "cert-manager",
		"--namespace", t.namespace,
		"--create-namespace",
		"--repo", "https://charts.jetstack.io",
		"--version", t.version,
		"--set", "installCRDs=true",
		"--set", "startupapicheck.enabled=false",
		"--wait",
		"--timeout", "5m",
		"cert-manager",
	)
	if err != nil {
		return err
	}

	// Wait for cert-manager pods to be ready
	err = t.isPodReady(cluster, "app.kubernetes.io/name=cert-manager")
	if err != nil {
		return err
	}
	err = t.isPodReady(cluster, "app.kubernetes.io/name=cainjector")
	if err != nil {
		return err
	}
	err = t.isPodReady(cluster, "app.kubernetes.io/name=webhook")
	if err != nil {
		return err
	}

	return nil
}

func (t *k8sDeployment) isPodReady(cluster framework.Cluster, selector string) error {
	err := k8s.WaitUntilNumPodsCreatedE(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.namespace),
		metav1.ListOptions{
			LabelSelector: selector,
		},
		1,
		framework.DefaultRetries,
		framework.DefaultTimeout)
	if err != nil {
		return err
	}
	return nil
}

func (t *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.namespace)
}
