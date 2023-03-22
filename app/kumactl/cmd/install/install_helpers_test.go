package install_test

import (
	"os"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/pkg/test/matchers/golden"
)

func ExpectMatchesGoldenFiles(actual []byte, goldenFilePath string) {
	actualManifests := data.SplitYAML(data.File{Data: actual})

	if golden.UpdateGoldenFiles() {
		if actual[len(actual)-1] != '\n' {
			actual = append(actual, '\n')
		}
		err := os.WriteFile(goldenFilePath, actual, 0o600)
		Expect(err).ToNot(HaveOccurred())
	}
	expected, err := os.ReadFile(goldenFilePath)
	Expect(err).ToNot(HaveOccurred())
	expectedManifests := data.SplitYAML(data.File{Data: expected})

	Expect(actualManifests).To(HaveLen(len(expectedManifests)), golden.RerunMsg(goldenFilePath))
	for i := range expectedManifests {
		Expect(actualManifests[i]).To(MatchYAML(expectedManifests[i]), golden.RerunMsg(goldenFilePath))
	}
}
