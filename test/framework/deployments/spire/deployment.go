package spire

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

const DeploymentName = "spire"

type Deployment interface {
	framework.Deployment
}

type deployOptions struct {
	namespace   string
	name        string
	trustDomain string
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
				namespace:   opts.namespace,
				name:        opts.name,
				trustDomain: opts.trustDomain,
				kubectlVersion: opts.kubectlVersion
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
