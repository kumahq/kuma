package testserver

import (
	"strings"

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

func (k *k8SDeployment) service() *corev1.Service {
	appProtocol := "http"
	if len(k.opts.healthcheckTCPArgs) > 0 {
		appProtocol = "tcp"
	}
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
					Name:        "main",
					Port:        80,
					AppProtocol: &appProtocol,
				},
			},
			Selector: map[string]string{
				"app": k.Name(),
			},
		},
	}
}

func (k *k8SDeployment) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.opts.ServiceAccount,
			Namespace: k.opts.Namespace,
		},
	}
}

func (k *k8SDeployment) deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta(k.opts.Namespace, k.Name()),
		Spec: appsv1.DeploymentSpec{
			Replicas: &k.opts.Replicas,
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

func (k *k8SDeployment) statefulSet() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta(k.opts.Namespace, k.Name()),
		Spec: appsv1.StatefulSetSpec{
			ServiceName: k.Name(),
			Replicas:    &k.opts.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": k.Name()},
			},
			Template: k.podSpec(),
		},
	}
}

func (k *k8SDeployment) podSpec() corev1.PodTemplateSpec {
	var args []string
	var liveness *corev1.Probe
	var readiness *corev1.Probe
	if len(k.opts.healthcheckTCPArgs) > 0 {
		args = k.opts.healthcheckTCPArgs
		liveness = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
		readiness = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
	} else {
		args = append([]string{"echo", "--port", "80", "--probes"}, k.opts.echoArgs...)
		liveness = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: `/probes?type=liveness`,
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
		readiness = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: `/probes?type=readiness`,
					Port: intstr.FromInt(80),
				},
			},
			InitialDelaySeconds: 3,
			PeriodSeconds:       3,
		}
	}
	if !k.opts.EnableProbes {
		liveness = nil
		readiness = nil
	}
	spec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{"app": k.Name()},
			Annotations: k.getAnnotations(),
		},
		Spec: corev1.PodSpec{
			NodeSelector:       k.opts.NodeSelector,
			ServiceAccountName: k.opts.ServiceAccount,
			Containers: []corev1.Container{
				{
					Name:            k.Name(),
					ImagePullPolicy: "IfNotPresent",
					ReadinessProbe:  readiness,
					LivenessProbe:   liveness,
					Image:           framework.Config.GetUniversalImage(),
					Ports: []corev1.ContainerPort{
						{ContainerPort: 80},
					},
					Env: []corev1.EnvVar{
						{
							Name: "POD_IP",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "status.podIP",
								},
							},
						},
					},
					Command: []string{"test-server"},
					Args:    args,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
						},
					},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{ // test-server does not handle graceful shutdown itself
								Command: []string{"/usr/bin/sleep", "30"},
							},
						},
					},
				},
			},
		},
	}
	if len(k.opts.ReachableServices) > 0 {
		spec.ObjectMeta.Annotations["kuma.io/transparent-proxying-reachable-services"] = strings.Join(k.opts.ReachableServices, ",")
	}
	return spec
}

func (k *k8SDeployment) getAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations["kuma.io/mesh"] = k.opts.Mesh
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
	var funcs []framework.InstallFunc
	if k.opts.ServiceAccount != "" {
		funcs = append(funcs, framework.YamlK8sObject(k.serviceAccount()))
	}
	if k.opts.WithStatefulSet {
		funcs = append(funcs, framework.YamlK8sObject(k.statefulSet()))
	} else {
		funcs = append(funcs, framework.YamlK8sObject(k.deployment()))
	}
	funcs = append(funcs, framework.YamlK8sObject(k.service()))
	if k.opts.WaitingToBeReady {
		funcs = append(funcs,
			framework.WaitService(k.opts.Namespace, k.Name()),
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
