package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	. "github.com/onsi/gomega"
)

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
//
// In case we get into limits of massive logs on stdout. We can save it to file, print file names and upload the files on CI as artifacts.
func DebugUniversal(cluster Cluster, mesh string) {
	ensureDebugDir()
	Logf("printing debug information of cluster %q for mesh %q", cluster.Name(), mesh)
	// we don't have command to print policies for given mesh, so it's better to print all than none.
	kumactlOpts := *cluster.GetKumactlOptions()
	kumactlOpts.Verbose = false
	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	Expect(err).ToNot(HaveOccurred())

	exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-export-%s", cluster.Name(), uuid.New().String()))
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o755)).To(Succeed())
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
}

func DebugKube(cluster Cluster, mesh string, namespace string) {
	ensureDebugDir()
	Logf("printing debug information of cluster %q for mesh %q and namespace %q", cluster.Name(), mesh, namespace)
	kubeOptions := *cluster.GetKubectlOptions(namespace) // copy to not override fields globally
	kubeOptions.Logger = logger.Discard                  // to not print on stdout
	out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "all", "-oyaml")
	Expect(err).ToNot(HaveOccurred())

	exportFilePath := filepath.Join(Config.DebugDir, fmt.Sprintf("%s-namespace-%s-%s", cluster.Name(), namespace, uuid.New().String()))
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o755)).To(Succeed())
	Logf("saving state of the namespace %q of cluster %q to a file %q", namespace, cluster.Name(), exportFilePath)

	kumactlOpts := *cluster.GetKumactlOptions() // copy to not override fields globally
	kumactlOpts.Verbose = false                 // to not print on stdout
	out, err = kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	Expect(err).ToNot(HaveOccurred())

	exportFilePath = filepath.Join(Config.DebugDir, fmt.Sprintf("%s-export-%s", cluster.Name(), uuid.New().String()))
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o755)).To(Succeed())
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
}

func ensureDebugDir() {
	err := os.MkdirAll(Config.DebugDir, 0755)
	if err == nil || os.IsNotExist(err) {
		return
	}
	Expect(err).ToNot(HaveOccurred())
}
