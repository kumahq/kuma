package ebpf_cleaner

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework"
)

type K8sDeployment struct {
	ServiceName      string
	Namespace        string
	Replicas         int32
	WaitingToBeReady bool
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

func WithReplicas(n int32) DeploymentOptsFn {
	return func(opts *K8sDeployment) {
		opts.Replicas = n
	}
}

func WithoutWaitingToBeReady() DeploymentOptsFn {
	return func(opts *K8sDeployment) {
		opts.WaitingToBeReady = false
	}
}

func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	deployment := K8sDeployment{
		ServiceName:      "ebpf-cleaner",
		Namespace:        "default",
		Replicas:         1,
		WaitingToBeReady: true,
	}
	for _, f := range fn {
		f(&deployment)
	}

	return func(cluster framework.Cluster) error {
		return cluster.Deploy(&deployment)
	}
}

func (k *K8sDeployment) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.Namespace,
		},
	}
}

func (k *K8sDeployment) clusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"watch", "delete", "deletecollection"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"watch", "create", "delete", "deletecollection"},
			},
		},
	}
}

func (k *K8sDeployment) clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     k.Name(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      k.Name(),
				Namespace: k.Namespace,
			},
		},
	}
}

func (k *K8sDeployment) job() *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: k.podSpec(),
		},
	}
}

func (k *K8sDeployment) podSpec() corev1.PodTemplateSpec {
	var args []string
	spec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": k.Name()},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: k.Name(),
			Containers: []corev1.Container{
				{
					Name:            k.Name(),
					ImagePullPolicy: "IfNotPresent",
					Image:           fmt.Sprintf("%s/%s:%s", framework.Config.KumaImageRegistry, framework.Config.KumactlImageRepo, framework.Config.KumaImageTag),
					Command:         []string{"kumactl", "uninstall", "ebpf", "--namespace", k.Namespace},
					Args:            args,
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	return spec
}

func (k *K8sDeployment) Name() string {
	return k.ServiceName
}

func (k *K8sDeployment) Deploy(cluster framework.Cluster) error {
	var funcs []framework.InstallFunc
	funcs = append(funcs, framework.YamlK8sObject(k.serviceAccount()))
	funcs = append(funcs, framework.YamlK8sObject(k.clusterRole()))
	funcs = append(funcs, framework.YamlK8sObject(k.clusterRoleBinding()))
	funcs = append(funcs, framework.YamlK8sObject(k.job()))
	if k.WaitingToBeReady {
		funcs = append(funcs,
			framework.WaitNumPods(k.Namespace, 1, k.Name()),
			framework.WaitPodsAvailable(k.Namespace, k.Name()),
		)
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
