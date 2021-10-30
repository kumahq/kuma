package install_test

import (
	"io/ioutil"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/pkg/test/golden"
)

func ExpectMatchesGoldenFiles(actual []byte, goldenFilePath string) {
	actualManifests := data.SplitYAML(data.File{Data: actual})

	if golden.UpdateGoldenFiles() {
		if actual[len(actual)-1] != '\n' {
			actual = append(actual, '\n')
		}
		err := ioutil.WriteFile(goldenFilePath, actual, 0664)
		Expect(err).ToNot(HaveOccurred())
	}
	expected, err := ioutil.ReadFile(goldenFilePath)
	Expect(err).ToNot(HaveOccurred())
	expectedManifests := data.SplitYAML(data.File{Data: expected})

	Expect(len(actualManifests)).To(Equal(len(expectedManifests)), golden.RerunMsg)
	for i := range expectedManifests {
		Expect(actualManifests[i]).To(MatchYAML(expectedManifests[i]), golden.RerunMsg)
	}
}
