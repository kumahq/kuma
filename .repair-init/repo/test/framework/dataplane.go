package framework

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// IsDataplaneOnline returns online, found, error
func IsDataplaneOnline(cluster Cluster, mesh, name string) (bool, bool, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh)
	if err != nil {
		return false, false, err
	}
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, name) {
			return strings.Contains(line, "Online"), true, nil
		}
	}
	return false, false, nil
}

func DataplaneReceivedConfig(cluster Cluster, mesh, name string) (bool, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh, "-o", "yaml", name)
	if err != nil {
		return false, err
	}
	return strings.Contains(out, `responsesAcknowledged`), nil
}

// WaitDataplaneInspectable polls until `kumactl inspect dataplane` succeeds for
// the given DPP, which requires its DataplaneInsight to be published. Useful
// before running table-driven envoyconfig tests so per-entry polling windows
// don't have to absorb the registration delay.
func WaitDataplaneInspectable(cluster Cluster, mesh, dpp string) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		_, err := cluster.GetKumactlOptions().
			RunKumactlAndGetOutput("inspect", "dataplane", dpp, "--type", "config", "--mesh", mesh)
		g.Expect(err).ToNot(HaveOccurred())
	}, "90s", "2s").Should(Succeed())
}
