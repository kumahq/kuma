package ebpf_checker

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type K8sDeployment struct {
	ServiceName    string
	Namespace      string
	Replicas       int32
	DisableSidecar bool // uses default namespace settings
}
type Deployment interface {
	framework.Deployment
}

type DeploymentOptsFn = func(*K8sDeployment)

func WithName(name string) DeploymentOptsFn {
	return func(opts *K8sDeployment) {
		opts.ServiceName = name
	}
}

func WithNamespace(namespace string) DeploymentOptsFn {
	return func(opts *K8sDeployment) {
		opts.Namespace = namespace
	}
}

func WithoutSidecar() DeploymentOptsFn {
	return func(opts *K8sDeployment) {
		opts.DisableSidecar = true
	}
}

func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	deployment := K8sDeployment{
		ServiceName:    "ebpf-checker",
		Namespace:      "default",
		DisableSidecar: false,
		Replicas:       1,
	}
	for _, f := range fn {
		f(&deployment)
	}

	return func(cluster framework.Cluster) error {
		return cluster.Deploy(&deployment)
	}
}

func (k *K8sDeployment) deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &k.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": k.Name()},
			},
			Template: k.podSpec(),
		},
	}
}

func (k *K8sDeployment) podSpec() corev1.PodTemplateSpec {
	var args []string
	labels := map[string]string{"app": k.Name()}
	if k.DisableSidecar {
		labels["kuma.io/sidecar-injection"] = "false"
	}
	spec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            k.Name(),
					ImagePullPolicy: "IfNotPresent",
					Image:           fmt.Sprintf("%s/%s:%s", framework.Config.KumaImageRegistry, framework.Config.KumaUniversalImageRepo, framework.Config.KumaImageTag),
					Ports: []corev1.ContainerPort{
						{ContainerPort: 80},
					},
					Command: []string{"/usr/bin/sleep", "3600"},
					Args:    args,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "bpf-fs-path",
							MountPath: "/sys/fs/bpf",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "bpf-fs-path",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/sys/fs/bpf",
						},
					},
				},
			},
		},
	}
	return spec
}

func (k *K8sDeployment) Name() string {
	return k.ServiceName
}

func (k *K8sDeployment) Deploy(cluster framework.Cluster) error {
	funcs := []framework.InstallFunc{
		framework.YamlK8sObject(k.deployment()),
		framework.WaitNumPods(k.Namespace, 1, k.Name()),
		framework.WaitPodsAvailable(k.Namespace, k.Name()),
	}
	return framework.Combine(funcs...)(cluster)
}

func (k *K8sDeployment) Delete(cluster framework.Cluster) error {
	// todo(jakubdyszkiewicz) right now we delete TestNamespace before we Dismiss the cluster
	// This means that namespace is no longer available so the code below would throw an error
	// If we ever switch DemoClient to be deployment and remove manual deletion of TestNamespace
	// then we can rely on code below to delete tht deployment.

	// k8s.KubectlDeleteFromString(
	// 	cluster.GetTesting(),
	// 	cluster.GetKubectlOptions(framework.TestNamespace),
	// 	service,
	// )
	// k8s.KubectlDeleteFromString(
	// 	cluster.GetTesting(),
	// 	cluster.GetKubectlOptions(framework.TestNamespace),
	// 	fmt.Sprintf(deployment, k.opts.Mesh, framework.GetUniversalImage()),
	// )
	return nil
}

var _ Deployment = &K8sDeployment{}
