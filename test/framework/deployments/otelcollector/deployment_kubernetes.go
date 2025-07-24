package otelcollector

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/test/framework"
)

type K8SDeployment struct {
	name               string
	namespace          string
	image              string
	waitingToBeReady   bool
	serviceAccountName string
	isIPv6             bool
}

var _ Deployment = &K8SDeployment{}

func (k *K8SDeployment) Name() string {
	return k.name
}

func (k *K8SDeployment) Deploy(cluster framework.Cluster) error {
	var funcs []framework.InstallFunc
	funcs = append(funcs,
		framework.YamlK8sObject(k.serviceAccount()),
		framework.YamlK8sObject(k.configMap()),
		framework.YamlK8sObject(k.deployment()),
		framework.YamlK8sObject(k.service()),
	)
	if k.waitingToBeReady {
		funcs = append(funcs,
			framework.WaitService(k.namespace, k.Name()),
			framework.WaitNumPods(k.namespace, 1, k.Name()),
			framework.WaitPodsAvailable(k.namespace, k.Name()),
		)
	}
	return framework.Combine(funcs...)(cluster)
}

func (k *K8SDeployment) Delete(cluster framework.Cluster) error {
	return nil
}

func (k *K8SDeployment) CollectorEndpoint() string {
	return fmt.Sprintf("%s.%s:%d", k.Name(), k.namespace, GRPCPort)
}

func (k *K8SDeployment) ExporterEndpoint() string {
	return fmt.Sprintf("%s.%s:%d/metrics", k.Name(), k.namespace, PrometheusExporterPort)
}

func newK8sDeployment() *K8SDeployment {
	return &K8SDeployment{}
}

func (k *K8SDeployment) WithName(name string) *K8SDeployment {
	k.name = name
	return k
}

func (k *K8SDeployment) WithNamespace(namespace string) *K8SDeployment {
	k.namespace = namespace
	return k
}

func (k *K8SDeployment) WithImage(image string) *K8SDeployment {
	k.image = image
	return k
}

func (k *K8SDeployment) WithWaitingToBeReady(waitingToBeReady bool) *K8SDeployment {
	k.waitingToBeReady = waitingToBeReady
	return k
}

func (k *K8SDeployment) WithServiceAccount(serviceAccountName string) *K8SDeployment {
	k.serviceAccountName = serviceAccountName
	return k
}

func (k *K8SDeployment) WithIPv6(isIPv6 bool) *K8SDeployment {
	k.isIPv6 = isIPv6
	return k
}

func (k *K8SDeployment) service() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.Name(),
			Namespace: k.namespace,
			Labels: map[string]string{
				"app": k.Name(),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "otlp-grpc",
					Port: GRPCPort,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: GRPCPort,
					},
					Protocol: "TCP",
				},
				{
					Name: "otlp-http",
					Port: HTTPPort,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: HTTPPort,
					},
					Protocol: "TCP",
				},
				{
					Name: "metrics",
					Port: MetricsPort,
				},
				{
					Name: "prometheus",
					Port: PrometheusExporterPort,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: PrometheusExporterPort,
					},
				},
			},
			Selector: map[string]string{
				"app": k.Name(),
			},
		},
	}
}

func (k *K8SDeployment) deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: meta(k.namespace, k.Name(), map[string]string{"app": k.Name()}),
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": k.Name(),
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
				},
			},
			MinReadySeconds:         int32(5),
			ProgressDeadlineSeconds: pointer.To(int32(120)),
			Template:                k.podSpec(),
		},
	}
}

func (k *K8SDeployment) podSpec() corev1.PodTemplateSpec {
	spec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app": k.Name(),
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: k.serviceAccountName,
			Containers: []corev1.Container{
				{
					Name:            k.Name(),
					ImagePullPolicy: "IfNotPresent",
					Image:           k.image,
					Args:            []string{"--config=/conf/otel-collector-config.yaml"},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("100m"),
							"memory": resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("500Mi"),
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "grpc-otel",
							ContainerPort: GRPCPort,
						},
						{
							Name:          "http-otel",
							ContainerPort: HTTPPort,
						},
						{
							Name:          "metrics",
							ContainerPort: MetricsPort,
						},
						{
							Name:          "prometheus",
							ContainerPort: PrometheusExporterPort,
						},
					},
					Env: []corev1.EnvVar{
						{Name: "MY_POD_IP", ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								APIVersion: "v1",
								FieldPath:  "status.podIP",
							},
						}},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "otel-collector-config-vol",
							ReadOnly:  true,
							MountPath: "/conf",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "otel-collector-config-vol",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "otel-collector-conf"},
							DefaultMode:          pointer.To(int32(0o555)),
							Items: []corev1.KeyToPath{{
								Key:  "config",
								Path: "otel-collector-config.yaml",
							}},
						},
					},
				},
			},
		},
	}
	return spec
}

func (k *K8SDeployment) configMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: meta(k.namespace, "otel-collector-conf", map[string]string{"app": k.Name()}),
		Data: map[string]string{
			"config": config(k.endpointBasedOnIP()),
		},
	}
}

func (k *K8SDeployment) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: meta(k.namespace, k.serviceAccountName, map[string]string{"app": k.Name()}),
	}
}

func meta(namespace, name string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
}

func (k *K8SDeployment) endpointBasedOnIP() string {
	if k.isIPv6 {
		return "[${env:MY_POD_IP}]"
	}
	return "${env:MY_POD_IP}"
}

func config(endpoint string) string {
	return fmt.Sprintf(`receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "%s:%d"
      http:
        endpoint: "%s:%d"
processors:
  batch:
    send_batch_size: 4096
    send_batch_max_size: 8192
  memory_limiter:
    limit_mib: 500
    spike_limit_mib: 400
    check_interval: 5s
extensions:
  zpages: {}
exporters:
  debug:
    verbosity: basic
  prometheus:
    endpoint: "%s:%d"
service:
  extensions: [zpages]
  pipelines:
    traces/1:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [debug]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug, prometheus]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]`, endpoint, GRPCPort, endpoint, HTTPPort, endpoint, PrometheusExporterPort)
}
