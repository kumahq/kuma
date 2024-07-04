package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	. "github.com/onsi/gomega"
)

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
// * CP logs (although we print this already on failure)
func DebugUniversal(cluster Cluster, mesh string) {
	ensureDebugDir()
	errorSeen := false
	Logf("printing debug information of cluster %q for mesh %q", cluster.Name(), mesh)
	// we don't have command to print policies for given mesh, so it's better to print all than none.
	kumactlOpts := *cluster.GetKumactlOptions()
	kumactlOpts.Verbose = false
	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	if err != nil {
		// We don't want to fail in the middle.
		out = fmt.Sprintf("kumactl export failed with error: %s", err)
		errorSeen = true
	}

	exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-export-%s", cluster.Name(), uuid.New().String()))
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o600)).To(Succeed())
	Expect(errorSeen).ToNot(BeTrue(), "some debug commands failed")
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
}

func DebugKube(cluster Cluster, mesh string, namespaces ...string) {
	ensureDebugDir()
	errorSeen := false

	Logf("debug nodes and print resource usage of cluster %q", cluster.Name())
	defaultKubeOptions := *cluster.GetKubectlOptions("default") // copy to not override fields globally
	defaultKubeOptions.Logger = logger.Discard
	nodes, err := k8s.GetNodesE(cluster.GetTesting(), &defaultKubeOptions)
	if err != nil {
		Logf("get nodes from cluster %q failed with error: %s", cluster.Name(), err)
		errorSeen = true
	} else {
		for _, node := range nodes {
			exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-node-%s-%s", cluster.Name(), node.Name, uuid.New().String()))
			out, e := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "describe", "node", node.Name)
			if e != nil {
				Logf("kubectl describe node %s failed with error: %s", node.Name, err)
				errorSeen = true
			} else {
				Expect(os.WriteFile(exportFilePath, []byte(out), 0o600)).To(Succeed())
				Logf("saving state of the node %q of cluster %q to a file %q", node.Name, cluster.Name(), exportFilePath)
			}
		}
	}

	Logf("printing debug information of cluster %q for mesh %q and namespaces %q", cluster.Name(), mesh, namespaces)
	for _, namespace := range namespaces {
		kubeOptions := *cluster.GetKubectlOptions(namespace) // copy to not override fields globally
		kubeOptions.Logger = logger.Discard                  // to not print on stdout
		out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "all,kuma,gateway-api", "-oyaml")
		if err != nil {
			out = fmt.Sprintf("kubectl get for namespace %s failed with error: %s", namespace, err)
			errorSeen = true
		}

		exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-namespace-%s-%s", cluster.Name(), namespace, uuid.New().String()))
		Expect(os.WriteFile(exportFilePath, []byte(out), 0o600)).To(Succeed())
		Logf("saving state of the namespace %q of cluster %q to a file %q", namespace, cluster.Name(), exportFilePath)
	}

	kumactlOpts := *cluster.GetKumactlOptions() // copy to not override fields globally
	kumactlOpts.Verbose = false                 // to not print on stdout
	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	if err != nil {
		out = fmt.Sprintf("kumactl export failed with error: %s", err)
		errorSeen = true
	}

	exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-export-%s", cluster.Name(), uuid.New().String()))
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o600)).To(Succeed())
	Expect(errorSeen).To(BeTrue(), "some debug commands failed")
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
}

func ensureDebugDir() {
	err := os.MkdirAll(Config.DebugDir, 0o755)
	if err == nil || os.IsNotExist(err) {
		return
	}
	Expect(err).ToNot(HaveOccurred())
}

func CpRestarted(cluster Cluster) bool {
	switch cluster.(type) {
	case *UniversalCluster:
		out, _, _ := cluster.Exec("", "", AppModeCP, "ps", "aux")
		return strings.Contains(out, "kuma-cp run")
	case *K8sCluster:
		restartCount := RestartCount(cluster.GetKuma().(*K8sControlPlane).GetKumaCPPods())
		return restartCount == 0
	default:
		return false
	}
}
