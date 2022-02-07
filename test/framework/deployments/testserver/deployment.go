package testserver

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Name              string
	Namespace         string
	Mesh              string
	ReachableServices []string
	WithStatefulSet   bool
	ServiceAccount    string
	Args              []string
	Replicas          int32
	WaitingToBeReady  bool
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh:             "default",
		Args:             []string{},
		Name:             "test-server",
		Namespace:        framework.TestNamespace,
		Replicas:         1,
		WaitingToBeReady: true,
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

func WithReachableServices(services ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.ReachableServices = services
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

func WithStatefulSet(apply bool) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WithStatefulSet = apply
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

func WithArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Args = args
	}
}

type TestServer interface {
}

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
