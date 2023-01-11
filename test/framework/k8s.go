package framework

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func PodIPOfApp(cluster Cluster, name string, namespace string) (string, error) {
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
	return pods[0].Status.PodIP, nil
}

// WaitUntilPodAvailableE it's a "fork" of original WaitUntilPodAvailableE from terratest with more information why the pod failed
// Remove when https://github.com/gruntwork-io/terratest/pull/1223 is merged.
func WaitUntilPodAvailableE(t testing.TestingT, options *k8s.KubectlOptions, podName string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for pod %s to be provisioned.", podName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			pod, err := k8s.GetPodE(t, options, podName)
			if err != nil {
				return "", err
			}
			if !k8s.IsPodAvailable(pod) {
				return "", errors.Errorf("Pod %s is not available, reason: %s, message: %s", pod.Name, pod.Status.Reason, pod.Status.Message)
			}
			return "Pod is now available", nil
		},
	)
	if err != nil {
		Logf("Timedout waiting for Pod to be provisioned: %s", err)
		return err
	}
	Logf(message)
	return nil
}
