package cni

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func newKubeClient(l logr.Logger, conf PluginConf) (*kubernetes.Clientset, error) {
	l = log.WithValues("kubeconfig", conf.Kubernetes.Kubeconfig)

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: conf.Kubernetes.Kubeconfig},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		l.Error(err, "failed setting up kubernetes client with kubeconfig")
		return nil, err
	}

	l.V(1).Info("set up kubernetes client with kubeconfig", "config", config)

	return kubernetes.NewForConfig(config)
}

// getK8sPodInfo returns information of a POD
func getKubePodInfo(ctx context.Context, client *kubernetes.Clientset, podName, podNamespace string) (int, map[string]struct{}, map[string]string, error) {
	pod, err := client.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
	log.V(1).Info("pod info", "pod", pod)
	if err != nil {
		log.Error(err, "can't get pod info")
		return 0, nil, nil, err
	}

	initContainers := map[string]struct{}{}
	for _, initContainer := range pod.Spec.InitContainers {
		initContainers[initContainer.Name] = struct{}{}
	}

	return len(pod.Spec.Containers), initContainers, pod.Annotations, nil
}
