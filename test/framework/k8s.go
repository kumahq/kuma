package framework

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PodsAvailable(cluster Cluster, name string, namespace string) (int, error) {
	pods, err := k8s.ListPodsE(cluster.GetTesting(), cluster.GetKubectlOptions(namespace),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", name)})
	if err != nil {
		return 0, err
	}
	podsAvailable := 0
	for _, p := range pods {
		if k8s.IsPodAvailable(&p) {
			podsAvailable++
		}
	}
	return podsAvailable, nil
}

func PodNameOfApp(cluster Cluster, name string, namespace string) (string, error) {
	pods, err := k8s.ListPodsE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(namespace),
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", name),
		},
	)
	if err != nil {
		return "", err
	}
	if len(pods) != 1 {
		return "", errors.Errorf("expected %d pods, got %d", 1, len(pods))
	}
	return pods[0].Name, nil
}
