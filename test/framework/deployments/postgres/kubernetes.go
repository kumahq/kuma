package postgres

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/helm"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	envVars map[string]string
	options *deployOptions
}

const (
	releaseName = "postgres-release"
	chart       = "oci://registry-1.docker.io/bitnamicharts/postgresql"
)

func (t *k8SDeployment) GetEnvVars() map[string]string {
	return t.envVars
}

var _ PostgresDeployment = &k8SDeployment{}

func (t *k8SDeployment) Name() string {
	return t.options.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	helmOpts := &helm.Options{
		Version: "12.6.0",
		SetValues: map[string]string{
			"global.postgresql.auth.postgresPassword": t.options.postgresPassword,
			"global.postgresql.auth.username":         t.options.username,
			"global.postgresql.auth.password":         t.options.password,
			"global.postgresql.auth.database":         t.options.database,
			"postgresql.primary.fullname":             t.options.primaryName,
		},
		KubectlOptions: cluster.GetKubectlOptions(t.options.namespace),
	}
	if t.options.initScript != "" {
		initScriptCfg := fmt.Sprintf(`
kind: ConfigMap
apiVersion: v1
metadata:
  name: init-script
  namespace: %s
data:
  0001-load.sql: "%s"
`, t.options.namespace, t.options.initScript)
		if err := cluster.Install(framework.YamlK8s(initScriptCfg)); err != nil {
			return err
		}
		helmOpts.SetValues["primary.initdb.scriptsConfigMap"] = "init-script"
		helmOpts.SetValues["primary.initdb.user"] = "postgres"
		helmOpts.SetValues["primary.initdb.password"] = t.options.postgresPassword
	}

	return helm.InstallE(cluster.GetTesting(), helmOpts, chart, releaseName)
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	// we delete the namespace anyway and helm.DeleteE is flaky here
	return nil
}

func NewK8SDeployment(opts *deployOptions) *k8SDeployment {
	return &k8SDeployment{
		options: opts,
	}
}
