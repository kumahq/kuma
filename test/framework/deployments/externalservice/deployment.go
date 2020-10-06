package externalservice

import (
	"github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

type ExternalService interface {
	Init(cluster framework.Cluster, name string, args []string) error
	GetExternalAppAddress() string
	Cleanup(cluster framework.Cluster) error
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
			deployment = &k8SDeployment{}
		case *framework.UniversalCluster:
			deployment = &universalDeployment{
				ports: map[string]string{},
			}
		default:
			return errors.New("invalid cluster")
		}

		err := deployment.Init(cluster, name, args)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		return cluster.Deploy(deployment)
	}
}
