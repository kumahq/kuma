package kic

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type k8sDeployment struct {
	ingressNamespace string
}

var DefaultNodePortHTTP = 30080
var DefaultNodePortHTTPS = 30443

var _ Deployment = &k8sDeployment{}

var ingressApp = "ingress-kong"

func NodePortHTTP() int {
	var port int
	var err error
	portStr := os.Getenv("E2E_KONG_INGRESS_HTTP_PORT")
	if portStr == "" {
		port = DefaultNodePortHTTP
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Sprintf("Invalid E2E_KONG_INGRESS_HTTP_PORT: %s", portStr))
		}
	}
	return port
}

func NodePortHTTPS() int {
	var port int
	var err error
	portStr := os.Getenv("E2E_KONG_INGRESS_HTTPS_PORT")
	if portStr == "" {
		port = DefaultNodePortHTTPS
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Sprintf("Invalid E2E_KONG_INGRESS_HTTPS_PORT: %s", portStr))
		}
	}
	return port
}

func (t *k8sDeployment) Name() string {
	return DeploymentName
}

func (t *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var yaml string
	var err error
	if t.ingressNamespace == "" {
		yaml, err = cluster.GetKumactlOptions().RunKumactlAndGetOutputV(framework.Verbose, "install", "gateway", "kong")
		t.ingressNamespace = framework.Config.DefaultGatewayNamespace
	} else {
		yaml, err = cluster.GetKumactlOptions().RunKumactlAndGetOutputV(framework.Verbose, "install", "gateway", "kong", "--namespace", t.ingressNamespace)
	}

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
