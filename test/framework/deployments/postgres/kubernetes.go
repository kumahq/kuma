package postgres

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"

	"github.com/kumahq/kuma/v2/test/framework"
)

// NOTE: We intentionally do not use the tag form like
// "oci://ghcr.io/cloudnative-pg/charts/cloudnative-pg:<version>@<digest>"
// for these charts because the tags appear to be mutable and sometimes
// get overridden by the maintainers with new digests. When that happens,
// our CI fails because it tries to fetch a tag with a matching digest,
// but the tag now resolves to a different digest and the pull cannot
// satisfy both. To avoid these CI failures, we pin the charts by
// immutable digest only
const (
	cnpgChart    = "oci://ghcr.io/cloudnative-pg/charts/cloudnative-pg@sha256:b294ea82771c9049b2f1418a56cbab21716343fd44fe68721967c95ca7f5c523" // 0.26.0
	clusterChart = "oci://ghcr.io/cloudnative-pg/charts/cluster@sha256:3f4f1a26dc0388f47bc456e0ec733255c1a8469b0742ce052df3885ba935c388"        // 0.3.1
)

// We must keep a tag with version here because CloudNativePG requires and checks
// for a tag during installation. Additionally, we use tags that include a date
// because those are the only stable tags provided by the project. To ensure
// immutability and reproducibility, we still pin by digest, using the
// <tag>@<digest> form
const postgresImage = "ghcr.io/cloudnative-pg/postgresql:18.0-202509290807-minimal-trixie@sha256:dd7c678167cc6d06c2caf4e6ea7bc7a89e39754bc7e0111a81f5d75ac3068f70"

type k8SDeployment struct {
	envVars map[string]string
	options *deployOptions
}

func (t *k8SDeployment) GetEnvVars() map[string]string {
	return t.envVars
}

var _ PostgresDeployment = &k8SDeployment{}

func (t *k8SDeployment) Name() string {
	return t.options.deploymentName
}

func (t *k8SDeployment) Deploy(cluster framework.Cluster) error {
	extraArgs := map[string][]string{"upgrade": {"--create-namespace", "--install", "--wait"}}

	if err := helm.UpgradeE(
		cluster.GetTesting(),
		&helm.Options{
			KubectlOptions: cluster.GetKubectlOptions(t.options.namespace),
			ExtraArgs:      extraArgs,
		},
		cnpgChart,
		"cnpg",
	); err != nil {
		return err
	}

	if err := k8s.KubectlApplyFromStringE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(t.options.namespace),
		dbSecrets(
			t.options.namespace,
			"postgres",
			t.options.postgresPassword,
			t.options.username,
			t.options.password,
		),
	); err != nil {
		return err
	}

	opts := &helm.Options{
		SetValues: map[string]string{
			"cluster.imageName":                        postgresImage,
			"cluster.instances":                        "1",
			"cluster.storage.size":                     "100Mi",
			"cluster.initdb.database":                  t.options.database,
			"cluster.initdb.owner":                     t.options.username,
			"cluster.initdb.secret.name":               fmt.Sprintf("db-%s-secret", t.options.username),
			"cluster.superuserSecret":                  "db-postgres-secret",
			"cluster.services.disabledDefaultServices": "{r,ro}",
		},
		KubectlOptions: cluster.GetKubectlOptions(t.options.namespace),
		ExtraArgs:      extraArgs,
	}

	for i, script := range t.options.initScripts {
		opts.SetValues[fmt.Sprintf("cluster.initdb.postInitSQL[%d]", i)] = script
	}

	return helm.UpgradeE(cluster.GetTesting(), opts, clusterChart, t.options.primaryName)
}

func (t *k8SDeployment) Delete(framework.Cluster) error {
	// we delete the namespace anyway and helm.DeleteE is flaky here
	return nil
}

func NewK8SDeployment(opts *deployOptions) PostgresDeployment {
	return &k8SDeployment{
		options: opts,
	}
}

func dbSecrets(namespace string, userPassPairs ...string) string {
	var result []string

	if len(userPassPairs)%2 != 0 {
		panic("userPassPairs must be a multiple of 2")
	}

	for pair := range slices.Chunk(userPassPairs, 2) {
		result = append(result, dbSecret(namespace, pair[0], pair[1]))
	}

	return strings.Join(result, "---\n")
}

func dbSecret(namespace, username, password string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Secret
type: kubernetes.io/basic-auth
metadata:
  name: db-%[2]s-secret
  namespace: %[1]s
stringData:
  username: %[2]s
  password: %[3]s
`, namespace, username, password)
}
