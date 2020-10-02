package externalservice

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type ExternalService interface {
	Init(name string, args []string) error
	GetExternalAppAddress() string
}

type Deployment interface {
	framework.Deployment
	ExternalService
}

const DeploymentName = "externalservice-"

var UniversalAppEchoServer = []string{"ncat", "-lk", "-p", "80", "--sh-exec", "'echo \"HTTP/1.1 200 OK\n\n Echo\n\"'"}
var UniversalAppHttpsEchoServer = []string{"ncat", "-lk", "-p", "443", "--ssl", "--sh-exec", "'echo \"HTTP/1.1 200 OK\n\n HTTPS Echo\n\"'"}

func From(cluster framework.Cluster, name string) ExternalService {
	return cluster.Deployment(DeploymentName + name).(ExternalService)
}

func Install(name string, args []string) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{}
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				ports: map[string]string{},
			}
		default:
			return errors.New("invalid cluster")
		}

		deployment.Init(name, args)
		return cluster.Deploy(deployment)
	}
}
