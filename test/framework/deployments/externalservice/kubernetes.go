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
	cmd  Command
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
	// Forcefully delete the namespace, no matter if there are other pods
	// this is to ensure that all the relevant resources are cleaned up.
	// We do ignore the error here, because any subsequent invocation of
	// this code will essentially fail.
	_ = k8s.DeleteNamespaceE(cluster.GetTesting(),
		cluster.GetKubectlOptions(externalServiceNamespace),
		externalServiceNamespace)

	cluster.(*framework.K8sCluster).WaitNamespaceDelete(externalServiceNamespace)

	return nil
}

func (k *k8SDeployment) GetExternalAppAddress() string {
	return k.ip
}

func (k *k8SDeployment) GetCert() string {
	// We do not implement Runtime Certificate injection on K8s
	// The functionality is test on Universal which is good for now
	panic("not implemented")
}
