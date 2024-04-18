package framework

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
)

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
//
// In case we get into limits of massive logs on stdout. We can save it to file, print file names and upload the files on CI as artifacts.
func DebugUniversal(cluster Cluster, mesh string) {
	Logf("printing debug information of cluster %q for mesh %q", cluster.Name(), mesh)
	// we don't have command to print policies for given mesh, so it's better to print all than none.
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("export", "--profile", "all")
	Expect(err).ToNot(HaveOccurred())
	Logf(out)
}

func DebugKube(cluster Cluster, mesh string, namespace string) {
	Logf("printing debug information of cluster %q for mesh %q and namespace %q", cluster.Name(), mesh, namespace)
	err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(namespace), "get", "all", "-oyaml")
	Expect(err).ToNot(HaveOccurred())
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("export", "--profile", "all")
	Expect(err).ToNot(HaveOccurred())
	Logf(out)
}
