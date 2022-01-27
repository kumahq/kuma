package injector

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	tp_k8s "github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
)

const (
	// serviceAccountTokenMountPath is a well-known location where Kubernetes mounts a ServiceAccount token.
	serviceAccountTokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

var log = core.Log.WithName("injector")

func New(
	cfg runtime_k8s.Injector,
	controlPlaneURL string,
	client kube_client.Client,
	converter k8s_common.Converter,
	envoyAdminPort uint32,
) (*KumaInjector, error) {
	var caCert string
	if cfg.CaCertFile != "" {
		bytes, err := os.ReadFile(cfg.CaCertFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read provided CA cert file %s", cfg.CaCertFile)
		}
		caCert = string(bytes)
	}
	return &KumaInjector{
		cfg:              cfg,
		client:           client,
		converter:        converter,
		defaultAdminPort: envoyAdminPort,
		proxyFactory: containers.NewDataplaneProxyFactory(controlPlaneURL, caCert, envoyAdminPort,
			cfg.SidecarContainer.DataplaneContainer, cfg.BuiltinDNS),
	}, nil
}

type KumaInjector struct {
	cfg              runtime_k8s.Injector
	client           kube_client.Client
	converter        k8s_common.Converter
	proxyFactory     *containers.DataplaneProxyFactory
	defaultAdminPort uint32
}

func (i *KumaInjector) InjectKuma(pod *kube_core.Pod) error {
	ns, err := i.namespaceFor(pod)
	if err != nil {
		return errors.Wrap(err, "could not retrieve namespace for pod")
	}
	if inject, err := i.needInject(pod, ns); err != nil {
		return err
	} else if !inject {
		log.V(1).Info("skip injecting Kuma", "name", pod.Name, "namespace", pod.Namespace)
		return nil
	}
	log.Info("injecting Kuma", "name", pod.GenerateName, "namespace", pod.Namespace)
	// sidecar container
	if pod.Spec.Containers == nil {
		pod.Spec.Containers = []kube_core.Container{}
	}
	container, err := i.NewSidecarContainer(pod, ns)
	if err != nil {
		return err
	}
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	mesh, err := i.meshFor(pod, ns)
	if err != nil {
		return errors.Wrap(err, "could not retrieve mesh for pod")
	}

	// annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	annotations, err := i.NewAnnotations(pod, mesh)
	if err != nil {
		return errors.Wrap(err, "could not generate annotations for pod")
	}
	for key, value := range annotations {
		pod.Annotations[key] = value
	}

	// init container
	if !i.cfg.CNIEnabled {
		if pod.Spec.InitContainers == nil {
			pod.Spec.InitContainers = []kube_core.Container{}
		}
		ic, err := i.NewInitContainer(pod)
		if err != nil {
			return err
		}
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, ic)
	}

	if err := i.overrideHTTPProbes(pod); err != nil {
		return err
	}

	return nil
}

func (i *KumaInjector) needInject(pod *kube_core.Pod, ns *kube_core.Namespace) (bool, error) {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	if i.isInjectionException(pod) {
		log.V(1).Info("pod fulfills exception requirements")
		return false, nil
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == k8s_util.KumaSidecarContainerName {
			log.V(1).Info("pod already has Kuma sidecar")
			return false, nil
		}
	}

	enabled, exist, err := metadata.Annotations(pod.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info(`pod has "kuma.io/sidecar-injection: disabled" label`)
		}
		return enabled, nil
	}

	// support annotations for backwards compatibility
	annotationWarningMsg := "WARNING: you are using kuma.io/sidecar-injection as annotation. Please migrate it to label to have strong guarantee that application can only start with sidecar"
	enabled, exist, err = metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		log.Info(annotationWarningMsg, "pod", pod.Name, "namespace", ns.Name)
		if !enabled {
			log.V(1).Info(`pod has "kuma.io/sidecar-injection: disabled" annotation`)
		}
		return enabled, nil
	}

	enabled, exist, err = metadata.Annotations(ns.Labels).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info(`namespace has "kuma.io/sidecar-injection: disabled" label`)
		}
		return enabled, nil
	}

	// support annotations for backwards compatibility
	enabled, exist, err = metadata.Annotations(ns.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		log.Info(annotationWarningMsg, "namespace", ns.Name)
		if !enabled {
			log.V(1).Info(`namespace has "kuma.io/sidecar-injection: disabled" annotation`)
		}
		return enabled, nil
	}
	return false, nil
}

func (i *KumaInjector) isInjectionException(pod *kube_core.Pod) bool {
	for key, value := range i.cfg.Exceptions.Labels {
		podValue, exist := pod.Labels[key]
		if exist && (value == "*" || value == podValue) {
			return true
		}
	}
	return false
}

func (i *KumaInjector) meshFor(pod *kube_core.Pod, ns *kube_core.Namespace) (*core_mesh.MeshResource, error) {
	meshName := k8s_util.MeshOf(pod, ns)
	mesh := &mesh_k8s.Mesh{}
	if err := i.client.Get(context.Background(), kube_types.NamespacedName{Name: meshName}, mesh); err != nil {
		return nil, err
	}
	meshResource := core_mesh.NewMeshResource()
	if err := i.converter.ToCoreResource(mesh, meshResource); err != nil {
		return nil, err
	}
	return meshResource, nil
}

func (i *KumaInjector) namespaceFor(pod *kube_core.Pod) (*kube_core.Namespace, error) {
	ns := &kube_core.Namespace{}
	nsName := "default"
	if pod.GetNamespace() != "" {
		nsName = pod.GetNamespace()
	}
	if err := i.client.Get(context.Background(), kube_types.NamespacedName{Name: nsName}, ns); err != nil {
		return nil, err
	}
	return ns, nil
}

func (i *KumaInjector) NewSidecarContainer(
	pod *kube_core.Pod,
	ns *kube_core.Namespace,
) (kube_core.Container, error) {
	container, err := i.proxyFactory.NewContainer(pod, ns)
	if err != nil {
		return container, err
	}

	// On versions of Kubernetes prior to v1.15.0
	// ServiceAccount admission plugin is called only once, prior to any mutating web hook.
	// That's why it is a responsibility of every mutating web hook to copy
	// ServiceAccount volume mount into containers it creates.
	container.VolumeMounts = i.NewVolumeMounts(pod)

	container.Name = k8s_util.KumaSidecarContainerName

	return container, nil
}

func (i *KumaInjector) NewVolumeMounts(pod *kube_core.Pod) []kube_core.VolumeMount {
	if tokenVolumeMount := i.FindServiceAccountToken(&pod.Spec); tokenVolumeMount != nil {
		return []kube_core.VolumeMount{*tokenVolumeMount}
	}
	return nil
}

func (i *KumaInjector) FindServiceAccountToken(podSpec *kube_core.PodSpec) *kube_core.VolumeMount {
	for i := range podSpec.Containers {
		for j := range podSpec.Containers[i].VolumeMounts {
			if podSpec.Containers[i].VolumeMounts[j].MountPath == serviceAccountTokenMountPath {
				return &podSpec.Containers[i].VolumeMounts[j]
			}
		}
	}
	// Notice that we consider valid a use case where a ServiceAccount token
	// is not mounted into Pod, e.g. due to Pod.Spec.AutomountServiceAccountToken == false
	// or ServiceAccount.Spec.AutomountServiceAccountToken == false.
	// In that case a side car should still be able to start and join a mesh with disabled mTLS.
	return nil
}

func (i *KumaInjector) NewInitContainer(pod *kube_core.Pod) (kube_core.Container, error) {
	podRedirect, err := tp_k8s.NewPodRedirectForPod(pod)
	if err != nil {
		return kube_core.Container{}, err
	}

	return kube_core.Container{
		Name:            k8s_util.KumaInitContainerName,
		Image:           i.cfg.InitContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Command:         []string{"/usr/bin/kumactl", "install", "transparent-proxy"},
		Args:            podRedirect.AsKumactlCommandLine(),
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  new(int64), // way to get pointer to int64(0)
			RunAsGroup: new(int64),
			Capabilities: &kube_core.Capabilities{
				Add: []kube_core.Capability{
					kube_core.Capability("NET_ADMIN"),
					kube_core.Capability("NET_RAW"),
				},
			},
		},
		Resources: kube_core.ResourceRequirements{
			Limits: kube_core.ResourceList{
				kube_core.ResourceCPU:    *kube_api.NewScaledQuantity(100, kube_api.Milli),
				kube_core.ResourceMemory: *kube_api.NewScaledQuantity(50, kube_api.Mega),
			},
			Requests: kube_core.ResourceList{
				kube_core.ResourceCPU:    *kube_api.NewScaledQuantity(10, kube_api.Milli),
				kube_core.ResourceMemory: *kube_api.NewScaledQuantity(10, kube_api.Mega),
			},
		},
	}, nil
}

func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod, mesh *core_mesh.MeshResource) (map[string]string, error) {
	annotations := map[string]string{
		metadata.KumaMeshAnnotation:                             mesh.GetMeta().GetName(), // either user-defined value or default
		metadata.KumaSidecarInjectedAnnotation:                  fmt.Sprintf("%t", true),
		metadata.KumaSidecarUID:                                 fmt.Sprintf("%d", i.cfg.SidecarContainer.UID),
		metadata.KumaTransparentProxyingAnnotation:              metadata.AnnotationEnabled,
		metadata.KumaTransparentProxyingInboundPortAnnotation:   fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPortInbound),
		metadata.KumaTransparentProxyingInboundPortAnnotationV6: fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPortInboundV6),
		metadata.KumaTransparentProxyingOutboundPortAnnotation:  fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPortOutbound),
	}
	if i.cfg.CNIEnabled {
		annotations[metadata.CNCFNetworkAnnotation] = metadata.KumaCNI
	}

	if i.cfg.BuiltinDNS.Enabled {
		annotations[metadata.KumaBuiltinDNS] = metadata.AnnotationEnabled
		annotations[metadata.KumaBuiltinDNSPort] = strconv.FormatInt(int64(i.cfg.BuiltinDNS.Port), 10)
	}

	if err := setVirtualProbesEnabledAnnotation(annotations, pod, i.cfg); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesAnnotation))
	}
	if err := setVirtualProbesPortAnnotation(annotations, pod, i.cfg); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesPortAnnotation))
	}

	if val, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeInboundPorts); exist {
		annotations[metadata.KumaTrafficExcludeInboundPorts] = val
	} else if len(i.cfg.SidecarTraffic.ExcludeInboundPorts) > 0 {
		annotations[metadata.KumaTrafficExcludeInboundPorts] = portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeInboundPorts)
	}
	if val, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeOutboundPorts); exist {
		annotations[metadata.KumaTrafficExcludeOutboundPorts] = val
	} else if len(i.cfg.SidecarTraffic.ExcludeOutboundPorts) > 0 {
		annotations[metadata.KumaTrafficExcludeOutboundPorts] = portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeOutboundPorts)
	}
	if _, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaEnvoyAdminPort); err != nil {
		return nil, err
	} else if !exist {
		annotations[metadata.KumaEnvoyAdminPort] = fmt.Sprintf("%d", i.defaultAdminPort)
	}
	return annotations, nil
}

func portsToAnnotationValue(ports []uint32) string {
	stringPorts := make([]string, len(ports))
	for i, port := range ports {
		stringPorts[i] = fmt.Sprintf("%d", port)
	}
	return strings.Join(stringPorts, ",")
}
