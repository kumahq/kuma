package injector

import (
	"fmt"

	"github.com/Kong/kuma/app/kuma-injector/pkg/injector/metadata"
	config "github.com/Kong/kuma/pkg/config/app/kuma-injector"

	kube_core "k8s.io/api/core/v1"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
)

const (
	KumaSidecarContainerName = "kuma-sidecar"
	KumaInitContainerName    = "kuma-init"
)

const (
	// serviceAccountTokenMountPath is a well-known location where Kubernetes mounts a ServiceAccount token.
	serviceAccountTokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

func New(cfg config.Injector) *KumaInjector {
	return &KumaInjector{
		cfg: cfg,
	}
}

type KumaInjector struct {
	cfg config.Injector
}

func (i *KumaInjector) InjectKuma(pod *kube_core.Pod) error {
	// sidecar container
	if pod.Spec.Containers == nil {
		pod.Spec.Containers = []kube_core.Container{}
	}
	pod.Spec.Containers = append(pod.Spec.Containers, i.NewSidecarContainer(pod))

	// init container
	if pod.Spec.InitContainers == nil {
		pod.Spec.InitContainers = []kube_core.Container{}
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, i.NewInitContainer())

	// annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	for key, value := range i.NewAnnotations(pod) {
		pod.Annotations[key] = value
	}
	return nil
}

func (i *KumaInjector) NewSidecarContainer(pod *kube_core.Pod) kube_core.Container {
	mesh := metadata.GetMesh(pod) // either user-defined value or default
	return kube_core.Container{
		Name:            KumaSidecarContainerName,
		Image:           i.cfg.SidecarContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Args: []string{
			"run",
			"--log-level=info",
		},
		Env: []kube_core.EnvVar{
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
			{
				Name:  "KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_URL",
				Value: i.cfg.ControlPlane.BootstrapServer.URL,
			},
			{
				Name:  "KUMA_DATAPLANE_MESH",
				Value: mesh,
			},
			{
				Name: "KUMA_DATAPLANE_NAME",
				// notice that Pod name might not be available at this time (in case of Deployment, ReplicaSet, etc)
				// that is why we have to use a runtime reference to POD_NAME instead
				Value: "$(POD_NAME).$(POD_NAMESPACE)", // variable references get expanded by Kubernetes
			},
			{
				Name:  "KUMA_DATAPLANE_ADMIN_PORT",
				Value: fmt.Sprintf("%d", i.cfg.SidecarContainer.AdminPort),
			},
			{
				Name:  "KUMA_DATAPLANE_DRAIN_TIME",
				Value: i.cfg.SidecarContainer.DrainTime.String(),
			},
			{
				Name:  "KUMA_DATAPLANE_RUNTIME_TOKEN_PATH",
				Value: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			},
		},
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  &i.cfg.SidecarContainer.UID,
			RunAsGroup: &i.cfg.SidecarContainer.GID,
		},
		LivenessProbe: &kube_core.Probe{
			Handler: kube_core.Handler{
				Exec: &kube_core.ExecAction{
					Command: []string{
						"wget",
						"-qO-",
						fmt.Sprintf("http://localhost:%d", i.cfg.SidecarContainer.AdminPort),
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
				Exec: &kube_core.ExecAction{
					Command: []string{
						"wget",
						"-qO-",
						fmt.Sprintf("http://localhost:%d", i.cfg.SidecarContainer.AdminPort),
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
	}
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

func (i *KumaInjector) NewInitContainer() kube_core.Container {
	return kube_core.Container{
		Name:            KumaInitContainerName,
		Image:           i.cfg.InitContainer.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Args: []string{
			"-p",
			fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPort),
			"-u",
			fmt.Sprintf("%d", i.cfg.SidecarContainer.UID),
			"-g",
			fmt.Sprintf("%d", i.cfg.SidecarContainer.GID),
			"-m",
			"REDIRECT",
			"-i",
			"*",
			"-b",
			"*",
		},
		SecurityContext: &kube_core.SecurityContext{
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
	}
}

func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod) map[string]string {
	return map[string]string{
		metadata.KumaMeshAnnotation:                    metadata.GetMesh(pod), // either user-defined value or default
		metadata.KumaSidecarInjectedAnnotation:         metadata.KumaSidecarInjected,
		metadata.KumaTransparentProxyingAnnotation:     metadata.KumaTransparentProxyingEnabled,
		metadata.KumaTransparentProxyingPortAnnotation: fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPort),
	}
}
