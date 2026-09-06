package certmanager

import (
	"context"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
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

	// The chart comes from an external repository, so a single unlucky fetch
	// ("charts.jetstack.io/index.yaml": EOF) would otherwise take down the whole
	// suite from BeforeAll. `upgrade --install` keeps the retry idempotent if an
	// attempt fails after the release was created.
	_, err := retry.DoWithRetryContextE(
		cluster.GetTesting(), context.Background(),
		"install cert-manager", 3, 5*time.Second,
		func() (string, error) {
			return helm.RunHelmCommandAndGetStdOutContextE(cluster.GetTesting(), context.Background(), &opts, "upgrade", "cert-manager",
				"--install",
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
		},
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
