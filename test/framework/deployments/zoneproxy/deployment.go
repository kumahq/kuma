package zoneproxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/test/framework"
)

const (
	DefaultIngressPort uint32 = 11001
	DefaultEgressPort  uint32 = 11002
)

func ProxyName(mesh string) string {
	return fmt.Sprintf("%s-zone-proxy", mesh)
}

func IngressName(mesh string) string {
	return ingressName(ProxyName(mesh))
}

func EgressName(mesh string) string {
	return egressName(ProxyName(mesh))
}

func ingressName(name string) string {
	return fmt.Sprintf("%s-ingress", name)
}

func egressName(name string) string {
	return fmt.Sprintf("%s-egress", name)
}

type DeploymentOpts struct {
	Name        string
	Namespace   string
	Mesh        string
	Workload    string
	Ingress     bool
	Egress      bool
	IngressPort uint32
	EgressPort  uint32
	// DpEnvs are extra kuma-dp environment variables applied to the proxy.
	// Universal-only; ignored on k8s deployments.
	DpEnvs map[string]string
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Namespace:   framework.TestNamespace,
		Mesh:        "default",
		IngressPort: DefaultIngressPort,
		EgressPort:  DefaultEgressPort,
	}
}

type DeploymentOptsFn = func(*DeploymentOpts)

// WithName overrides the deployment name, which otherwise defaults to
// ProxyName(mesh).
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

// WithIngress deploys an ingress on DefaultIngressPort.
func WithIngress() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Ingress = true
	}
}

// WithIngressPort deploys an ingress on the given port.
func WithIngressPort(port uint32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Ingress = true
		opts.IngressPort = port
	}
}

// WithEgress deploys an egress on DefaultEgressPort.
func WithEgress() DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Egress = true
	}
}

// WithEgressPort deploys an egress on the given port.
func WithEgressPort(port uint32) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Egress = true
		opts.EgressPort = port
	}
}

// WithDpEnvs sets kuma-dp environment overrides for the zone proxy DPP.
// Universal-only.
func WithDpEnvs(envs map[string]string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.DpEnvs = envs
	}
}

// Install deploys the zone proxies of a mesh. Both an ingress and an egress are
// deployed unless the caller asks for one of them, so a mesh gets its proxies
// with Install(WithMesh(mesh)).
func Install(fn ...DeploymentOptsFn) framework.InstallFunc {
	opts := DefaultDeploymentOpts()
	for _, f := range fn {
		f(&opts)
	}
	if !opts.Ingress && !opts.Egress {
		opts.Ingress = true
		opts.Egress = true
	}
	if opts.Name == "" {
		opts.Name = ProxyName(opts.Mesh)
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
