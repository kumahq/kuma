package kic

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

type k8sDeployment struct {
	ingressNamespace string
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
	opts := helm.Options{
		KubectlOptions: cluster.GetKubectlOptions(t.ingressNamespace),
	}
	_, err = helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(), &opts, "install", t.name,
		"--namespace", t.ingressNamespace,
		"--repo", "https://charts.konghq.com",
		"--set", "controller.ingressController.ingressClass="+t.name,
		"--set", "controller.podAnnotations.kuma\\.io/mesh="+t.mesh,
		"--set", "gateway.podAnnotations.kuma\\.io/mesh="+t.mesh,
		"ingress",
	)
	if err != nil {
		return err
	}

	for _, app := range []string{fmt.Sprintf("%s-controller", t.name), fmt.Sprintf("%s-gateway", t.name)} {
		err := k8s.WaitUntilNumPodsCreatedE(cluster.GetTesting(),
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

		pods := k8s.ListPods(cluster.GetTesting(),
			cluster.GetKubectlOptions(t.ingressNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			},
		)
		if len(pods) != 1 {
			return errors.Errorf("counting KIC pods. Got: %d. Expected: 1", len(pods))
		}

		err = k8s.WaitUntilPodAvailableE(cluster.GetTesting(),
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
	return cluster.DeleteNamespace(t.ingressNamespace)
}

func (t *k8sDeployment) IP(namespace string) (string, error) {
	ip, err := retry.DoWithRetryInterfaceE(
		kubernetes.Cluster.GetTesting(),
		"get the clusterIP of the Kong Ingress Controller Service",
		60,
		time.Second,
		func() (interface{}, error) {
			svc, err := k8s.GetServiceE(
				kubernetes.Cluster.GetTesting(),
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
