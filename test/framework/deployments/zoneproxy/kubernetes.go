package zoneproxy

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/v2/test/framework"
)

const zoneProxyTypeLabel = "k8s.kuma.io/zone-proxy-type"

type k8sDeployment struct {
	opts DeploymentOpts
}

func (d *k8sDeployment) Name() string {
	return d.opts.Name
}

func (d *k8sDeployment) ingressName() string {
	return fmt.Sprintf("%s-ingress", d.opts.Name)
}

func (d *k8sDeployment) egressName() string {
	return fmt.Sprintf("%s-egress", d.opts.Name)
}

func (d *k8sDeployment) ingressDeployment() *appsv1.Deployment {
	replicas := int32(1)
	name := d.ingressName()
	port := int32(d.opts.IngressPort)
	podLabels := map[string]string{
		"app":          name,
		"kuma.io/mesh": d.opts.Mesh,
	}
	if d.opts.Workload != "" {
		podLabels["kuma.io/workload"] = d.opts.Workload
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.opts.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           framework.Config.GetUniversalImage(),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/usr/bin/sleep"},
							Args:            []string{"3600"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
									Name:          "zi-main",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *k8sDeployment) ingressService() *corev1.Service {
	name := d.ingressName()
	port := int32(d.opts.IngressPort)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.opts.Namespace,
			Labels: map[string]string{
				zoneProxyTypeLabel: "ingress",
				"kuma.io/mesh":     d.opts.Mesh,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"app": name},
			Ports: []corev1.ServicePort{
				{
					Name:       "zi-main",
					Protocol:   corev1.ProtocolTCP,
					Port:       port,
					TargetPort: intstr.FromInt32(port),
				},
			},
		},
	}
}

func (d *k8sDeployment) egressDeployment() *appsv1.Deployment {
	replicas := int32(1)
	name := d.egressName()
	port := int32(d.opts.EgressPort)
	podLabels := map[string]string{
		"app":          name,
		"kuma.io/mesh": d.opts.Mesh,
	}
	if d.opts.Workload != "" {
		podLabels["kuma.io/workload"] = d.opts.Workload
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.opts.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           framework.Config.GetUniversalImage(),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/usr/bin/sleep"},
							Args:            []string{"3600"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
									Name:          "ze-main",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *k8sDeployment) egressService() *corev1.Service {
	name := d.egressName()
	port := int32(d.opts.EgressPort)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.opts.Namespace,
			Labels: map[string]string{
				zoneProxyTypeLabel: "egress",
				"kuma.io/mesh":     d.opts.Mesh,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports: []corev1.ServicePort{
				{
					Name:       "ze-main",
					Protocol:   corev1.ProtocolTCP,
					Port:       port,
					TargetPort: intstr.FromInt32(port),
				},
			},
		},
	}
}

func (d *k8sDeployment) Deploy(cluster framework.Cluster) error {
	var funcs []framework.InstallFunc
	if d.opts.IngressPort > 0 {
		name := d.ingressName()
		funcs = append(funcs,
			framework.YamlK8sObject(d.ingressDeployment()),
			framework.YamlK8sObject(d.ingressService()),
			framework.WaitNumPods(d.opts.Namespace, 1, name),
			framework.WaitPodsAvailable(d.opts.Namespace, name),
			framework.WaitService(d.opts.Namespace, name),
		)
	}
	if d.opts.EgressPort > 0 {
		name := d.egressName()
		funcs = append(funcs,
			framework.YamlK8sObject(d.egressDeployment()),
			framework.YamlK8sObject(d.egressService()),
			framework.WaitNumPods(d.opts.Namespace, 1, name),
			framework.WaitPodsAvailable(d.opts.Namespace, name),
			framework.WaitService(d.opts.Namespace, name),
		)
	}
	return framework.Combine(funcs...)(cluster)
}

func (d *k8sDeployment) Delete(_ framework.Cluster) error {
	return nil
}

var _ framework.Deployment = &k8sDeployment{}
