package spire

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type k8sDeployment struct {
	namespace      string
	name           string
	trustDomain    string
	kubectlVersion string
}

var _ Deployment = &k8sDeployment{}

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var err error
	if t.namespace == "" {
		t.namespace = "spire"
	}

	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(t.namespace),
	}
	// install crds
	_, err = helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "spire-crds",
		"--namespace", t.namespace,
		"--repo", "https://spiffe.github.io/helm-charts-hardened/",
		"spire-crds",
	)
	if err != nil {
		return err
	}

	_, err = helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "spire",
		"--namespace", t.namespace,
		"--repo", "https://spiffe.github.io/helm-charts-hardened/",
		"--set", "global.spire.trustDomain="+t.trustDomain,
		"--set", "global.spire.tools.kubectl.tag="+t.kubectlVersion,
		"spire",
	)
	if err != nil {
		return err
	}

	err = t.isPodReady(cluster, "app.kubernetes.io/name=agent")
	if err != nil {
		return err
	}
	err = t.isPodReady(cluster, "app.kubernetes.io/name=server")
	if err != nil {
		return err
	}
	err = t.isPodReady(cluster, "app.kubernetes.io/name=spiffe-csi-driver")
	if err != nil {
		return err
	}
	err = t.isPodReady(cluster, "app.kubernetes.io/name=spiffe-oidc-discovery-provider")
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
