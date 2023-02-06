package postgres

import (
	"github.com/gruntwork-io/terratest/modules/helm"

	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	envVars map[string]string
	options *deployOptions
}

const releaseName = "postgres-release"

func (t *k8SDeployment) GetEnvVars() map[string]string {
	return t.envVars
}

var _ PostgresDeployment = &k8SDeployment{}

func (t *k8SDeployment) Name() string {
	return t.options.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	helmOpts := &helm.Options{
		SetValues: map[string]string{
			"global.postgresql.auth.username": t.options.username,
			"global.postgresql.auth.password": t.options.password,
			"global.postgresql.auth.database": t.options.database,
			"postgresql.primary.fullname":     t.options.primaryName,
		},
		KubectlOptions: cluster.GetKubectlOptions(t.options.namespace),
	}

	err := helm.AddRepoE(cluster.GetTesting(), helmOpts, "bitnami", "https://charts.bitnami.com/bitnami")
	if err != nil {
		return err
	}

	return helm.InstallE(cluster.GetTesting(), helmOpts, "bitnami/postgresql", releaseName)
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
