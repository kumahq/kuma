package zoneproxy

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/test/framework"
)

type DeploymentOpts struct {
	Name        string
	Namespace   string
	Mesh        string
	Workload    string
	IngressPort uint32
	EgressPort  uint32
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Name:      "zone-proxy",
		Namespace: framework.TestNamespace,
		Mesh:      "default",
	}
}

type DeploymentOptsFn = func(*DeploymentOpts)

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

func WithMesh(mesh string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Mesh = mesh
	}
}

func WithWorkload(workload string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Workload = workload
	}
}

func WithIngressPort(port uint32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.IngressPort = port
	}
}

func WithEgressPort(port uint32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.EgressPort = port
	}
}

// Install deploys zone-proxy-ingress and/or zone-proxy-egress based on which
// ports are set. Set IngressPort to deploy ingress, EgressPort to deploy egress.
func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	opts := DefaultDeploymentOpts()
	for _, f := range fn {
		f(&opts)
	}
	return func(cluster framework.Cluster) error {
		switch cluster.(type) {
		case *framework.K8sCluster:
			return cluster.Deploy(&k8sDeployment{opts: opts})
		case *framework.UniversalCluster:
			return cluster.Deploy(&universalDeployment{opts: opts})
		default:
			return errors.New("zone proxy deployment is not supported on this cluster type")
		}
	}
}
