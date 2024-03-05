package democlient

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Name             string
	Namespace        string
	Mesh             string
	WaitingToBeReady bool
	Service          bool
	PodAnnotations   map[string]string
	NodeSelector     map[string]string
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh:             "default",
		Name:             "demo-client",
		Namespace:        framework.TestNamespace,
		WaitingToBeReady: true,
		PodAnnotations:   map[string]string{},
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

func WithoutWaitingToBeReady() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.WaitingToBeReady = false
	}
}

func WithPodAnnotations(annotations map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.PodAnnotations = annotations
	}
}

func WithNodeSelector(selector map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.NodeSelector = selector
	}
}

func WithService(service bool) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Service = service
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
