package kic

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
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

	k8s.WaitUntilPodAvailable(cluster.GetTesting(),
		cluster.GetKubectlOptions(t.ingressNamespace),
		pods[0].Name,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	return nil
}

func (t *k8sDeployment) Delete(cluster framework.Cluster) error {
	return cluster.DeleteNamespace(t.ingressNamespace)
}
