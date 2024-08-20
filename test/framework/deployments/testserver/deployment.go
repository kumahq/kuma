package testserver

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Name                string
	Namespace           string
	Mesh                string
	ReachableServices   []string
	ReachableBackends   string
	WithStatefulSet     bool
	ServiceAccount      string
	echoArgs            []string
	healthcheckTCPArgs  []string
	Replicas            int32
	WaitingToBeReady    bool
	EnableProbes        bool
	EnableService       bool
	HeadlessService     bool
	PodAnnotations      map[string]string
	PodLabels           map[string]string
	NodeSelector        map[string]string
	protocol            string
	tlsKey              string
	tlsCrt              string
	initContainersToAdd []corev1.Container
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh:             "default",
		Name:             "test-server",
		Namespace:        framework.TestNamespace,
		Replicas:         1,
		WaitingToBeReady: true,
		PodAnnotations:   map[string]string{},
		EnableProbes:     true,
		EnableService:    true,
		protocol:         "http",
	}
}

type DeploymentOptsFn = func(*DeploymentOpts)

func WithMesh(mesh string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Mesh = mesh
	}
}

func WithName(name string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Name = name
	}
}

func WithTLS(key, crt string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.tlsKey = key
		opts.tlsCrt = crt
	}
}

func WithReachableServices(services ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ReachableServices = services
	}
}

func WithReachableBackends(config string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ReachableBackends = config
	}
}

func WithNamespace(namespace string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Namespace = namespace
	}
}

func WithReplicas(n int32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Replicas = n
	}
}

func WithStatefulSet() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WithStatefulSet = true
	}
}

func WithServiceAccount(serviceAccountName string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ServiceAccount = serviceAccountName
	}
}

func WithoutWaitingToBeReady() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WaitingToBeReady = false
	}
}

func WithEchoArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.echoArgs = args
	}
}

func WithHealthCheckTCPArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.healthcheckTCPArgs = args
	}
}

func WithPodAnnotations(annotations map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.PodAnnotations = annotations
	}
}

func WithPodLabels(labels map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.PodLabels = labels
	}
}

func WithoutProbes() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.EnableProbes = false
	}
}

func WithoutService() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.EnableService = false
	}
}

func WithHeadlessService() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.HeadlessService = true
	}
}

func WithNodeSelector(selector map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.NodeSelector = selector
	}
}

func WithServicePortAppProtocol(protocol string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.protocol = protocol
	}
}

func AddInitContainer(initContainer corev1.Container) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.initContainersToAdd = append(opts.initContainersToAdd, initContainer)
	}
}

type TestServer interface{}

type Deployment interface {
	framework.Deployment
	TestServer
}

func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	opts := DefaultDeploymentOpts()
	for _, f := range fn {
		f(&opts)
	}

	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{
				opts: opts,
			}
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}
