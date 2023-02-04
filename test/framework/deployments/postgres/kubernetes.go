package postgres

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/kumahq/kuma/test/framework"
)

type k8SDeployment struct {
	deploymentName  string
	namespace       string
	envVars   map[string]string
}


func (t *k8SDeployment) GetEnvVars() map[string]string {
	return t.envVars
}

var _ PostgresDeployment = &k8SDeployment{}

func (t *k8SDeployment) Name() string {
	return t.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	releaseName := "postgres-release"

	helmOpts := &helm.Options{
		SetValues: map[string]string{
			"global.postgresql.auth.username": "mesh",
			"global.postgresql.auth.password": "mesh",
			"global.postgresql.auth.database": "mesh",
			"postgresql.primary.fullname": "postgres",
		},
		KubectlOptions: cluster.GetKubectlOptions("kuma-system"),

		ExtraArgs: map[string][]string{
			"--create-namespace": nil,
		},
	}

	err := helm.AddRepoE(cluster.GetTesting(), helmOpts, "bitnami", "https://charts.bitnami.com/bitnami")
	if err != nil {
		return err
	}

	return helm.InstallE(cluster.GetTesting(), helmOpts, "bitnami/postgresql", releaseName)
}

func (t *k8SDeployment) Delete(cluster framework.Cluster) error {
	return cluster.DeleteNamespace(t.namespace)
}

func NewK8SDeployment(opts *deployOptions) *k8SDeployment {
	return &k8SDeployment{
		deploymentName: opts.deploymentName,
		namespace:      opts.namespace,
		envVars:        nil,
	}
}
