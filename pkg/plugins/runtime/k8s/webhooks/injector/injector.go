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
	kube_podcmd "k8s.io/kubectl/pkg/cmd/util/podcmd"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	core_config "github.com/kumahq/kuma/pkg/config"
	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/core"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	tproxy_config "github.com/kumahq/kuma/pkg/transparentproxy/config"
	tproxy_consts "github.com/kumahq/kuma/pkg/transparentproxy/consts"
	tproxy_k8s "github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

const (
	// serviceAccountTokenMountPath is a well-known location where Kubernetes mounts a ServiceAccount token.
	serviceAccountTokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount" // #nosec G101 -- this isn't a secret
	mountPathTPBase              = "/tmp/transparent-proxy/base"
	mountPathTPCustom            = "/tmp/transparent-proxy/custom"
	volumeNameTPCustom           = "transparent-proxy-custom"
)

var (
	volumeInitTmp = kube_core.Volume{
		Name: "kuma-init-tmp",
		VolumeSource: kube_core.VolumeSource{
			EmptyDir: &kube_core.EmptyDirVolumeSource{
				SizeLimit: kube_api.NewScaledQuantity(10, kube_api.Mega),
			},
		},
	}
	volumeSidecarTmp = kube_core.Volume{
		Name: "kuma-sidecar-tmp",
		VolumeSource: kube_core.VolumeSource{
			EmptyDir: &kube_core.EmptyDirVolumeSource{
				SizeLimit: kube_api.NewScaledQuantity(10, kube_api.Mega),
			},
		},
	}
	volumeTPBase = kube_core.Volume{
		Name: "transparent-proxy-base",
		VolumeSource: kube_core.VolumeSource{
			DownwardAPI: &kube_core.DownwardAPIVolumeSource{
				Items: []kube_core.DownwardAPIVolumeFile{
					{
						Path: tproxy_consts.KubernetesConfigMapDataKey,
						FieldRef: &kube_core.ObjectFieldSelector{
							FieldPath: fmt.Sprintf(
								"metadata.annotations['%s']",
								metadata.KumaTrafficTransparentProxyConfig,
							),
						},
					},
				},
			},
		},
	}
)

var (
	mountSidecarTmp = kube_core.VolumeMount{Name: volumeSidecarTmp.Name, MountPath: "/tmp"}
	mountInitTmp    = kube_core.VolumeMount{Name: volumeInitTmp.Name, MountPath: "/tmp"}
	mountTPBase     = kube_core.VolumeMount{Name: volumeTPBase.Name, MountPath: mountPathTPBase, ReadOnly: true}
	mountTPCustom   = kube_core.VolumeMount{Name: volumeNameTPCustom, MountPath: mountPathTPCustom, ReadOnly: true}
)

var (
	flagTPConfigBase   = fmt.Sprintf("--transparent-proxy-config=%s/%s", mountPathTPBase, tproxy_consts.KubernetesConfigMapDataKey)
	flagConfigBase     = fmt.Sprintf("--config=%s/%s", mountPathTPBase, tproxy_consts.KubernetesConfigMapDataKey)
	flagTPConfigCustom = fmt.Sprintf("--transparent-proxy-config=%s/%s", mountPathTPCustom, tproxy_consts.KubernetesConfigMapDataKey)
	flagConfigCustom   = fmt.Sprintf("--config=%s/%s", mountPathTPCustom, tproxy_consts.KubernetesConfigMapDataKey)
)

var log = core.Log.WithName("injector")

func New(
	cfg runtime_k8s.Injector,
	controlPlaneURL string,
	client kube_client.Client,
	sidecarContainersEnabled bool,
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
		cfg:                      cfg,
		client:                   client,
		sidecarContainersEnabled: sidecarContainersEnabled,
		converter:                converter,
		defaultAdminPort:         envoyAdminPort,
		proxyFactory: containers.NewDataplaneProxyFactory(
			controlPlaneURL, caCert, envoyAdminPort, cfg.SidecarContainer.DataplaneContainer,
			cfg.BuiltinDNS, cfg.SidecarContainer.WaitForDataplaneReady, sidecarContainersEnabled,
			cfg.VirtualProbesEnabled, cfg.ApplicationProbeProxyPort, cfg.UnifiedResourceNamingEnabled,
			cfg.Spire.Enabled,
		),
		systemNamespace: systemNamespace,
	}, nil
}

type KumaInjector struct {
	cfg                      runtime_k8s.Injector
	client                   kube_client.Client
	sidecarContainersEnabled bool
	converter                k8s_common.Converter
	proxyFactory             *containers.DataplaneProxyFactory
	defaultAdminPort         uint32
	systemNamespace          string
}

func (i *KumaInjector) InjectKuma(ctx context.Context, pod *kube_core.Pod) error {
	logger := log.WithValues("pod", pod.GenerateName, "namespace", pod.Namespace)

	meshName, err := i.preCheck(ctx, pod, logger)
	if meshName == "" || err != nil {
		return err
	}

	logger.Info("injecting Kuma")

	// annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	if _, ok := pod.Annotations[kube_podcmd.DefaultContainerAnnotationName]; len(pod.Spec.Containers) == 1 && !ok {
		pod.Annotations[kube_podcmd.DefaultContainerAnnotationName] = pod.Spec.Containers[0].Name
	}

	sidecarPatches, initPatches, err := i.loadContainerPatches(ctx, logger, pod)
	if err != nil {
		return err
	}
	container, err := i.NewSidecarContainer(pod, meshName)
	if err != nil {
		return err
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, volumeInitTmp, volumeSidecarTmp)

	if enabled, _, err := metadata.Annotations(pod.Annotations).GetEnabledWithDefault(i.cfg.Spire.Enabled, metadata.KumaSpireSupport); err != nil {
		return errors.Wrapf(err, "getting %s annotation failed", metadata.KumaSpireSupport)
	} else if enabled {
		pod.Spec.Volumes = append(pod.Spec.Volumes, kube_core.Volume{
			Name: "kuma-spire-agent-socket",
			VolumeSource: kube_core.VolumeSource{
				CSI: &kube_core.CSIVolumeSource{
					Driver:   "csi.spiffe.io",
					ReadOnly: pointer.To(true),
				},
			},
		})
	}

	patchedContainer, err := i.applyCustomPatches(logger, container, sidecarPatches)
	if err != nil {
		return err
	}

	var injectedInitContainer *kube_core.Container

	if i.cfg.TransparentProxyConfigMapName != "" {
		tpCfgBase, err := i.getTransparentProxyConfigMap(ctx, i.cfg.TransparentProxyConfigMapName, i.systemNamespace, logger)
		if err != nil {
			return errors.Wrap(err, "could not retrieve transparent proxy configuration")
		}

		tpCfg, err := tproxy_k8s.ConfigForKubernetes(tpCfgBase, i.cfg, pod.Annotations, logger)
		if err != nil {
			return err
		}

		pod.Spec.Volumes = append(pod.Spec.Volumes, volumeTPBase)

		if v := pod.Annotations[metadata.KumaTrafficTransparentProxyConfigMapName]; v != "" {
			pod.Spec.Volumes = append(
				pod.Spec.Volumes,
				kube_core.Volume{
					Name: volumeNameTPCustom,
					VolumeSource: kube_core.VolumeSource{
						ConfigMap: &kube_core.ConfigMapVolumeSource{
							LocalObjectReference: kube_core.LocalObjectReference{
								Name: v,
							},
						},
					},
				},
			)
		}

		annotations, err := tproxy_k8s.ConfigToAnnotations(tpCfg, i.cfg, pod.Annotations, i.defaultAdminPort)
		if err != nil {
			return errors.Wrap(err, "could not generate annotations for pod")
		}

		for key, value := range annotations {
			pod.Annotations[key] = value
		}

		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}
		pod.Labels[metadata.KumaMeshLabel] = meshName

		switch {
		case !tpCfg.CNIMode:
			initContainer := i.NewInitContainer(nil, pod.Annotations)
			injected, err := i.applyCustomPatches(logger, initContainer, initPatches)
			if err != nil {
				return err
			}
			injectedInitContainer = &injected
		case tpCfg.Redirect.Inbound.Enabled:
			injected, err := i.applyCustomPatches(logger, i.NewValidationContainer(pod), initPatches)
			if err != nil {
				return err
			}
			injectedInitContainer = &injected
		}
	} else { // this is legacy and deprecated - will be removed soon
		annotations, err := i.NewAnnotations(pod, logger)
		if err != nil {
			return errors.Wrap(err, "could not generate annotations for pod")
		}

		for key, value := range annotations {
			pod.Annotations[key] = value
		}

		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}
		pod.Labels[metadata.KumaMeshLabel] = meshName

		podRedirect, err := tproxy_k8s.NewPodRedirectFromAnnotations(pod.Annotations)
		if err != nil {
			return err
		}

		if !i.cfg.CNIEnabled {
			initContainer := i.NewInitContainer(podRedirect.AsKumactlCommandLine(), pod.Annotations)
			injected, err := i.applyCustomPatches(logger, initContainer, initPatches)
			if err != nil {
				return err
			}
			injectedInitContainer = &injected
		} else if podRedirect.RedirectInbound {
			injected, err := i.applyCustomPatches(logger, i.NewValidationContainer(pod), initPatches)
			if err != nil {
				return err
			}
			injectedInitContainer = &injected
		}
	}

	if i.cfg.EBPF.Enabled {
		pod.Spec.Volumes = append(pod.Spec.Volumes, kube_core.Volume{
			Name: "sys-fs-cgroup",
			VolumeSource: kube_core.VolumeSource{
				HostPath: &kube_core.HostPathVolumeSource{
					Path: i.cfg.EBPF.CgroupPath,
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

	initFirst, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaInitFirst)
	if err != nil {
		return err
	}

	var prependInitContainers []kube_core.Container
	var appendInitContainers []kube_core.Container

	if injectedInitContainer != nil {
		if initFirst || i.sidecarContainersEnabled {
			prependInitContainers = append(prependInitContainers, *injectedInitContainer)
		} else {
			appendInitContainers = append(appendInitContainers, *injectedInitContainer)
		}
	}

	if i.sidecarContainersEnabled {
		// inject sidecar after init
		patchedContainer.RestartPolicy = pointer.To(kube_core.ContainerRestartPolicyAlways)
		patchedContainer.Lifecycle = &kube_core.Lifecycle{
			PreStop: &kube_core.LifecycleHandler{
				Exec: &kube_core.ExecAction{
					Command: []string{"killall", "-USR2", "kuma-dp"},
				},
			},
		}
		prependInitContainers = append(prependInitContainers, patchedContainer)
	} else {
		// inject sidecar as first container
		pod.Spec.Containers = append([]kube_core.Container{patchedContainer}, pod.Spec.Containers...)
	}

	pod.Spec.InitContainers = append(append(prependInitContainers, pod.Spec.InitContainers...), appendInitContainers...)

	disabledAppProbeProxy, err := probes.ApplicationProbeProxyDisabled(pod)
	if err != nil {
		return err
	}

	if disabledAppProbeProxy {
		if err := i.overrideHTTPProbes(pod); err != nil {
			return err
		}
	} else {
		if err := probes.SetupAppProbeProxies(pod, log); err != nil {
			return err
		}
	}

	return nil
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
) (namedContainerPatches, namedContainerPatches, error) {
	patchNames := i.cfg.ContainerPatches
	otherPatches, _ := metadata.Annotations(pod.Annotations).GetList(metadata.KumaContainerPatches)
	patchNames = append(patchNames, otherPatches...)

	var missingPatches []string

	var initPatches namedContainerPatches
	var sidecarPatches namedContainerPatches
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

func (i *KumaInjector) getTransparentProxyConfigMap(
	ctx context.Context,
	name string,
	namespace string,
	logger logr.Logger,
) (tproxy_config.Config, error) {
	var err error
	defer func() {
		if err != nil {
			logger.V(1).Info(
				"[WARNING]: unable to retrieve transparent proxy configuration from ConfigMap; applying default configuration",
				"configMapName", name,
				"configMapNamespace", namespace,
				"error", err,
			)
		}
	}()

	cfg := tproxy_config.DefaultConfig()
	loader := core_config.NewLoader(&cfg)
	namespacedName := kube_types.NamespacedName{Name: name, Namespace: namespace}

	var cm kube_core.ConfigMap
	if err = i.client.Get(ctx, namespacedName, &cm); err != nil {
		return tproxy_config.Config{}, err
	}

	if c := cm.Data[tproxy_consts.KubernetesConfigMapDataKey]; c != "" {
		if err = loader.LoadBytes([]byte(c)); err != nil {
			return tproxy_config.Config{}, err
		}

		return cfg, nil
	}

	err = errors.Errorf(
		"key '%s' is missing or empty",
		tproxy_consts.KubernetesConfigMapDataKey,
	)

	return tproxy_config.Config{}, err
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

	if i.cfg.TransparentProxyConfigMapName != "" {
		container.Args = append(container.Args, flagTPConfigBase)
		if v := pod.Annotations[metadata.KumaTrafficTransparentProxyConfigMapName]; v != "" {
			container.Args = append(container.Args, flagTPConfigCustom)
		}
	}

	container.Name = k8s_util.KumaSidecarContainerName
	container.SecurityContext.ReadOnlyRootFilesystem = pointer.To(true)
	container.SecurityContext.AllowPrivilegeEscalation = pointer.To(false)

	return container, nil
}

func (i *KumaInjector) NewVolumeMounts(pod *kube_core.Pod) ([]kube_core.VolumeMount, error) {
	out := []kube_core.VolumeMount{mountSidecarTmp}

	if i.cfg.TransparentProxyConfigMapName != "" {
		out = append(out, mountTPBase)
		if v := pod.Annotations[metadata.KumaTrafficTransparentProxyConfigMapName]; v != "" {
			out = append(out, mountTPCustom)
		}
	}

	// If the user specifies a volume containing a service account token, we will mount and use that.
	if volumeName, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaSidecarTokenVolumeAnnotation); exists {
		// Ensure the volume specified exists on the pod spec, otherwise error.
		for _, v := range pod.Spec.Volumes {
			if v.Name == volumeName {
				return append(
					out,
					kube_core.VolumeMount{
						Name:      volumeName,
						MountPath: serviceAccountTokenMountPath,
						ReadOnly:  true,
					},
				), nil
			}
		}

		return nil, errors.Errorf("volume (%s) specified for %s but volume does not exist in pod spec", volumeName, metadata.KumaSidecarTokenVolumeAnnotation)
	}

	if enabled, _, err := metadata.Annotations(pod.Annotations).GetEnabledWithDefault(i.cfg.Spire.Enabled, metadata.KumaSpireSupport); err != nil {
		return nil, errors.Wrapf(err, "getting %s annotation failed", metadata.KumaSpireSupport)
	} else if enabled {
		out = append(out, kube_core.VolumeMount{Name: "kuma-spire-agent-socket", MountPath: i.cfg.Spire.MountPath, ReadOnly: true})
	}

	// If not specified with the above annotation, instead query each container in the pod to find a
	// service account token to mount.
	if tokenVolumeMount := i.FindServiceAccountToken(&pod.Spec); tokenVolumeMount != nil {
		return append(out, *tokenVolumeMount), nil
	}

	return out, nil
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

func (i *KumaInjector) NewInitContainer(args []string, annotations map[string]string) kube_core.Container {
	mounts := []kube_core.VolumeMount{mountInitTmp}

	if i.cfg.TransparentProxyConfigMapName != "" {
		mounts = append(mounts, mountTPBase)
		args = append(args, flagConfigBase)
		if v := annotations[metadata.KumaTrafficTransparentProxyConfigMapName]; v != "" {
			mounts = append(mounts, mountTPCustom)
			args = append(args, flagConfigCustom)
		}
	}

	container := kube_core.Container{
		Name:            k8s_util.KumaInitContainerName,
		Image:           i.cfg.InitContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Command:         []string{"/usr/bin/kumactl", "install", "transparent-proxy"},
		Args:            args,
		Env: []kube_core.EnvVar{
			// iptables needs this lock file to be writable:
			// source: https://git.netfilter.org/iptables/tree/iptables/xshared.c?h=v1.8.7#n258
			{
				Name:  "XTABLES_LOCKFILE",
				Value: "/tmp/xtables.lock",
			},
		},
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  new(int64), // way to get pointer to int64(0)
			RunAsGroup: new(int64),
			Capabilities: &kube_core.Capabilities{
				Add: []kube_core.Capability{
					"NET_ADMIN",
					"NET_RAW",
				},
				Drop: []kube_core.Capability{
					"ALL",
				},
			},
			ReadOnlyRootFilesystem: pointer.To(true),
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
		VolumeMounts: mounts,
	}

	if i.cfg.EBPF.Enabled {
		bidirectional := kube_core.MountPropagationBidirectional

		container.SecurityContext.Capabilities = &kube_core.Capabilities{}
		container.SecurityContext.Privileged = pointer.To(true)

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

		container.Resources.Limits = kube_core.ResourceList{
			kube_core.ResourceCPU:    *kube_api.NewScaledQuantity(100, kube_api.Milli),
			kube_core.ResourceMemory: *kube_api.NewScaledQuantity(80, kube_api.Mega),
		}

		container.VolumeMounts = append(container.VolumeMounts,
			kube_core.VolumeMount{Name: "sys-fs-cgroup", MountPath: i.cfg.EBPF.CgroupPath},
			kube_core.VolumeMount{Name: "bpf-fs", MountPath: i.cfg.EBPF.BPFFSPath, MountPropagation: &bidirectional},
		)
	}

	return container
}

func (i *KumaInjector) NewValidationContainer(pod *kube_core.Pod) kube_core.Container {
	annotations := metadata.Annotations(pod.Annotations)
	mounts := []kube_core.VolumeMount{mountSidecarTmp}
	args := []string{"--config-file=/tmp/.kumactl"}

	if i.cfg.TransparentProxyConfigMapName != "" {
		mounts = append(mounts, mountTPBase)
		args = append(args, flagTPConfigBase)
		if v := pod.Annotations[metadata.KumaTrafficTransparentProxyConfigMapName]; v != "" {
			mounts = append(mounts, mountTPCustom)
			args = append(args, flagTPConfigCustom)
		}
	} else {
		ipFamilyMode := metadata.IpFamilyModeDualStack
		if v, _ := annotations.GetString(metadata.KumaTransparentProxyingIPFamilyMode); v != "" {
			ipFamilyMode = v
		}

		port := fmt.Sprintf("%d", tproxy_config.DefaultConfig().Redirect.Inbound.Port)
		if v, ok, err := annotations.GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotation); ok && err == nil {
			port = fmt.Sprintf("%d", v)
		}

		args = append(
			args,
			fmt.Sprintf("--ip-family-mode=%s", ipFamilyMode),
			fmt.Sprintf("--validation-server-port=%s", port),
		)
	}

	return kube_core.Container{
		Name:            k8s_util.KumaCniValidationContainerName,
		Image:           i.cfg.InitContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Command:         []string{"/usr/bin/kumactl", "install", "transparent-proxy-validator"},
		Args:            args,
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  &i.cfg.SidecarContainer.UID,
			RunAsGroup: &i.cfg.SidecarContainer.GID,
			Capabilities: &kube_core.Capabilities{
				Drop: []kube_core.Capability{
					"ALL",
				},
			},
			ReadOnlyRootFilesystem:   pointer.To(true),
			AllowPrivilegeEscalation: pointer.To(false),
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
		VolumeMounts: mounts,
	}
}

// Deprecated
func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod, logger logr.Logger) (map[string]string, error) {
	portOutbound := i.cfg.SidecarContainer.RedirectPortOutbound
	portInbound := i.cfg.SidecarContainer.RedirectPortInbound

	result := map[string]string{
		metadata.KumaSidecarInjectedAnnotation:                 metadata.AnnotationTrue,
		metadata.KumaTransparentProxyingAnnotation:             metadata.AnnotationEnabled,
		metadata.KumaSidecarUID:                                fmt.Sprintf("%d", i.cfg.SidecarContainer.UID),
		metadata.KumaTransparentProxyingOutboundPortAnnotation: fmt.Sprintf("%d", portOutbound),
		metadata.KumaTransparentProxyingInboundPortAnnotation:  fmt.Sprintf("%d", portInbound),
	}

	if i.cfg.CNIEnabled {
		result[metadata.CNCFNetworkAnnotation] = metadata.KumaCNI
	}

	annotations := metadata.Annotations(pod.Annotations)

	if v, exists, _ := annotations.GetEnabled(
		metadata.KumaTransparentProxyingAnnotation,
	); exists && !v {
		logger.Info(fmt.Sprintf(
			"[WARNING]: cannot change the value of annotation %s as the transparent proxy must be enabled in Kubernetes",
			metadata.KumaTransparentProxyingAnnotation,
		))
	}

	if v, exists, _ := annotations.GetUint32(
		metadata.KumaTransparentProxyingInboundPortAnnotation,
	); exists && v != portInbound {
		logger.Info(fmt.Sprintf(
			"[WARNING]: cannot change the value of annotation %s on a per pod basis. The global setting will be used",
			metadata.KumaTransparentProxyingInboundPortAnnotation,
		))
	}

	if v, exists, _ := annotations.GetUint32(
		metadata.KumaTransparentProxyingOutboundPortAnnotation,
	); exists && v != portOutbound {
		logger.Info(fmt.Sprintf(
			"[WARNING]: cannot change the value of annotation %s on a per pod basis. The global setting will be used",
			metadata.KumaTransparentProxyingOutboundPortAnnotation,
		))
	}

	if ebpfEnabled, _, err := annotations.GetEnabledWithDefault(
		i.cfg.EBPF.Enabled,
		metadata.KumaTransparentProxyingEbpf,
	); err != nil {
		return nil, errors.Wrapf(err, "getting %s annotation failed", metadata.KumaTransparentProxyingEbpf)
	} else if ebpfEnabled {
		result[metadata.KumaTransparentProxyingEbpf] = metadata.AnnotationEnabled

		if v, _ := annotations.GetStringWithDefault(
			i.cfg.EBPF.BPFFSPath,
			metadata.KumaTransparentProxyingEbpfBPFFSPath,
		); v != "" {
			result[metadata.KumaTransparentProxyingEbpfBPFFSPath] = v
		}

		if v, _ := annotations.GetStringWithDefault(
			i.cfg.EBPF.CgroupPath,
			metadata.KumaTransparentProxyingEbpfCgroupPath,
		); v != "" {
			result[metadata.KumaTransparentProxyingEbpfCgroupPath] = v
		}

		if v, _ := annotations.GetStringWithDefault(
			i.cfg.EBPF.TCAttachIface,
			metadata.KumaTransparentProxyingEbpfTCAttachIface,
		); v != "" {
			result[metadata.KumaTransparentProxyingEbpfTCAttachIface] = v
		}

		if v, _ := annotations.GetStringWithDefault(
			i.cfg.EBPF.ProgramsSourcePath,
			metadata.KumaTransparentProxyingEbpfProgramsSourcePath,
		); v != "" {
			result[metadata.KumaTransparentProxyingEbpfProgramsSourcePath] = v
		}

		if v, _ := annotations.GetString(
			i.cfg.EBPF.InstanceIPEnvVarName,
			metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName,
		); v != "" {
			result[metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName] = v
		}
	}

	if dnsEnabled, _, err := annotations.GetEnabledWithDefault(
		i.cfg.BuiltinDNS.Enabled,
		metadata.KumaBuiltinDNS,
	); err != nil {
		return nil, err
	} else if dnsEnabled {
		result[metadata.KumaBuiltinDNS] = metadata.AnnotationEnabled

		if v, _, err := annotations.GetUint32WithDefault(
			i.cfg.BuiltinDNS.Port,
			metadata.KumaBuiltinDNSPort,
		); err != nil {
			return nil, err
		} else {
			result[metadata.KumaBuiltinDNSPort] = fmt.Sprintf("%d", v)
		}

		if v, _, err := annotations.GetEnabledWithDefault(
			i.cfg.BuiltinDNS.Logging,
			metadata.KumaBuiltinDNSLogging,
		); err != nil {
			return nil, err
		} else {
			result[metadata.KumaBuiltinDNSLogging] = strconv.FormatBool(v)
		}
	}

	if err := probes.SetVirtualProbesEnabledAnnotation(
		result,
		pod.Annotations,
		i.cfg.VirtualProbesEnabled,
	); err != nil {
		return nil, errors.Wrap(
			err,
			fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesAnnotation),
		)
	}

	if err := setVirtualProbesPortAnnotation(result, pod, i.cfg); err != nil {
		return nil, errors.Wrap(
			err,
			fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesPortAnnotation),
		)
	}

	if err := probes.SetApplicationProbeProxyPortAnnotation(
		result,
		pod.Annotations,
		i.cfg.ApplicationProbeProxyPort,
	); err != nil {
		return nil, errors.Wrap(
			err,
			fmt.Sprintf("unable to set %s", metadata.KumaApplicationProbeProxyPortAnnotation),
		)
	}

	if v, _ := annotations.GetStringWithDefault(
		portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeInboundPorts),
		metadata.KumaTrafficExcludeInboundPorts,
	); v != "" {
		result[metadata.KumaTrafficExcludeInboundPorts] = v
	}

	if v, _ := annotations.GetStringWithDefault(
		portsToAnnotationValue(i.cfg.SidecarTraffic.ExcludeOutboundPorts),
		metadata.KumaTrafficExcludeOutboundPorts,
	); v != "" {
		result[metadata.KumaTrafficExcludeOutboundPorts] = v
	}

	if v, _ := annotations.GetStringWithDefault(
		i.cfg.SidecarContainer.IpFamilyMode,
		metadata.KumaTransparentProxyingIPFamilyMode,
	); v != "" {
		result[metadata.KumaTransparentProxyingIPFamilyMode] = v
	} else {
		result[metadata.KumaTransparentProxyingIPFamilyMode] = string(tproxy_config.IPFamilyModeDualStack)
	}

	if v, _, err := annotations.GetUint32WithDefault(
		i.defaultAdminPort,
		metadata.KumaEnvoyAdminPort,
	); err != nil {
		return nil, err
	} else {
		result[metadata.KumaEnvoyAdminPort] = fmt.Sprintf("%d", v)
	}

	if v, _ := annotations.GetStringWithDefault(
		strings.Join(i.cfg.SidecarTraffic.ExcludeOutboundIPs, ","),
		metadata.KumaTrafficExcludeOutboundIPs,
	); v != "" {
		result[metadata.KumaTrafficExcludeOutboundIPs] = v
	}

	if v, _ := annotations.GetStringWithDefault(
		strings.Join(i.cfg.SidecarTraffic.ExcludeInboundIPs, ","),
		metadata.KumaTrafficExcludeInboundIPs,
	); v != "" {
		result[metadata.KumaTrafficExcludeInboundIPs] = v
	}

	return result, nil
}

// Deprecated
func portsToAnnotationValue(ports []uint32) string {
	stringPorts := make([]string, len(ports))
	for i, port := range ports {
		stringPorts[i] = fmt.Sprintf("%d", port)
	}
	return strings.Join(stringPorts, ",")
}
