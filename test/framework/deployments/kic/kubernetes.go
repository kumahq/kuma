package kic

import (
	"fmt"
	"time"

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
}

var _ Deployment = &k8sDeployment{}

var ingressApp = "ingress-kong"

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var yaml string
	var err error
	if t.ingressNamespace == "" {
		t.ingressNamespace = framework.Config.DefaultGatewayNamespace
	}
	yaml, err = cluster.GetKumactlOptions().RunKumactlAndGetOutputV(framework.Verbose,
		"install", "gateway", "kong",
		"--namespace", t.ingressNamespace,
		"--mesh", t.mesh,
	)
	if err != nil {
		return err
	}

	err = k8s.KubectlApplyFromStringE(cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		yaml)
	if err != nil {
		return err
	}

	k8s.WaitUntilNumPodsCreated(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.ingressNamespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", ingressApp),
		},
		1,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	pods := k8s.ListPods(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.ingressNamespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", ingressApp),
		},
	)
	if len(pods) != 1 {
		return errors.Errorf("counting KIC pods. Got: %d. Expected: 1", len(pods))
	}

	return k8s.WaitUntilPodAvailableE(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.ingressNamespace),
		pods[0].Name,
		framework.DefaultRetries,
		framework.DefaultTimeout)
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
