package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/kumactl"
)

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
// * CP logs (although we print this already on failure)
func DebugUniversal(cluster Cluster, mesh string) {
	ginkgo.GinkgoHelper()

	Expect(ensureDebugDir()).To(Succeed())

	id := uuid.New().String()

	Logf("printing debug information of cluster %q for mesh %q", cluster.Name(), mesh)
	// we don't have command to print policies for given mesh, so it's better to print all than none.
	kumactlOpts := *cluster.GetKumactlOptions()
	kumactlOpts.Verbose = false

	seenErrors := []bool{
		debugUniversalExport(cluster, mesh, id, kumactlOpts),
		debugUniversalInspectDPs(cluster, mesh, id, kumactlOpts),
	}

	Expect(seenErrors).ToNot(ContainElement(true), "some debug commands failed")
}

func debugUniversalExport(
	cluster Cluster,
	mesh string,
	id string,
	kumactlOpts kumactl.KumactlOptions,
) bool {
	var errorSeen bool

	filePath := filepath.Join(
		Config.DebugDir,
		fmt.Sprintf("%s-export-%s.yaml", cluster.Name(), id),
	)

	logMsg := fmt.Sprintf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, filePath)

	Logf(logMsg)

	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	if err != nil {
		// We don't want to fail in the middle.
		errorSeen = true
		msg := fmt.Sprintf("kumactl export --profile all failed with error: %s", err)
		out = fmt.Sprintf("# %s", msg)
		Logf(fmt.Sprintf("%s failed: %s", logMsg, msg))
	}

	if err := os.WriteFile(filePath, []byte(out), 0o600); err != nil {
		errorSeen = true
		Logf("%s failed with error: %s", logMsg, err)
	}

	return errorSeen
}

func debugUniversalInspectDPs(
	cluster Cluster,
	mesh string,
	id string,
	kumactlOpts kumactl.KumactlOptions,
) bool {
	var errorSeen bool

	Logf("saving dataplane inspections from cluster %q for mesh %q", cluster.Name(), mesh)

	for _, dpName := range cluster.(*UniversalCluster).GetDataplanes() {
		for typ, extension := range map[string]string{
			"config-dump": ".json",
			"config":      ".json",
			"policies":    "",
			"stats":       "",
			"clusters":    "",
		} {
			var out string
			var err error

			if out, err = kumactlOpts.RunKumactlAndGetOutput(
				"inspect", "dataplane", dpName,
				"--mesh", mesh,
				"--type", typ,
			); err != nil {
				// We don't want to fail in the middle.
				errorSeen = true
				msg := fmt.Sprintf("kumactl inspect dataplane %s --mesh %s --type %s failed with error: %s", dpName, mesh, typ, err)
				out = fmt.Sprintf("%q", msg)
				Logf(msg)
			}

			filePath := filepath.Join(
				Config.DebugDir,
				fmt.Sprintf("%s-inspect-dataplane-%s-%s-%s%s", cluster.Name(), typ, dpName, id, extension),
			)

			if err := os.WriteFile(filePath, []byte(out), 0o600); err != nil {
				errorSeen = true
				Logf("saving %q for dataplane %q to file %q failed with error: %s", typ, dpName, filePath, err)
			}
		}
	}

	return errorSeen
}

func DebugKube(cluster Cluster, mesh string, namespaces ...string) {
	ginkgo.GinkgoHelper()

	Expect(ensureDebugDir()).To(Succeed())

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
	Expect(errorSeen).NotTo(BeTrue(), "some debug commands failed")
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
}

func ensureDebugDir() error {
	if err := os.MkdirAll(Config.DebugDir, 0o755); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
