package externalservice

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type ExternalService interface {
	GetExternalAppAddress() string
	GetCert() string
}

type Deployment interface {
	framework.Deployment
	ExternalService
}

const (
	DeploymentName = "externalservice-"
	HttpServer     = "http-server"
	HttpsServer    = "https-server"
)

func From(cluster framework.Cluster, name string) ExternalService {
	return cluster.Deployment(DeploymentName + name).(ExternalService)
}

func Install(name string, args []string) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{
				name: name,
				args: args,
			}
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				name:  name,
				args:  args,
				ports: map[string]string{},
			}
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
