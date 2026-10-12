package version

import (
	"os"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

var _ = Describe("lastMinorOfMajor", func() {
	It("should match the newest release of every major in versions.yml", func() {
		content, err := os.ReadFile("../../versions.yml")
		Expect(err).ToNot(HaveOccurred())

		var entries []struct {
			Version string `json:"version"`
		}
		Expect(yaml.Unmarshal(content, &entries)).To(Succeed())

		newestMinor := map[uint64]uint64{}
		for _, entry := range entries {
			release, err := semver.NewVersion(entry.Version)
			if err != nil {
				continue
			}
			if release.Minor() > newestMinor[release.Major()] {
				newestMinor[release.Major()] = release.Minor()
			}
		}

		Expect(lastMinorOfMajor).To(Equal(newestMinor),
			"pkg/version/zz_generated.compatibility.go is out of date; run make generate/versions")
	})
})
