package framework

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/k8s"
)

type TestFramework struct {
	testing.T
	k8sclusters []string
	kumactl     string
	verbose     bool
}

const (
	Verbose = true
	Silent  = false
)

const defaultKubeConfigPathPattern = "${HOME}/.kube/kind-kuma-%d-config"

// NewK8sTest gets the number of the clusters to use in the tests, and the pattern
// to locate the KUBECONFIG for them. The second argument can be empty
func NewK8sTest(numClusters int, kubeConfigPathPattern string, verbose bool) *TestFramework {
	if numClusters < 1 || numClusters > 3 {
		log.Error("Invalid cluster number. Should be in the range [1,3], but it is ", numClusters)
		return nil
	}
	if kubeConfigPathPattern == "" {
		kubeConfigPathPattern = defaultKubeConfigPathPattern
	}
	kubeconfigs := []string{""} // have an empty cluster at [0]
	for i := 1; i <= numClusters; i++ {
		kubeconfigs = append(kubeconfigs,
			os.ExpandEnv(fmt.Sprintf(kubeConfigPathPattern, i)))
	}

	kumactl := os.Getenv("KUMACTL")
	if kumactl == "" {
		log.Error("Unable to find kumactl, please supply valid KUMACTL environment variable.")
	}

	return &TestFramework{
		k8sclusters: kubeconfigs,
		kumactl:     kumactl,
		verbose:     verbose,
	}
}

func (t *TestFramework) ApplyOnK8sCluster(idx int, namespace string, yaml string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(t, options, t.k8sclusters[idx])
	k8s.KubectlApply(t, options, yaml)
	return nil
}

func (t *TestFramework) ApplyAndWaitServiceOnK8sCluster(idx int, namespace string, yaml string, service string) error {
	options := k8s.NewKubectlOptions("", "", namespace)
	defer k8s.KubectlDelete(t, options, t.k8sclusters[idx])
	k8s.KubectlApply(t, options, yaml)
	k8s.WaitUntilServiceAvailable(t, options, service, 10, 1*time.Second)
	return nil
}

func (t *TestFramework) DeployKumaOnK8sCluster(idx int) {
	require.NoError(t, t.DeployKumaOnK8sClusterE(idx))
}

func (t *TestFramework) DeployKumaOnK8sClusterE(idx int) error {
	options := NewKumactlOptions("", "", t.verbose)

	err := k8s.KubectlApplyFromStringE(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
		},
		KumactlInstallCP(t, options))
	if err != nil {
		return err
	}

	k8s.WaitUntilServiceAvailable(t,
		&k8s.KubectlOptions{
			ConfigPath: t.k8sclusters[idx],
			Namespace:  "kuma-system",
		},
		"kuma-control-plane", 10, 1*time.Second)

	return err
}
