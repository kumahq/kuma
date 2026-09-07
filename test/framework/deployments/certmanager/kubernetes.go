package certmanager

import (
	"context"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/test/framework"
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

	// Pull through the shared chart cache: the first run fetches from the
	// external repository with retries, later runs reuse the tarball and never
	// touch the network. Installing from a local chart also keeps the install
	// itself off the network, so a flaky index fetch cannot fail BeforeAll and
	// take the whole suite with it.
	chartPath, err := framework.HelmChartFromRepoE(
		cluster.GetTesting(),
		"https://charts.jetstack.io",
		"cert-manager",
		t.version,
	)
	if err != nil {
		return err
	}

	// `upgrade --install` keeps this idempotent if an earlier attempt failed
	// after the release was created.
	_, err = helm.RunHelmCommandAndGetStdOutContextE(cluster.GetTesting(), context.Background(), &opts, "upgrade", "cert-manager",
		"--install",
		"--namespace", t.namespace,
		"--create-namespace",
		"--set", "installCRDs=true",
		"--set", "startupapicheck.enabled=false",
		"--wait",
		"--timeout", "5m",
		chartPath,
	)
	if err != nil {
		return err
	}

	// Wait for cert-manager pods to be ready
	podSelectors := []string{
		"app.kubernetes.io/name=cert-manager",
		"app.kubernetes.io/name=cainjector",
		"app.kubernetes.io/name=webhook",
	}
	for _, selector := range podSelectors {
		if err := t.isPodReady(cluster, selector); err != nil {
			return err
		}
	}

	return nil
}

func (t *k8sDeployment) isPodReady(cluster framework.Cluster, selector string) error {
	return errors.Wrapf(
		k8s.WaitUntilNumPodsCreatedContextE(cluster.GetTesting(), context.Background(),
			cluster.GetKubectlOptions(t.namespace),
			metav1.ListOptions{
				LabelSelector: selector,
			},
			1,
			framework.DefaultRetries,
			framework.DefaultTimeout),
		"cert-manager pod with selector %q in namespace %q failed to become ready",
		selector, t.namespace)
}

func (t *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.namespace)
}
