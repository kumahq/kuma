package injector

import (
	"context"
	"fmt"

	runtime_k8s "github.com/Kong/kuma/pkg/config/plugins/runtime/k8s"

	"github.com/pkg/errors"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector/metadata"

	kube_core "k8s.io/api/core/v1"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KumaSidecarContainerName = "kuma-sidecar"
	KumaInitContainerName    = "kuma-init"
)

const (
	// serviceAccountTokenMountPath is a well-known location where Kubernetes mounts a ServiceAccount token.
	serviceAccountTokenMountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

func New(cfg runtime_k8s.Injector, controlPlaneUrl string, client kube_client.Client) *KumaInjector {
	return &KumaInjector{
		cfg:             cfg,
		controlPlaneUrl: controlPlaneUrl,
		client:          client,
		converter:       k8s_resources.DefaultConverter(),
	}
}

type KumaInjector struct {
	cfg             runtime_k8s.Injector
	controlPlaneUrl string
	client          kube_client.Client
	converter       k8s_resources.Converter
}

func (i *KumaInjector) InjectKuma(pod *kube_core.Pod) error {
	// first, give users a chance to opt out of side-car injection into a given Pod
	if pod.Annotations[metadata.KumaSidecarInjectionAnnotation] == metadata.KumaSidecarInjectionDisabled {
		return nil
	}
	// sidecar container
	if pod.Spec.Containers == nil {
		pod.Spec.Containers = []kube_core.Container{}
	}
	pod.Spec.Containers = append(pod.Spec.Containers, i.NewSidecarContainer(pod))

	mesh, err := i.meshFor(pod)
	if err != nil {
		return errors.Wrap(err, "could not retrieve mesh for pod")
	}

	// annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	for key, value := range i.NewAnnotations(pod, mesh) {
		pod.Annotations[key] = value
	}

	// init container
	if !i.cfg.CNIEnabled {
		if pod.Spec.InitContainers == nil {
			pod.Spec.InitContainers = []kube_core.Container{}
		}
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, i.NewInitContainer(pod))
	}

	return nil
}

func (i *KumaInjector) meshFor(pod *kube_core.Pod) (*mesh_core.MeshResource, error) {
	meshName := metadata.GetMesh(pod) // either user-defined value or default
	mesh := &mesh_k8s.Mesh{}
	if err := i.client.Get(context.Background(), kube_types.NamespacedName{Name: meshName}, mesh); err != nil {
		return nil, err
	}
	meshResource := &mesh_core.MeshResource{}
	if err := i.converter.ToCoreResource(mesh, meshResource); err != nil {
		return nil, err
	}
	return meshResource, nil
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
				Name:  "KUMA_CONTROL_PLANE_API_SERVER_URL",
				Value: i.controlPlaneUrl,
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
						fmt.Sprintf("http://127.0.0.1:%d", i.cfg.SidecarContainer.AdminPort),
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
						fmt.Sprintf("http://127.0.0.1:%d", i.cfg.SidecarContainer.AdminPort),
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

func (i *KumaInjector) NewInitContainer(pod *kube_core.Pod) kube_core.Container {
	inboundPortsToIntercept := "*"
	if pod.GetAnnotations()[metadata.KumaGatewayAnnotation] == metadata.KumaGatewayEnabled {
		inboundPortsToIntercept = ""
	}
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
			inboundPortsToIntercept,
		},
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
	}
}

func (i *KumaInjector) NewAnnotations(pod *kube_core.Pod, mesh *mesh_core.MeshResource) map[string]string {
	annotations := map[string]string{
		metadata.KumaMeshAnnotation:                    mesh.GetMeta().GetName(), // either user-defined value or default
		metadata.KumaSidecarInjectedAnnotation:         metadata.KumaSidecarInjected,
		metadata.KumaTransparentProxyingAnnotation:     metadata.KumaTransparentProxyingEnabled,
		metadata.KumaTransparentProxyingPortAnnotation: fmt.Sprintf("%d", i.cfg.SidecarContainer.RedirectPort),
	}
	if i.cfg.CNIEnabled {
		annotations[metadata.CNCFNetworkAnnotation] = metadata.KumaCNI
	}
	return annotations
}
