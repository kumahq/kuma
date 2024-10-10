package democlient

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	opts DeploymentOpts
}

func (k *k8SDeployment) Name() string {
	return k.opts.Name
}

func (k *k8SDeployment) deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta(k.opts.Namespace, k.Name()),
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": k.Name()},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
				},
			},
			Template: k.podSpec(),
		},
	}
}

func (k *k8SDeployment) podSpec() corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{"app": k.Name(), "kuma.io/mesh": k.opts.Mesh},
			Annotations: k.getAnnotations(),
		},
		Spec: corev1.PodSpec{
			NodeSelector: k.opts.NodeSelector,
			Containers: []corev1.Container{
				{
					Name:            k.Name(),
					ImagePullPolicy: "IfNotPresent",
					Image:           framework.Config.GetUniversalImage(),
					Ports: []corev1.ContainerPort{
						{ContainerPort: 3000, Name: "main"},
					},
					Command: []string{"ncat"},
					Args:    []string{"-lk", "-p", "3000"},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
						},
					},
				},
			},
		},
	}
}

func (k *k8SDeployment) service() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.opts.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "main",
					Port:       3000,
					TargetPort: intstr.FromString("main"),
				},
			},
			Selector: map[string]string{
				"app": k.Name(),
			},
		},
	}
}

func (k *k8SDeployment) getAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, value := range k.opts.PodAnnotations {
		annotations[key] = value
	}
	return annotations
}

func meta(namespace string, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    map[string]string{"app": name},
	}
}

func (k *k8SDeployment) Deploy(cluster framework.Cluster) error {
	funcs := []framework.InstallFunc{framework.YamlK8sObject(k.deployment())}

	if k.opts.Service {
		funcs = append(funcs, framework.YamlK8sObject(k.service()))
	}

	if k.opts.WaitingToBeReady {
		funcs = append(funcs,
			framework.WaitNumPods(k.opts.Namespace, 1, k.Name()),
			framework.WaitPodsAvailable(k.opts.Namespace, k.Name()),
		)
	}
	return framework.Combine(funcs...)(cluster)
}

func (k *k8SDeployment) Delete(cluster framework.Cluster) error {
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

var _ Deployment = &k8SDeployment{}
