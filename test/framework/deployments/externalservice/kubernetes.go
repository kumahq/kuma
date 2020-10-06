package externalservice

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	ip   string
	port uint32
	name string
	args []string
}

var _ Deployment = &k8SDeployment{}

const externalServiceNamespace = DeploymentName + "namespace"

func (k *k8SDeployment) Name() string {
	return DeploymentName + k.name
}

func (k *k8SDeployment) Deploy(cluster framework.Cluster) error {
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
	return nil
}

func (k *k8SDeployment) Init(cluster framework.Cluster, name string, args []string) error {
	k.name = name
	k.args = args

	err := framework.Namespace(externalServiceNamespace)(cluster)
	if err != nil {
		return err
	}

	return nil
}

func (k *k8SDeployment) GetExternalAppAddress() string {
	return k.ip
}

func (k *k8SDeployment) Cleanup(cluster framework.Cluster) error {
	err := k8s.DeleteNamespaceE(cluster.GetTesting(),
		cluster.GetKubectlOptions(externalServiceNamespace),
		externalServiceNamespace)
	if err != nil {
		return err
	}

	cluster.(*framework.K8sCluster).WaitNamespaceDelete(externalServiceNamespace)
	return nil
}
