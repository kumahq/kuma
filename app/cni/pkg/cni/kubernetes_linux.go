package cni

import (
	"context"
	"slices"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func newKubeClient(logger logr.Logger, conf PluginConf) (*kubernetes.Clientset, error) {
	logger = logger.WithValues("kubeconfig", conf.Kubernetes.Kubeconfig)

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: conf.Kubernetes.Kubeconfig},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		logger.Error(err, "failed setting up kubernetes client with kubeconfig")
		return nil, err
	}

	logger.V(1).Info("set up kubernetes client with kubeconfig", "config", config)

	return kubernetes.NewForConfig(config)
}

// The returned bool indicates whether the pod should be skipped
// - If true, the pod is skipped, and any returned error is logged but doesn't
//   stop the process
// - If false, an error occurred during validation, and the CNI action should
//   fail
func getAndValidatePodAnnotations(
	ctx context.Context,
	logger logr.Logger,
	conf *PluginConf,
	k8sArgs K8sArgs,
) (map[string]string, bool, error) {
	name := string(k8sArgs.K8S_POD_NAME)
	namespace := string(k8sArgs.K8S_POD_NAMESPACE)

	if namespace == "" || name == "" {
		return nil, true, errors.New("pod namespace or pod name is empty")
	}

	if slices.Contains(conf.Kubernetes.ExcludeNamespaces, namespace) {
		return nil, true, errors.New("namespace is in the exclude list")
	}

	client, err := newKubeClient(logger, *conf)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to create Kubernetes client")
	}

	backoff := retry.WithMaxRetries(
		podRetrievalMaxRetries,
		retry.NewConstant(podRetrievalInterval),
	)

	var pod *corev1.Pod
	if err := retry.Do(
		ctx,
		backoff,
		func(ctx context.Context) error {
			pod, err = client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			return retry.RetryableError(err)
		},
	); err != nil {
		return nil, false, errors.Wrap(err, "failed to retrieve pod data from Kubernetes API")
	}

	logger.V(1).Info(
		"retrieved pod data",
		"name", pod.Name,
		"namespace", pod.Namespace,
		"annotations", pod.Annotations,
		"containersCount", len(pod.Spec.Containers),
		"initContainersCount", len(pod.Spec.InitContainers),
	)

	containers := make(map[string]struct{}, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		containers[container.Name] = struct{}{}
	}

	initContainers := make(map[string]struct{}, len(pod.Spec.InitContainers))
	for _, container := range pod.Spec.InitContainers {
		initContainers[container.Name] = struct{}{}
	}

	if _, ok := initContainers[k8s_util.KumaInitContainerName]; ok {
		return nil, true, errors.New("pod already injected with kuma-init container")
	}

	_, sidecarInContainers := containers[k8s_util.KumaSidecarContainerName]
	_, sidecarInInitContainers := initContainers[k8s_util.KumaSidecarContainerName]
	if !sidecarInContainers && !sidecarInInitContainers {
		return nil, true, errors.New("missing required kuma-sidecar container")
	}

	if pod.Annotations[metadata.KumaSidecarInjectedAnnotation] != "true" {
		return nil, true, errors.Errorf(
			"annotation '%s' is missing or is not set to 'true'",
			metadata.KumaSidecarInjectedAnnotation,
		)
	}

	return pod.Annotations, true, nil
}
