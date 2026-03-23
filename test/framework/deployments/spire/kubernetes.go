package spire

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v2/test/framework"
)

type k8sDeployment struct {
	namespace      string
	name           string
	trustDomain    string
	kubectlVersion string
}

func (*k8sDeployment) GetIP() (string, error) {
	panic("not implemented")
}

var _ framework.Deployment = &k8sDeployment{}

func (*k8sDeployment) Name() string {
	return DeploymentName
}

func (d *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var err error
	if d.namespace == "" {
		d.namespace = "spire"
	}

	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(d.namespace),
	}
	// install crds
	_, err = helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "spire-crds",
		"--namespace", d.namespace,
		"--repo", "https://spiffe.github.io/helm-charts-hardened/",
		"spire-crds",
	)
	if err != nil {
		return err
	}

	_, err = helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "spire",
		"--namespace", d.namespace,
		"--repo", "https://spiffe.github.io/helm-charts-hardened/",
		"--set", "global.spire.trustDomain="+d.trustDomain,
		"--set", "global.spire.tools.kubectl.tag="+d.kubectlVersion,
		"spire",
	)
	if err != nil {
		return err
	}

	err = d.isPodReady(cluster, "app.kubernetes.io/name=agent")
	if err != nil {
		return err
	}
	err = d.isPodReady(cluster, "app.kubernetes.io/name=server")
	if err != nil {
		return err
	}
	err = d.isPodReady(cluster, "app.kubernetes.io/name=spiffe-csi-driver")
	if err != nil {
		return err
	}
	err = d.isPodReady(cluster, "app.kubernetes.io/name=spiffe-oidc-discovery-provider")
	if err != nil {
		return err
	}

	return nil
}

func (d *k8sDeployment) isPodReady(cluster framework.Cluster, selector string) error {
	err := k8s.WaitUntilNumPodsCreatedE(cluster.GetTesting(),
		cluster.GetKubectlOptions(d.namespace),
		metav1.ListOptions{
			LabelSelector: selector,
		},
		1,
		framework.DefaultRetries*3, // spire is fetched from the internet. Increase the timeout to prevent long downloads of images.
		framework.DefaultTimeout)
	if err != nil {
		return err
	}

	pods := k8s.ListPods(cluster.GetTesting(),
		cluster.GetKubectlOptions(d.namespace),
		metav1.ListOptions{
			LabelSelector: selector,
		},
	)
	if len(pods) == 0 {
		return errors.Errorf("no pods found with selector %q in namespace %q", selector, d.namespace)
	}

	for _, pod := range pods {
		if err := k8s.WaitUntilPodAvailableE(cluster.GetTesting(),
			cluster.GetKubectlOptions(d.namespace),
			pod.Name,
			framework.DefaultRetries*3, // spire is fetched from the internet. Increase the timeout to prevent long downloads of images.
			framework.DefaultTimeout); err != nil {
			return err
		}
	}

	return nil
}

func (d *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(d.namespace)
}
