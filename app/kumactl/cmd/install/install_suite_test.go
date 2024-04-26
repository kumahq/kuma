package install_test

import (
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers/golden"
	"github.com/kumahq/kuma/pkg/util/data"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

func TestInstallCmd(t *testing.T) {
	test.RunSpecs(t, "Install Cmd Suite")
}

var (
	backupBuildInfo kuma_version.BuildInfo
	_               = ginkgo.BeforeSuite(func() {
		backupBuildInfo = kuma_version.Build
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
	})
)

var _ = ginkgo.AfterSuite(func() {
	kuma_version.Build = backupBuildInfo
})

func ExpectMatchesGoldenFiles(actual []byte, goldenFilePath string) {
	actualManifests := data.SplitYAML(data.File{Data: actual})

	if golden.UpdateGoldenFiles() {
		if actual[len(actual)-1] != '\n' {
			actual = append(actual, '\n')
		}
		err := os.WriteFile(goldenFilePath, actual, 0o600)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}
	expected, err := os.ReadFile(goldenFilePath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	expectedManifests := data.SplitYAML(data.File{Data: expected})

	gomega.Expect(actualManifests).To(gomega.HaveLen(len(expectedManifests)), golden.RerunMsg(goldenFilePath))
	for i := range expectedManifests {
		gomega.Expect(actualManifests[i]).To(gomega.MatchYAML(expectedManifests[i]), golden.RerunMsg(goldenFilePath))
	}
}
