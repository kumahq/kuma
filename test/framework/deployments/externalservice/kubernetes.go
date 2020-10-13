package externalservice

import (
	"fmt"
	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	ip   string
	name string
	args []string
}

var _ Deployment = &k8SDeployment{}

const externalServiceNamespace = DeploymentName + "namespace"

func (k *k8SDeployment) Name() string {
	return DeploymentName + k.name
}

func (k *k8SDeployment) Deploy(cluster framework.Cluster) error {
	if _, err := k8s.GetNamespaceE(cluster.GetTesting(),
		cluster.GetKubectlOptions(externalServiceNamespace),
		externalServiceNamespace); err != nil {
		// create the namespace
		err := framework.Namespace(externalServiceNamespace)(cluster)
		if err != nil {
			return err
		}
	}

	err := framework.YamlPathK8s(filepath.Join("testdata", fmt.Sprintf("%s.yaml", k.Name())))(cluster)
	if err != nil {
		return err
	}

	k8s.WaitUntilNumPodsCreated(cluster.GetTesting(),
		cluster.GetKubectlOptions(externalServiceNamespace),
		metav1.ListOptions{
			LabelSelector: "app=" + k.Name(),
		},
		1,
		framework.DefaultRetries,
		framework.DefaultTimeout)

	err = framework.WaitPodsAvailable(externalServiceNamespace, k.Name())(cluster)
	if err != nil {
		return err
	}

	k.ip = k.Name() + "." + externalServiceNamespace

	return nil
}

func (k *k8SDeployment) Delete(cluster framework.Cluster) error {
	err := k8s.KubectlDeleteE(cluster.GetTesting(),
		cluster.GetKubectlOptions(externalServiceNamespace),
		filepath.Join("testdata", fmt.Sprintf("%s.yaml", k.Name())))
	if err != nil {
		return err
	}

	framework.WaitPodsNotAvailable(externalServiceNamespace, k.Name())

	pods, err := k8s.ListPodsE(cluster.GetTesting(), cluster.GetKubectlOptions(externalServiceNamespace), metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		err := k8s.DeleteNamespaceE(cluster.GetTesting(),
			cluster.GetKubectlOptions(externalServiceNamespace),
			externalServiceNamespace)
		if err != nil {
			return err
		}
		cluster.(*framework.K8sCluster).WaitNamespaceDelete(externalServiceNamespace)
	}

	return nil
}

func (k *k8SDeployment) GetExternalAppAddress() string {
	return k.ip
}
