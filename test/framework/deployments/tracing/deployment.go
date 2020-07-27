package tracing

import "github.com/kumahq/kuma/test/framework"

type Tracing interface {
	ZipkinCollectorURL() string
	TracedServices() ([]string, error)
}

type Deployment interface {
	framework.Deployment
	Tracing
}

const DeploymentName = "tracing"

func From(cluster framework.Cluster) Tracing {
	return cluster.Deployment(DeploymentName).(Tracing)
}

func Install() framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		switch cluster.(type) {
		case *framework.K8sCluster:
			k8sCluster := cluster.(*framework.K8sCluster)
			deployment := &k8SDeployment{}
			k8sCluster.Deployments[DeploymentName] = deployment
			if err := deployment.Deploy(cluster); err != nil {
				return err
			}
		case *framework.UniversalCluster:
			universalCluster := cluster.(*framework.UniversalCluster)
			deployment := &universalDeployment{
				ports: map[string]string{},
			}
			universalCluster.Deployments[DeploymentName] = deployment
			if err := deployment.Deploy(cluster); err != nil {
				return err
			}
		}
		return nil
	}
}
