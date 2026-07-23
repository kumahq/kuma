package kic

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

const (
	kicChartRepo    = "https://charts.konghq.com"
	kicChartName    = "ingress"
	kicChartVersion = "0.24.0"
)

type k8sDeployment struct {
	ingressNamespace string
	watchNamespaces  []string
	mesh             string
	name             string
}

var _ Deployment = &k8sDeployment{}

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var err error
	if t.ingressNamespace == "" {
		t.ingressNamespace = framework.Config.DefaultGatewayNamespace
	}

	if len(t.watchNamespaces) == 0 {
		t.watchNamespaces = []string{t.ingressNamespace}
	} else if !slices.Contains(t.watchNamespaces, t.ingressNamespace) {
		t.watchNamespaces = append(t.watchNamespaces, t.ingressNamespace)
	}
	watchNamespacesVal := strings.Join(t.watchNamespaces, ",")

	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(t.ingressNamespace),
	}

	chartPath, err := framework.HelmChartFromRepoE(
		cluster.GetTesting(),
		kicChartRepo,
		kicChartName,
		kicChartVersion,
	)
	if err != nil {
		return err
	}

	_, err = helm.RunHelmCommandAndGetStdOutContextE(cluster.GetTesting(), context.Background(), &opts, "install", t.name,
		"--namespace", t.ingressNamespace,
		"--set", "controller.ingressController.watchNamespaces={"+watchNamespacesVal+"}",
		"--set", "controller.ingressController.ingressClass="+t.name,
		"--set", "controller.podLabels.kuma\\.io/mesh="+t.mesh,
		"--set", "gateway.podLabels.kuma\\.io/mesh="+t.mesh,
		chartPath,
	)
	if err != nil {
		return err
	}

	for _, app := range []string{fmt.Sprintf("%s-controller", t.name), fmt.Sprintf("%s-gateway", t.name)} {
		err := k8s.WaitUntilNumPodsCreatedContextE(cluster.GetTesting(), context.Background(),
			cluster.GetKubectlOptions(t.ingressNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			},
			1,
			framework.DefaultRetries,
			framework.DefaultTimeout)
		if err != nil {
			return err
		}

		pods := k8s.ListPodsContext(cluster.GetTesting(), context.Background(),
			cluster.GetKubectlOptions(t.ingressNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			},
		)
		if len(pods) != 1 {
			return errors.Errorf("counting KIC pods. Got: %d. Expected: 1", len(pods))
		}

		err = k8s.WaitUntilPodAvailableContextE(cluster.GetTesting(), context.Background(),
			cluster.GetKubectlOptions(t.ingressNamespace),
			pods[0].Name,
			framework.DefaultRetries*3, // KIC is fetched from the internet. Increase the timeout to prevent long downloads of images.
			framework.DefaultTimeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.(*framework.K8sCluster).TriggerDeleteNamespace(t.ingressNamespace)
}

func (t *k8sDeployment) IP(namespace string) (string, error) {
	ip, err := retry.DoWithRetryInterfaceContextE(
		kubernetes.Cluster.GetTesting(), context.Background(),
		"get the clusterIP of the Kong Ingress Controller Service",
		60,
		time.Second,
		func() (any, error) {
			svc, err := k8s.GetServiceContextE(
				kubernetes.Cluster.GetTesting(), context.Background(),
				kubernetes.Cluster.GetKubectlOptions(namespace),
				"gateway",
			)
			if err != nil || svc.Spec.ClusterIP == "" {
				return nil, errors.Wrapf(err, "could not get clusterIP")
			}

			return svc.Spec.ClusterIP, nil
		},
	)
	if err != nil {
		return "", err
	}

	return ip.(string), nil
}
