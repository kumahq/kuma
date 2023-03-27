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

type Command []string

const (
	DeploymentName = "externalservice-"
	HttpServer     = "http-server"
	HttpsServer    = "https-server"
	TcpSink        = "tcp-sink"
)

func From(cluster framework.Cluster, name string) ExternalService {
	return cluster.Deployment(DeploymentName + name).(ExternalService)
}

func Install(name string, commands ...Command) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		var deployment Deployment
		if len(commands) < 1 {
			return errors.New("command list can't be empty")
		}
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8SDeployment{
				name: name,
				cmd:  commands[0],
			}
		case *framework.UniversalCluster:
			deployment = NewUniversalDeployment().
				WithName(name).
				WithCommands(commands...).
				WithVerbose(cluster.Verbose())
		default:
			return errors.New("invalid cluster")
		}

		return cluster.Deploy(deployment)
	}
}
