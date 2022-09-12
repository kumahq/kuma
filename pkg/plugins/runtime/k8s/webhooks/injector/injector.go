package injector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/core"
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
	systemNamespace string,
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
		systemNamespace: systemNamespace,
	}, nil
}

type KumaInjector struct {
	cfg              runtime_k8s.Injector
	client           kube_client.Client
	converter        k8s_common.Converter
	proxyFactory     *containers.DataplaneProxyFactory
	defaultAdminPort uint32
	systemNamespace  string
}

func (i *KumaInjector) InjectKuma(ctx context.Context, pod *kube_core.Pod) error {
	ns, err := i.namespaceFor(ctx, pod)
	if err != nil {
		return errors.Wrap(err, "could not retrieve namespace for pod")
	}
	logger := log.WithValues("pod", pod.GenerateName, "namespace", pod.Namespace)
	// Log deprecated annotations
	for _, d := range metadata.PodAnnotationDeprecations {
		if _, exists := pod.Annotations[d.Key]; exists {
			logger.Info("WARNING: using deprecated pod annotation", "key", d.Key, "message", d.Message)
		}
	}
	if inject, err := i.needInject(pod, ns); err != nil {
		return err
	} else if !inject {
		logger.V(1).Info("skip injecting Kuma")
		return nil
	}
	meshName := k8s_util.MeshOf(pod, ns)
	logger = logger.WithValues("mesh", meshName)
	// Check mesh exists
	if err := i.client.Get(ctx, kube_types.NamespacedName{Name: meshName}, &mesh_k8s.Mesh{}); err != nil {
		return err
	}

	logger.Info("injecting Kuma")

	sidecarPatches, initPatches, err := i.loadContainerPatches(ctx, logger, pod)
	if err != nil {
		return err
	}
	container, err := i.NewSidecarContainer(pod, meshName)
	if err != nil {
		return err
	}
	patchedContainer, err := i.applyCustomPatches(logger, container, sidecarPatches)
	if err != nil {
		return err
	}
	pod.Spec.Containers = append(pod.Spec.Containers, patchedContainer)

	// annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	annotations, err := i.NewAnnotations(pod, meshName, logger)
	if err != nil {
		return errors.Wrap(err, "could not generate annotations for pod")
	}
	for key, value := range annotations {
		pod.Annotations[key] = value
	}

	if i.cfg.EBPF.Enabled {
		pod.Spec.Volumes = append(pod.Spec.Volumes, kube_core.Volume{
			Name: "sys-fs-cgroup",
			VolumeSource: kube_core.VolumeSource{
				HostPath: &kube_core.HostPathVolumeSource{
					Path: "/sys/fs/cgroup",
				},
			},
		}, kube_core.Volume{
			Name: "bpf-fs",
			VolumeSource: kube_core.VolumeSource{
				HostPath: &kube_core.HostPathVolumeSource{
					Path: i.cfg.EBPF.BPFFSPath,
				},
			},
		})
	}

	// init container
	if !i.cfg.CNIEnabled {
		ic, err := i.NewInitContainer(pod)
		if err != nil {
			return err
		}
		patchedIc, err := i.applyCustomPatches(logger, ic, initPatches)
		if err != nil {
			return err
		}
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, patchedIc)
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
	// https://github.com/kumahq/kuma/issues/4005
	enabled, exist, err = metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
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
	// https://github.com/kumahq/kuma/issues/4005
	enabled, exist, err = metadata.Annotations(ns.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
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

type namedContainerPatches struct {
	names   []string
	patches []mesh_k8s.JsonPatchBlock
}

// loadContainerPatches loads the ContainerPatch CRDs associated with the given pod, divides out
// the sidecar and init patches, and concatenates each type into its own list for return.
func (i *KumaInjector) loadContainerPatches(
	ctx context.Context,
	logger logr.Logger,
	pod *kube_core.Pod,
) (sidecarPatches namedContainerPatches, initPatches namedContainerPatches, err error) {
	patchNames := i.cfg.ContainerPatches
	otherPatches, _ := metadata.Annotations(pod.Annotations).GetList(metadata.KumaContainerPatches)
	patchNames = append(patchNames, otherPatches...)

	var missingPatches []string

	for _, patchName := range patchNames {
		containerPatch := &mesh_k8s.ContainerPatch{}
		if err := i.client.Get(ctx, kube_types.NamespacedName{Namespace: i.systemNamespace, Name: patchName}, containerPatch); err != nil {
			if kube_errors.IsNotFound(err) {
				missingPatches = append(missingPatches, patchName)
				continue
			}

			logger.Error(err, "could not get ContainerPatch", "name", patchName)

			return namedContainerPatches{}, namedContainerPatches{}, err
		}
		if len(containerPatch.Spec.SidecarPatch) > 0 {
			sidecarPatches.names = append(sidecarPatches.names, patchName)
			sidecarPatches.patches = append(sidecarPatches.patches, containerPatch.Spec.SidecarPatch...)
		}
		if len(containerPatch.Spec.InitPatch) > 0 {
			initPatches.names = append(initPatches.names, patchName)
			initPatches.patches = append(initPatches.patches, containerPatch.Spec.InitPatch...)
		}
	}

	if len(missingPatches) > 0 {
		err := fmt.Errorf(
			"it appears some expected container patches are missing: %q",
			missingPatches,
		)

		logger.Error(err,
			"loading container patches failed",
			"expected", patchNames,
			"missing", missingPatches,
		)

		return namedContainerPatches{}, namedContainerPatches{}, err
	}

	return sidecarPatches, initPatches, nil
}

// applyCustomPatches applies the block of patches to the given container and returns a new,
// patched container. If patch list is empty, the same unaltered container is returned.
func (i *KumaInjector) applyCustomPatches(
	logger logr.Logger,
	container kube_core.Container,
	patches namedContainerPatches,
) (kube_core.Container, error) {
	if len(patches.patches) == 0 {
		return container, nil
	}

	var patchedContainer kube_core.Container
	containerJson, err := json.Marshal(&container)
	if err != nil {
		return kube_core.Container{}, err
	}
	logger.Info("applying a patches to the container", "patches", patches.names)

	patchOptions := jsonpatch.NewApplyOptions()
	patchOptions.EnsurePathExistsOnAdd = true

	containerJson, err = mesh_k8s.ToJsonPatch(patches.patches).ApplyWithOptions(containerJson, patchOptions)
	if err != nil {
		return kube_core.Container{}, errors.Wrapf(err, "could not apply patches %q", patches.names)
	}

	err = json.Unmarshal(containerJson, &patchedContainer)
	if err != nil {
		return kube_core.Container{}, err
	}
	return patchedContainer, nil
}

func (i *KumaInjector) namespaceFor(ctx context.Context, pod *kube_core.Pod) (*kube_core.Namespace, error) {
	ns := &kube_core.Namespace{}
	nsName := "default"
	if pod.GetNamespace() != "" {
		nsName = pod.GetNamespace()
	}
	if err := i.client.Get(ctx, kube_types.NamespacedName{Name: nsName}, ns); err != nil {
		return nil, err
	}
	return ns, nil
}

func (i *KumaInjector) NewSidecarContainer(
	pod *kube_core.Pod,
	mesh string,
) (kube_core.Container, error) {
	container, err := i.proxyFactory.NewContainer(pod, mesh)
	if err != nil {
		return container, err
	}

	// On versions of Kubernetes prior to v1.15.0
	// ServiceAccount admission plugin is called only once, prior to any mutating web hook.
	// That's why it is a responsibility of every mutating web hook to copy
	// ServiceAccount volume mount into containers it creates.
	container.VolumeMounts, err = i.NewVolumeMounts(pod)
	if err != nil {
		return container, err
	}

	container.Name = k8s_util.KumaSidecarContainerName

	return container, nil
}
func (i *KumaInjector) NewVolumeMounts(pod *kube_core.Pod) ([]kube_core.VolumeMount, error) {
	// If the user specifies a volume containing a service account token, we will mount and use that.
	if volumeName, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaSidecarTokenVolumeAnnotation); exists {
		// Ensure the volume specified exists on the pod spec, otherwise error.
		for _, v := range pod.Spec.Volumes {
			if v.Name == volumeName {
				return []kube_core.VolumeMount{{
					Name:      volumeName,
					ReadOnly:  true,
					MountPath: serviceAccountTokenMountPath,
				}}, nil
			}
		}
		return nil, errors.Errorf("volume (%s) specified for %s but volume does not exist in pod spec", volumeName, metadata.KumaSidecarTokenVolumeAnnotation)
	}

	// If not specified with the above annotation, instead query each container in the pod to find a
	// service account token to mount.
	if tokenVolumeMount := i.FindServiceAccountToken(&pod.Spec); tokenVolumeMount != nil {
		return []kube_core.VolumeMount{*tokenVolumeMount}, nil
	}
	return nil, nil
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
	// In that case a sidecar should still be able to start and join a mesh with disabled mTLS.
	return nil
}

func (i *KumaInjector) NewInitContainer(pod *kube_core.Pod) (kube_core.Container, error) {
	podRedirect, err := tp_k8s.NewPodRedirectForPod(i.cfg.TransparentProxyV2, pod)
	if err != nil {
		return kube_core.Container{}, err
	}

	container := kube_core.Container{
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
					"NET_ADMIN",
					"NET_RAW",
				},
			},
		},
		Resources: kube_core.ResourceRequirements{
			Limits: kube_core.ResourceList{
				kube_core.ResourceCPU:    *kube_api.NewScaledQuantity(100, kube_api.Milli),
				kube_core.ResourceMemory: *kube_api.NewScaledQuantity(50, kube_api.Mega),
			},
			Requests: kube_core.ResourceList{
				kube_core.ResourceCPU:    *kube_api.NewScaledQuantity(20, kube_api.Milli),
				kube_core.ResourceMemory: *kube_api.NewScaledQuantity(20, kube_api.Mega),
			},
		},
	}

	if i.cfg.EBPF.Enabled {
		// container.SecurityContext.Privileged expects to have a reference
		// to the bool value
		tru := true
		bidirectional := kube_core.MountPropagationBidirectional

		container.SecurityContext.Capabilities = &kube_core.Capabilities{}
		container.SecurityContext.Privileged = &tru

		container.Env = []kube_core.EnvVar{
			{
				Name: i.cfg.EBPF.InstanceIPEnvVarName,
				ValueFrom: &kube_core.EnvVarSource{
					FieldRef: &kube_core.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		}

		container.VolumeMounts = []kube_core.VolumeMount{
			{Name: "sys-fs-cgroup", MountPath: "/sys/fs/cgroup"},
			{Name: "bpf-fs", MountPath: i.cfg.EBPF.BPFFSPath, MountPropagation: &bidirectional},
		}
	}

	return container, nil
}

func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod, mesh string, logger logr.Logger) (map[string]string, error) {
	annotations := map[string]string{
		metadata.KumaMeshAnnotation:                             mesh, // either user-defined value or default
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

	podAnnotations := metadata.Annotations(pod.Annotations)

	ebpfEnabled, _, err := podAnnotations.GetEnabledWithDefault(i.cfg.EBPF.Enabled, metadata.KumaTransparentProxyingEbpf)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s annotation failed", metadata.KumaTransparentProxyingEbpf)
	}
	annotations[metadata.KumaTransparentProxyingEbpf] = metadata.BoolToEnabled(ebpfEnabled)

	if ebpfEnabled {
		annotations[metadata.KumaTransparentProxyingEbpfBPFFSPath], _ = podAnnotations.GetStringWithDefault(i.cfg.EBPF.BPFFSPath, metadata.KumaTransparentProxyingEbpfBPFFSPath)
		annotations[metadata.KumaTransparentProxyingEbpfProgramsSourcePath], _ = podAnnotations.GetStringWithDefault(i.cfg.EBPF.ProgramsSourcePath, metadata.KumaTransparentProxyingEbpfProgramsSourcePath)
		if value, exists := podAnnotations.GetString(i.cfg.EBPF.InstanceIPEnvVarName, metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName); exists {
			annotations[metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName] = value
		}

		// ebpf works only with experimental transparent proxy engine, so instead of
		// failing when no annotation enabling it is present (bad user experience)
		// we implicitly add it and set to true
		enabled, exists, err := podAnnotations.GetEnabled(metadata.KumaTransparentProxyingExperimentalEngine)
		if err != nil {
			return nil, errors.Wrapf(err, "getting %s annotation failed", metadata.KumaTransparentProxyingExperimentalEngine)
		}
		if !exists || !enabled {
			logger.V(1).Info(fmt.Sprintf("missing %s annotation which has to be %s for ebpf to work. The annotation will be implicitly added",
				metadata.KumaTransparentProxyingExperimentalEngine, metadata.AnnotationEnabled),
				"annotation", metadata.KumaTransparentProxyingExperimentalEngine)

			annotations[metadata.KumaTransparentProxyingExperimentalEngine] = metadata.AnnotationEnabled
		}
	}

	enabled, _, err := podAnnotations.GetEnabledWithDefault(i.cfg.BuiltinDNS.Enabled, metadata.KumaBuiltinDNSDeprecated, metadata.KumaBuiltinDNS)
	if err != nil {
		return nil, err
	}
	port, _, err := podAnnotations.GetUint32WithDefault(i.cfg.BuiltinDNS.Port, metadata.KumaBuiltinDNSPortDeprecated, metadata.KumaBuiltinDNSPort)
	if err != nil {
		return nil, err
	}

	if enabled {
		portVal := strconv.Itoa(int(port))
		annotations[metadata.KumaBuiltinDNS] = metadata.AnnotationEnabled
		annotations[metadata.KumaBuiltinDNSPort] = portVal
		// TODO remove deprecation issue
		annotations[metadata.KumaBuiltinDNSDeprecated] = metadata.AnnotationEnabled
		annotations[metadata.KumaBuiltinDNSPortDeprecated] = portVal
	}

	if err := setVirtualProbesEnabledAnnotation(annotations, pod, i.cfg); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesAnnotation))
	}
	if err := setVirtualProbesPortAnnotation(annotations, pod, i.cfg); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesPortAnnotation))
	}

	if val, _ := metadata.Annotations(pod.Annotations).GetStringWithDefault(portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeInboundPorts), metadata.KumaTrafficExcludeInboundPorts); val != "" {
		annotations[metadata.KumaTrafficExcludeInboundPorts] = val
	}
	if val, _ := metadata.Annotations(pod.Annotations).GetStringWithDefault(portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeOutboundPorts), metadata.KumaTrafficExcludeOutboundPorts); val != "" {
		annotations[metadata.KumaTrafficExcludeOutboundPorts] = val
	}
	val, _, err := metadata.Annotations(pod.Annotations).GetUint32WithDefault(i.defaultAdminPort, metadata.KumaEnvoyAdminPort)
	if err != nil {
		return nil, err
	}
	annotations[metadata.KumaEnvoyAdminPort] = fmt.Sprintf("%d", val)
	return annotations, nil
}

func portsToAnnotationValue(ports []uint32) string {
	stringPorts := make([]string, len(ports))
	for i, port := range ports {
		stringPorts[i] = fmt.Sprintf("%d", port)
	}
	return strings.Join(stringPorts, ",")
}
