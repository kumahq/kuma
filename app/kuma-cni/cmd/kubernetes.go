package cmd

import (
	"context"
	"strconv"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
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
	Log.Infof("pod info %+v", pod)
	if err != nil {
		return nil, err
	}

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
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return err
	}
	Log.Infof("Getting identifiers with arguments: %s", args.Args)
	Log.Infof("Loaded k8s arguments: %v", k8sArgs)

	// Check if the workload is running under Kubernetes.
	if string(k8sArgs.K8S_POD_NAMESPACE) != "" && string(k8sArgs.K8S_POD_NAME) != "" {
		excludePod := false
		for _, excludeNs := range conf.Kubernetes.ExcludeNamespaces {
			if string(k8sArgs.K8S_POD_NAMESPACE) == excludeNs {
				excludePod = true
				break
			}
		}
		if !excludePod {
			client, err := newKubeClient(*conf)
			if err != nil {
				return err
			}
			Log.Debugf("Created Kubernetes client: %v", client)
			pi := &PodInfo{}
			var k8sErr error
			for attempt := 1; attempt <= podRetrievalMaxRetries; attempt++ {
				pi, k8sErr = getKubePodInfo(client, string(k8sArgs.K8S_POD_NAME), string(k8sArgs.K8S_POD_NAMESPACE))
				if k8sErr == nil {
					break
				}
				Log.Warnf("err %s, attempt %d: Waiting for pod metadata", k8sErr, attempt)
				time.Sleep(podRetrievalInterval)
			}
			if k8sErr != nil {
				Log.Errorf("Failed to get pod data: %v", k8sErr)
				return k8sErr
			}

			// Check if kuma-init container is present; in that case exclude pod
			if _, present := pi.InitContainers[KUMAINIT]; present {
				Log.Infof("pod %s namespace %s: pod excluded due to being already injected with kuma-init container",
					string(k8sArgs.K8S_POD_NAME),
					string(k8sArgs.K8S_POD_NAMESPACE))
				excludePod = true
			}

			Log.Infof("Found containers %v", pi.Containers)
			if len(pi.Containers) > 1 {
				Log.Infof("ContainerID %s, netns %s, pod %s, Namespace %s annotations %s: Checking annotations prior to redirect for Kuma proxy",
					args.ContainerID, args.Netns,
					string(k8sArgs.K8S_POD_NAME), string(k8sArgs.K8S_POD_NAMESPACE),
					pi.pod.Annotations)

				if val, ok := pi.pod.Annotations[injectAnnotationKey]; ok {
					Log.Infof("Pod %s contains inject annotation: %s", string(k8sArgs.K8S_POD_NAME), val)
					if injectEnabled, err := strconv.ParseBool(val); err == nil {
						if !injectEnabled {
							Log.Infof("Pod excluded due to inject-disabled annotation")
							excludePod = true
						}
					}
				}

				if _, ok := pi.pod.Annotations[sidecarStatusKey]; !ok {
					Log.Infof("Pod %s excluded due to not containing sidecar annotation", string(k8sArgs.K8S_POD_NAME))
					excludePod = true
				}

				if !excludePod {
					Log.Infof("setting up redirect")
					// Invoke redirect
					return redirect(args.Netns, pi)
				}
			}
		} else {
			Log.Infof("Pod excluded")
		}
	} else {
		Log.Infof("No Kubernetes Data")
	}

	return nil
}
