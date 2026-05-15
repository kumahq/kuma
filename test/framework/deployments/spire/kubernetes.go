package spire

import (
	"regexp"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v2/test/framework"
)

const (
	spireChartRepo        = "https://spiffe.github.io/helm-charts-hardened/"
	spireChartName        = "spire"
	spireChartVersion     = "0.28.4"
	spireCRDsChartName    = "spire-crds"
	spireCRDsChartVersion = "0.5.0"
)

type k8sDeployment struct {
	namespace      string
	name           string
	trustDomain    string
	kubectlVersion string
}

func (u *k8sDeployment) GetIP() (string, error) {
	panic("not implemented")
}

var _ framework.Deployment = &k8sDeployment{}

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	if t.namespace == "" {
		t.namespace = "spire"
	}

	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(t.namespace),
	}

	spireSetValues := []string{
		"--set", "global.spire.trustDomain=" + t.trustDomain,
		"--set", "global.spire.tools.kubectl.tag=" + t.kubectlVersion,
	}

	spireCRDsChart, err := framework.HelmChartFromRepoE(
		cluster.GetTesting(),
		spireChartRepo,
		spireCRDsChartName,
		spireCRDsChartVersion,
	)
	if err != nil {
		return err
	}
	spireChart, err := framework.HelmChartFromRepoE(
		cluster.GetTesting(),
		spireChartRepo,
		spireChartName,
		spireChartVersion,
	)
	if err != nil {
		return err
	}

	// Preload chart images into the cluster's container runtime so pod startup
	// does not hit ghcr.io / docker.io at test time. Transient registry
	// timeouts during runtime pulls are the dominant cause of ImagePullBackOff
	// flakes on this BeforeAll.
	if k8sCluster, ok := cluster.(*framework.K8sCluster); ok {
		images, err := chartImages(cluster, "spire-crds", &opts, spireCRDsChart)
		if err != nil {
			return errors.Wrap(err, "render spire-crds chart")
		}
		more, err := chartImages(cluster, "spire", &opts, spireChart, spireSetValues...)
		if err != nil {
			return errors.Wrap(err, "render spire chart")
		}
		if err := k8sCluster.PreloadImages(dedupe(append(images, more...))...); err != nil {
			return errors.Wrap(err, "preload spire images")
		}
	}

	if _, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", "spire-crds",
		"--namespace", t.namespace,
		spireCRDsChart,
	); err != nil {
		return err
	}

	installArgs := []string{
		"--namespace", t.namespace,
	}
	installArgs = append(installArgs, spireSetValues...)
	installArgs = append(installArgs, spireChart)
	if _, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", append([]string{"spire"}, installArgs...)...); err != nil {
		return err
	}

	for _, selector := range []string{
		"app.kubernetes.io/name=agent",
		"app.kubernetes.io/name=server",
		"app.kubernetes.io/name=spiffe-csi-driver",
		"app.kubernetes.io/name=spiffe-oidc-discovery-provider",
	} {
		if err := t.isPodReady(cluster, selector); err != nil {
			return err
		}
	}
	return nil
}

// chartImages renders the named chart with `helm template` and returns the
// distinct container images referenced in the resulting manifests.
func chartImages(cluster framework.Cluster, releaseName string, opts *helm.Options, chart string, extraArgs ...string) ([]string, error) {
	args := []string{releaseName, chart}
	args = append(args, extraArgs...)
	out, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), opts, "template", args...)
	if err != nil {
		return nil, err
	}
	return parseImages(out), nil
}

var imageLineRE = regexp.MustCompile(`(?m)^\s*image:\s*["']?([^"'\s]+)["']?\s*$`)

func parseImages(rendered string) []string {
	matches := imageLineRE.FindAllStringSubmatch(rendered, -1)
	seen := map[string]struct{}{}
	var images []string
	for _, m := range matches {
		img := m[1]
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		images = append(images, img)
	}
	return images
}

func dedupe(images []string) []string {
	seen := map[string]struct{}{}
	out := images[:0]
	for _, img := range images {
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		out = append(out, img)
	}
	return out
}

func (t *k8sDeployment) isPodReady(cluster framework.Cluster, selector string) error {
	err := k8s.WaitUntilNumPodsCreatedE(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.namespace),
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
		cluster.GetKubectlOptions(t.namespace),
		metav1.ListOptions{
			LabelSelector: selector,
		},
	)
	if len(pods) == 0 {
		return errors.Errorf("no pods found with selector %q in namespace %q", selector, t.namespace)
	}

	for _, pod := range pods {
		if err := k8s.WaitUntilPodAvailableE(cluster.GetTesting(),
			cluster.GetKubectlOptions(t.namespace),
			pod.Name,
			framework.DefaultRetries*3, // spire is fetched from the internet. Increase the timeout to prevent long downloads of images.
			framework.DefaultTimeout); err != nil {
			return err
		}
	}

	return nil
}

func (t *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.namespace)
}
