package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/test/framework/kumactl"
	"github.com/kumahq/kuma/test/framework/universal_logs"
)

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
// * CP logs (although we print this already on failure)
func DebugUniversal(cluster Cluster, mesh string) {
	ginkgo.GinkgoHelper()

	debugDir := prepareDebugDir()

	Logf("printing debug information of cluster %q for mesh %q", cluster.Name(), mesh)
	// we don't have command to print policies for given mesh, so it's better to print all than none.
	kumactlOpts := *cluster.GetKumactlOptions()
	kumactlOpts.Verbose = false

	errs := slices.Concat(
		debugUniversalCopyLogs(debugDir),
		debugUniversalExport(cluster, mesh, debugDir, kumactlOpts),
		debugUniversalInspectDPs(cluster, mesh, debugDir, kumactlOpts),
	)

	for _, err := range errs {
		Logf("[WARNING]: %s", err)
	}
}

func DebugUniversalCPLogs(cluster Cluster) {
	ginkgo.GinkgoHelper()

	debugDir := prepareDebugDir()

	logs, err := cluster.GetKumaCPLogs()
	if err != nil {
		Logf("[WARNING]: could not retrieve cp logs, error: %s", err.Error())
	} else {
		cpLogsExportPath := filepath.Join(debugDir, fmt.Sprintf("cp-log-%s.log", cluster.Name()))
		Logf("saving CP logs of cluster %q to %q", cluster.Name(), cpLogsExportPath)
		Expect(os.WriteFile(cpLogsExportPath, []byte(logs), 0o600)).To(Succeed())
	}
}

func debugUniversalCopyLogs(debugPath string) []error {
	srcPath := universal_logs.GetLogsPath(
		ginkgo.CurrentSpecReport(),
		Config.UniversalE2ELogsPath,
	).Describe
	destPath := filepath.Join(debugPath, "logs")

	Logf("copying logs from %q to %q", srcPath, destPath)

	if err := os.CopyFS(destPath, os.DirFS(srcPath)); err != nil {
		return []error{errors.Wrapf(err, "failed to copy logs from %q to %q", srcPath, destPath)}
	}

	return nil
}

func debugUniversalExport(
	cluster Cluster,
	mesh string,
	debugPath string,
	kumactlOpts kumactl.KumactlOptions,
) []error {
	var errs []error

	filePath := filepath.Join(
		debugPath,
		fmt.Sprintf("%s-export.yaml", cluster.Name()),
	)

	Logf("saving export for cluster %q and mesh %q to file %q", cluster.Name(), mesh, filePath)

	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	if err != nil {
		wrappedErr := errors.Wrap(err, "failed to run 'kumactl export --profile all'")
		errs = append(errs, wrappedErr)
		out = fmt.Sprintf("# export failed: %s", wrappedErr)
	}

	if err := os.WriteFile(filePath, []byte(out), 0o600); err != nil {
		errs = append(errs, errors.Wrapf(err, "failed to write export to file %q", filePath))
	}

	return errs
}

func debugUniversalInspectDPs(
	cluster Cluster,
	mesh string,
	debugPath string,
	kumactlOpts kumactl.KumactlOptions,
) []error {
	var errs []error

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
				err := errors.Wrapf(err, "kumactl inspect dataplane %s --mesh %s --type %s failed", dpName, mesh, typ)
				errs = append(errs, err)
				out = fmt.Sprintf("%q", err)
			}

			filePath := filepath.Join(
				debugPath,
				fmt.Sprintf("%s-inspect-dataplane-%s-%s%s", cluster.Name(), dpName, typ, extension),
			)

			if err := os.WriteFile(filePath, []byte(out), 0o600); err != nil {
				errs = append(errs, errors.Wrapf(err, "failed to write file %q", filePath))
			}
		}
	}

	return errs
}

func DebugKube(cluster Cluster, mesh string, namespaces ...string) {
	ginkgo.GinkgoHelper()

	debugPath := prepareDebugDir()

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
			exportFilePath := filepath.Join(debugPath, fmt.Sprintf("%s-node-%s-%s", cluster.Name(), node.Name, uuid.New().String()))
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
		out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "all,kuma", "-oyaml")
		if err != nil {
			out = fmt.Sprintf("kubectl get for namespace %s failed with error: %s", namespace, err.Error())
			errorSeen = true
		}

		deployments, err := k8s.ListDeploymentsE(cluster.GetTesting(), &kubeOptions, kube_meta.ListOptions{})
		if err == nil {
			for _, deployment := range deployments {
				if !k8s.IsDeploymentAvailable(&deployment) {
					deployDetails := ExtractDeploymentDetails(cluster.GetTesting(), &kubeOptions, deployment.Name)
					out += MarshalObjectDetails(deployDetails)
				}
			}
		} else {
			out += fmt.Sprintf("failed to list deployments in namespace %s with error: %s", namespace, err.Error())
			errorSeen = true
		}

		// Ignore it if we don't have Gateway API resources installed
		gatewayAPIOut, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "gateway-api", "-oyaml")
		if err == nil {
			out += gatewayAPIOut
		} else {
			Logf("Gateway API CRDs not installed in cluster %q", cluster.Name())
		}

		exportFilePath := filepath.Join(debugPath, fmt.Sprintf("%s-namespace-%s-%s", cluster.Name(), namespace, uuid.New().String()))
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

	exportFilePath := filepath.Join(debugPath, fmt.Sprintf("%s-export-%s", cluster.Name(), uuid.New().String()))
	Logf("saving export of cluster %q for mesh %q to a file %q", cluster.Name(), mesh, exportFilePath)
	Expect(os.WriteFile(exportFilePath, []byte(out), 0o600)).To(Succeed())

	if errorSeen {
		Logf("[WARNING]: some debug commands failed")
	}
}

func prepareDebugDir() string {
	ginkgo.GinkgoHelper()

	path := filepath.Join(Config.DebugDir, uuid.New().String())

	Expect(os.MkdirAll(path, 0o755)).To(Or(
		Not(HaveOccurred()),
		Satisfy(os.IsNotExist),
	))

	return path
}

func CpRestarted(cluster Cluster) bool {
	switch cluster.(type) {
	case *UniversalCluster:
		// CP does not recover restart on universal. If it crashed, we can just check if the process is still running.
		out, _, _ := cluster.Exec("", "", AppModeCP, "ps", "aux")
		return !strings.Contains(out, "kuma-cp run")
	case *K8sCluster:
		restartCount := RestartCount(cluster.GetKuma().(*K8sControlPlane).GetKumaCPPods())
		return restartCount > 0
	default:
		return false
	}
}
