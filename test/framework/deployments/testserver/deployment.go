package testserver

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Name            string
	Namespace       string
	Mesh            string
	WithStatefulSet bool
	Args            []string
	WithHTTPProbes  bool
	Replicas        int32
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh:      "default",
		Args:      []string{},
		Name:      "test-server",
		Namespace: framework.TestNamespace,
		Replicas:  1,
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

func WithArgs(args ...string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Args = args
	}
}

func WithHTTPProbes() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WithHTTPProbes = true
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
