package injector

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	tp_k8s "github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"

	"github.com/pkg/errors"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"

	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"

	kube_core "k8s.io/api/core/v1"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// serviceAccountTokenMountPath is a well-known location where Kubernetes mounts a ServiceAccount token.
	serviceAccountTokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

var log = core.Log.WithName("injector")

func New(
	cfg runtime_k8s.Injector,
	controlPlaneUrl string,
	client kube_client.Client,
	converter k8s_common.Converter,
) (*KumaInjector, error) {
	var caCert string
	if cfg.CaCertFile != "" {
		bytes, err := ioutil.ReadFile(cfg.CaCertFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read provided CA cert file %s", cfg.CaCertFile)
		}
		caCert = string(bytes)
	}
	return &KumaInjector{
		cfg:             cfg,
		controlPlaneUrl: controlPlaneUrl,
		client:          client,
		converter:       converter,
		caCert:          caCert,
	}, nil
}

type KumaInjector struct {
	cfg             runtime_k8s.Injector
	controlPlaneUrl string
	client          kube_client.Client
	converter       k8s_common.Converter
	caCert          string
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
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info("pod has kuma.io/sidecar-injection: disabled annotation")
		}
		return enabled, nil
	}
	enabled, exist, err = metadata.Annotations(ns.Annotations).GetEnabled(metadata.KumaSidecarInjectionAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		if !enabled {
			log.V(1).Info("namespace has kuma.io/sidecar-injection: disabled annotation")
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

func meshName(pod *kube_core.Pod, ns *kube_core.Namespace) string {
	if mesh, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMeshAnnotation); exist {
		return mesh
	}
	if mesh, exist := metadata.Annotations(ns.Annotations).GetString(metadata.KumaMeshAnnotation); exist {
		return mesh
	}
	return core_model.DefaultMesh
}

func (i *KumaInjector) meshFor(pod *kube_core.Pod, ns *kube_core.Namespace) (*mesh_core.MeshResource, error) {
	meshName := meshName(pod, ns)
	mesh := &mesh_k8s.Mesh{}
	if err := i.client.Get(context.Background(), kube_types.NamespacedName{Name: meshName}, mesh); err != nil {
		return nil, err
	}
	meshResource := mesh_core.NewMeshResource()
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

func (i *KumaInjector) NewSidecarContainer(pod *kube_core.Pod, ns *kube_core.Namespace) (kube_core.Container, error) {
	mesh := meshName(pod, ns)
	env, err := i.sidecarEnvVars(mesh, pod.GetAnnotations())
	if err != nil {
		return kube_core.Container{}, err
	}
	return kube_core.Container{
		Name:            util.KumaSidecarContainerName,
		Image:           i.cfg.SidecarContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Args: []string{
			"run",
			"--log-level=info",
		},
		Env: env,
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  &i.cfg.SidecarContainer.UID,
			RunAsGroup: &i.cfg.SidecarContainer.GID,
		},
		LivenessProbe: &kube_core.Probe{
			Handler: kube_core.Handler{
				HTTPGet: &kube_core.HTTPGetAction{
					Path: "/ready",
					Port: kube_intstr.IntOrString{
						IntVal: int32(i.cfg.SidecarContainer.AdminPort),
					},
				},
			},
			InitialDelaySeconds: i.cfg.SidecarContainer.LivenessProbe.InitialDelaySeconds,
			TimeoutSeconds:      i.cfg.SidecarContainer.LivenessProbe.TimeoutSeconds,
			PeriodSeconds:       i.cfg.SidecarContainer.LivenessProbe.PeriodSeconds,
			SuccessThreshold:    1,
			FailureThreshold:    i.cfg.SidecarContainer.LivenessProbe.FailureThreshold,
		},
		ReadinessProbe: &kube_core.Probe{
			Handler: kube_core.Handler{
				HTTPGet: &kube_core.HTTPGetAction{
					Path: "/ready",
					Port: kube_intstr.IntOrString{
						IntVal: int32(i.cfg.SidecarContainer.AdminPort),
					},
				},
			},
			InitialDelaySeconds: i.cfg.SidecarContainer.ReadinessProbe.InitialDelaySeconds,
			TimeoutSeconds:      i.cfg.SidecarContainer.ReadinessProbe.TimeoutSeconds,
			PeriodSeconds:       i.cfg.SidecarContainer.ReadinessProbe.PeriodSeconds,
			SuccessThreshold:    i.cfg.SidecarContainer.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    i.cfg.SidecarContainer.ReadinessProbe.FailureThreshold,
		},
		Resources: kube_core.ResourceRequirements{
			Requests: kube_core.ResourceList{
				kube_core.ResourceCPU:    kube_api.MustParse(i.cfg.SidecarContainer.Resources.Requests.CPU),
				kube_core.ResourceMemory: kube_api.MustParse(i.cfg.SidecarContainer.Resources.Requests.Memory),
			},
			Limits: kube_core.ResourceList{
				kube_core.ResourceCPU:    kube_api.MustParse(i.cfg.SidecarContainer.Resources.Limits.CPU),
				kube_core.ResourceMemory: kube_api.MustParse(i.cfg.SidecarContainer.Resources.Limits.Memory),
			},
		},
		// On versions of Kubernetes prior to v1.15.0
		// ServiceAccount admission plugin is called only once, prior to any mutating web hook.
		// That's why it is a responsibility of every mutating web hook to copy
		// ServiceAccount volume mount into containers it creates.
		VolumeMounts: i.NewVolumeMounts(pod),
	}, nil
}

func (i *KumaInjector) sidecarEnvVars(mesh string, podAnnotations map[string]string) ([]kube_core.EnvVar, error) {
	envVars := map[string]kube_core.EnvVar{
		"KUMA_CONTROL_PLANE_URL": {
			Name:  "KUMA_CONTROL_PLANE_URL",
			Value: i.controlPlaneUrl,
		},
		"KUMA_DATAPLANE_MESH": {
			Name:  "KUMA_DATAPLANE_MESH",
			Value: mesh,
		},
		"KUMA_DATAPLANE_NAME": {
			Name: "KUMA_DATAPLANE_NAME",
			// notice that Pod name might not be available at this time (in case of Deployment, ReplicaSet, etc)
			// that is why we have to use a runtime reference to POD_NAME instead
			Value: "$(POD_NAME).$(POD_NAMESPACE)", // variable references get expanded by Kubernetes
		},
		"KUMA_DATAPLANE_ADMIN_PORT": {
			Name:  "KUMA_DATAPLANE_ADMIN_PORT",
			Value: fmt.Sprintf("%d", i.cfg.SidecarContainer.AdminPort),
		},
		"KUMA_DATAPLANE_DRAIN_TIME": {
			Name:  "KUMA_DATAPLANE_DRAIN_TIME",
			Value: i.cfg.SidecarContainer.DrainTime.String(),
		},
		"KUMA_DATAPLANE_RUNTIME_TOKEN_PATH": {
			Name:  "KUMA_DATAPLANE_RUNTIME_TOKEN_PATH",
			Value: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		},
		"KUMA_CONTROL_PLANE_CA_CERT": {
			Name:  "KUMA_CONTROL_PLANE_CA_CERT",
			Value: i.caCert,
		},
	}
	if i.cfg.BuiltinDNS.Enabled {
		envVars["KUMA_DNS_ENABLED"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENABLED",
			Value: "true",
		}

		envVars["KUMA_DNS_CORE_DNS_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_CORE_DNS_PORT",
			Value: strconv.FormatInt(int64(i.cfg.BuiltinDNS.Port), 10),
		}

		envVars["KUMA_DNS_CORE_DNS_EMPTY_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_CORE_DNS_EMPTY_PORT",
			Value: strconv.FormatInt(int64(i.cfg.BuiltinDNS.Port+1), 10),
		}

		envVars["KUMA_DNS_ENVOY_DNS_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENVOY_DNS_PORT",
			Value: strconv.FormatInt(int64(i.cfg.BuiltinDNS.Port+2), 10),
		}

		envVars["KUMA_DNS_CORE_DNS_BINARY_PATH"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_CORE_DNS_BINARY_PATH",
			Value: "coredns",
		}
	}

	// override defaults with cfg env vars
	for envName, envVal := range i.cfg.SidecarContainer.EnvVars {
		envVars[envName] = kube_core.EnvVar{
			Name:  envName,
			Value: envVal,
		}
	}

	// override defaults and cfg env vars with annotations
	annotationEnvVars, err := metadata.Annotations(podAnnotations).GetMap(metadata.KumaSidecarEnvVarsAnnotation)
	if err != nil {
		return nil, err
	}
	for envName, envVal := range annotationEnvVars {
		envVars[envName] = kube_core.EnvVar{
			Name:  envName,
			Value: envVal,
		}
	}

	var result []kube_core.EnvVar
	for _, v := range envVars {
		result = append(result, v)
	}
	sort.Stable(EnvVarsByName(result))

	// those values needs to be added before other vars, otherwise expressions like "$(POD_NAME).$(POD_NAMESPACE)" won't be evaluated
	result = append([]kube_core.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "INSTANCE_IP",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}, result...)

	return result, nil
}

func (i *KumaInjector) NewVolumeMounts(pod *kube_core.Pod) []kube_core.VolumeMount {
	if tokenVolumeMount := i.FindServiceAccountToken(pod); tokenVolumeMount != nil {
		return []kube_core.VolumeMount{*tokenVolumeMount}
	}
	return nil
}

func (i *KumaInjector) FindServiceAccountToken(pod *kube_core.Pod) *kube_core.VolumeMount {
	for i := range pod.Spec.Containers {
		for j := range pod.Spec.Containers[i].VolumeMounts {
			if pod.Spec.Containers[i].VolumeMounts[j].MountPath == serviceAccountTokenMountPath {
				return &pod.Spec.Containers[i].VolumeMounts[j]
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
		Name:            util.KumaInitContainerName,
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

func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod, mesh *mesh_core.MeshResource) (map[string]string, error) {
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
	return annotations, nil
}

func portsToAnnotationValue(ports []uint32) string {
	stringPorts := make([]string, len(ports))
	for i, port := range ports {
		stringPorts[i] = fmt.Sprintf("%d", port)
	}
	return strings.Join(stringPorts, ",")
}

type EnvVarsByName []kube_core.EnvVar

func (a EnvVarsByName) Len() int      { return len(a) }
func (a EnvVarsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EnvVarsByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}
