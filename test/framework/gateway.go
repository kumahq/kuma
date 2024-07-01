package framework

import "fmt"

func GatewayProxyUniversal(mesh, name string, appDeploymentOptions ...AppDeploymentOption) InstallFunc {
	return func(cluster Cluster) error {
		token, err := cluster.GetKuma().GenerateDpToken(mesh, name)
		if err != nil {
			return err
		}

		dataplaneYaml := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  gateway:
    type: BUILTIN
    tags:
      kuma.io/service: %s
`, mesh, name)

		deploymentOpts := []AppDeploymentOption{WithName(name), WithMesh(mesh), WithToken(token), WithVerbose(), WithYaml(dataplaneYaml)}
		deploymentOpts = append(deploymentOpts, appDeploymentOptions...)
		return cluster.DeployApp(deploymentOpts...)
	}
}
