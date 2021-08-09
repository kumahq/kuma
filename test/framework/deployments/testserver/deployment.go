package testserver

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type DeploymentOpts struct {
	Mesh string
}

func DefaultDeploymentOpts() DeploymentOpts {
	return DeploymentOpts{
		Mesh: "default",
	}
}

type DeploymentOptsFn = func(*DeploymentOpts)

func WithMesh(mesh string) DeploymentOptsFn {
	return func(opts *DeploymentOpts) {
		opts.Mesh = mesh
	}
}

type TestServer interface {
}

type Deployment interface {
	framework.Deployment
	TestServer
}

const DeploymentName = "test-server"

func From(cluster framework.Cluster) TestServer {
	return cluster.Deployment(DeploymentName).(TestServer)
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
