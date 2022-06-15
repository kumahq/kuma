package main

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"kuma.io/cni/pkg/logger"
)

func newKubeClient(conf PluginConf) (*kubernetes.Clientset, error) {
	kubeconfig := conf.Kubernetes.Kubeconfig
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()

	if err != nil {
		logger.Default.Error("failed setting up kubernetes client with kubeconfig", zap.String("kubeconfig", kubeconfig))
		return nil, err
	}

	logger.Default.Debug("set up kubernetes client with kubeconfig",
		zap.String("kubeconfig", kubeconfig),
		zap.Any("config", config),
	)

	return kubernetes.NewForConfig(config)
}

// getK8sPodInfo returns information of a POD
func getKubePodInfo(client *kubernetes.Clientset, podName, podNamespace string) (containers int, initContainers map[string]struct{}, annotations map[string]string, err error) {
	pod, err := client.CoreV1().Pods(podNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	logger.Default.Debug("pod info", zap.Any("pod", pod))
	if err != nil {
		logger.Default.Error("can't get pod info", zap.Error(err))
		return 0, nil, nil, err
	}

	initContainers = map[string]struct{}{}
	for _, initContainer := range pod.Spec.InitContainers {
		initContainers[initContainer.Name] = struct{}{}
	}

	return len(pod.Spec.Containers), initContainers, pod.Annotations, nil
}
