package spire

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/test/framework"
)

type SpireServer interface {
	GetAgentJoinToken(framework.Cluster, string) (string, error)
	RegisterWorkload(framework.Cluster, string, string, string) error
	ExecSpireServerCommand(framework.Cluster, ...string) (string, error)
	GetIP() (string, error)
	framework.Deployment
}

const DeploymentName = "spire"

type deployOptions struct {
	namespace      string
	name           string
	trustDomain    string
	kubectlVersion string
}

type deployOptionsFunc func(*deployOptions)

func newDeployOpt(fs ...deployOptionsFunc) *deployOptions {
	rv := &deployOptions{}
	for _, f := range fs {
		f(rv)
	}
	return rv
}

func From(deploymentName string, cluster framework.Cluster) SpireServer {
	return cluster.Deployment(deploymentName).(SpireServer)
}

func Install(fs ...deployOptionsFunc) framework.InstallFunc {
	opts := newDeployOpt(fs...)
	return func(cluster framework.Cluster) error {
		var deployment framework.Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8sDeployment{
				namespace:      opts.namespace,
				name:           opts.name,
				trustDomain:    opts.trustDomain,
				kubectlVersion: opts.kubectlVersion,
			}
		case *framework.UniversalCluster:
			deployment = NewUniversalDeployment(cluster, opts.name, opts)
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}

func WithNamespace(namespace string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.namespace = namespace
	}
}

func WithName(name string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.name = name
	}
}

func WithTrustDomain(td string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.trustDomain = td
	}
}

func WithKubectlVersion(version string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.kubectlVersion = version
	}
}
