package framework

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

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

	seenErrors := []bool{
		debugUniversalCopyLogs(debugDir),
		debugUniversalExport(cluster, mesh, debugDir, kumactlOpts),
		debugUniversalInspectDPs(cluster, mesh, debugDir, kumactlOpts),
	}

	if slices.Contains(seenErrors, true) {
		Logf("[WARNING]: some debug commands failed")
	}
}

func debugUniversalCopyLogs(debugPath string) bool {
	currPath := universal_logs.GetLogsPath(
		ginkgo.CurrentSpecReport(),
		Config.UniversalE2ELogsPath,
	).Describe
	copyPath := filepath.Join(debugPath, "logs")

	logMsg := fmt.Sprintf("copying logs from %q to %q", currPath, copyPath)

	Logf(logMsg)

	if err := CopyDirectory(currPath, copyPath); err != nil {
		Logf("%s failed with error: %s", logMsg, err)
		return true
	}

	return false
}

func debugUniversalExport(
	cluster Cluster,
	mesh string,
	debugPath string,
	kumactlOpts kumactl.KumactlOptions,
) bool {
	var errorSeen bool

	filePath := filepath.Join(
		debugPath,
		fmt.Sprintf("%s-export.yaml", cluster.Name()),
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
	debugPath string,
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
				debugPath,
				fmt.Sprintf("%s-inspect-dataplane-%s-%s%s", cluster.Name(), dpName, typ, extension),
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
		out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "all,kuma,gateway-api", "-oyaml")
		if err != nil {
			out = fmt.Sprintf("kubectl get for namespace %s failed with error: %s", namespace, err)
			errorSeen = true
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

// When we'll update our package to Go 1.23, below helper functions are can be replaced with
// err = os.CopyFS(destDir, os.DirFS(srcDir))
// ref#1. https://stackoverflow.com/a/56314145
// ref#2. https://github.com/golang/go/issues/62484

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return errors.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0o755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer in.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return errors.Wrapf(err, "failed to create directory: '%s'", dir)
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
