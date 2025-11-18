package certmanager

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/test/framework"
)

const DeploymentName = "cert-manager"

type Deployment interface {
	framework.Deployment
}

type deployOptions struct {
	namespace string
	version   string
}

type deployOptionsFunc func(*deployOptions)

func newDeployOpt(fs ...deployOptionsFunc) *deployOptions {
	rv := &deployOptions{
		namespace: "cert-manager",
		version:   "v1.13.0",
	}
	for _, f := range fs {
		f(rv)
	}
	return rv
}

func From(cluster framework.Cluster) Deployment {
	return cluster.Deployment(DeploymentName)
}

func Install(fs ...deployOptionsFunc) framework.InstallFunc {
	opts := newDeployOpt(fs...)
	return func(cluster framework.Cluster) error {
		var deployment *k8sDeployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8sDeployment{
				namespace: opts.namespace,
				version:   opts.version,
			}
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

func WithVersion(version string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.version = version
	}
}
