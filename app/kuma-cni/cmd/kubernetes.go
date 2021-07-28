package cmd

import (
	"context"
	"strconv"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/kumahq/kuma/app/kuma-cni/pkg/log"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

type PodInfo struct {
	pod            *kube_core.Pod
	Containers     []string
	InitContainers map[string]struct{}
}

func newKubeClient(conf PluginConf) (*kubernetes.Clientset, error) {
	// Some config can be passed in a kubeconfig file
	kubeconfig := conf.Kubernetes.Kubeconfig

	config, err := k8s.DefaultClientConfig(kubeconfig, "")
	if err != nil {
		Log.Infof("Failed setting up kubernetes client with kubeconfig %s", kubeconfig)
		return nil, err
	}

	Log.Infof("Set up kubernetes client with kubeconfig %s", kubeconfig)
	Log.Infof("Kubernetes config %v", config)

	return kubernetes.NewForConfig(config)
}

func getKubePodInfo(client *kubernetes.Clientset, podName, podNamespace string) (*PodInfo, error) {
	pod, err := client.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	Log.Infof("pod info %+v", pod)

	pi := &PodInfo{
		pod:            pod,
		InitContainers: make(map[string]struct{}),
		Containers:     make([]string, len(pod.Spec.Containers)),
	}

	for _, initContainer := range pi.pod.Spec.InitContainers {
		pi.InitContainers[initContainer.Name] = struct{}{}
	}

	for containerIdx, container := range pi.pod.Spec.Containers {
		Log.Debugf("pod %s container %s: Inspecting container", podName, container.Name)

		pi.Containers[containerIdx] = container.Name
	}

	return pi, nil
}

func ProcessK8s(args *skel.CmdArgs, conf *PluginConf) error {
	// Determine if running under k8s by checking the CNI args
	var k8sArgs K8sArgs
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return err
	}

	Log.Infof("Getting identifiers with arguments: %s", args.Args)
	Log.Infof("Loaded k8s arguments: %v", k8sArgs)

	podNamespace := string(k8sArgs.K8S_POD_NAMESPACE)
	podName := string(k8sArgs.K8S_POD_NAME)

	// Check if the workload is running under Kubernetes.
	if podNamespace == "" || podName == "" {
		Log.Infof("No Kubernetes Data")

		return nil
	}

	if isNamespaceExcluded(conf.Kubernetes.ExcludeNamespaces, podNamespace) {
		Log.Infof("Pod excluded")

		return nil
	}

	client, err := newKubeClient(*conf)
	if err != nil {
		return err
	}

	Log.Debugf("Created Kubernetes client: %v", client)

	pi, err := tryRetrievePodInfo(client, podName, podNamespace)
	if err != nil {
		return err
	}

	// Check if kuma-init container is present; in that case exclude pod
	if _, present := pi.InitContainers[KUMAINIT]; present {
		Log.Infof(
			"pod %s namespace %s: pod excluded due to being already injected with kuma-init container",
			podName,
			podNamespace,
		)

		return nil
	}

	Log.Infof("Found containers %v", pi.Containers)

	if len(pi.Containers) > 1 {
		Log.Infof(
			"ContainerID %s, netns %s, pod %s, Namespace %s annotations %s: Checking annotations prior to redirect for Kuma proxy",
			args.ContainerID,
			args.Netns,
			podName,
			podNamespace,
			pi.pod.Annotations,
		)

		if val, ok := pi.pod.Annotations[injectAnnotationKey]; ok {
			Log.Infof("Pod %s contains inject annotation: %s", podName, val)

			if injectEnabled, err := strconv.ParseBool(val); err == nil {
				if !injectEnabled {
					Log.Infof("Pod excluded due to inject-disabled annotation")

					return nil
				}
			}
		}

		if _, ok := pi.pod.Annotations[sidecarStatusKey]; !ok {
			Log.Infof("Pod %s excluded due to not containing sidecar annotation", podName)

			return nil
		}

		Log.Infof("setting up redirect")

		// Invoke redirect
		return redirect(args.Netns, pi)
	}

	return nil
}

func tryRetrievePodInfo(client *kubernetes.Clientset, podName string, podNamespace string) (*PodInfo, error) {
	for attempt := 1; attempt <= podRetrievalMaxRetries; attempt++ {
		pi, err := getKubePodInfo(client, podName, podNamespace)
		if err == nil {
			return pi, nil
		}

		Log.Warnf("err %s, attempt %d: Waiting for pod metadata", err, attempt)

		time.Sleep(podRetrievalInterval)
	}

	return nil, errors.New("Failed to get pod data: max retries reached")
}

func isNamespaceExcluded(excludedNamespaces []string, podNamespace string) bool {
	for _, excludeNs := range excludedNamespaces {
		if podNamespace == excludeNs {
			return true
		}
	}

	return false
}
