package framework

import "fmt"

func GatewayProxyUniversal(mesh, name string) InstallFunc {
	return func(cluster Cluster) error {
		token, err := cluster.GetKuma().GenerateDpToken(mesh, "edge-gateway")
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
      kuma.io/service: edge-gateway
`, mesh)
		return cluster.DeployApp(
			WithKumactlFlow(),
			WithName(name),
			WithMesh(mesh),
			WithToken(token),
			WithVerbose(),
			WithYaml(dataplaneYaml),
		)
	}
}
