package version

import (
	"os"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

var _ = Describe("lastMinorOfMajor", func() {
	It("should match the newest release of each recorded major in versions.yml", func() {
		// given
		content, err := os.ReadFile("../../versions.yml")
		Expect(err).ToNot(HaveOccurred())
		var entries []struct {
			Version string `json:"version"`
		}
		Expect(yaml.Unmarshal(content, &entries)).To(Succeed())

		newestMinor := map[uint64]uint64{}
		for _, entry := range entries {
			v, err := semver.NewVersion(entry.Version)
			if err != nil {
				// the "preview" row has no semver
				continue
			}
			if v.Minor() > newestMinor[v.Major()] {
				newestMinor[v.Major()] = v.Minor()
			}
		}

		// then
		for major, minor := range lastMinorOfMajor {
			Expect(newestMinor).To(HaveKeyWithValue(major, minor),
				"versions.yml lists a %d.x release newer than lastMinorOfMajor[%d] = %d; update the map", major, major, minor)
		}
	})
})
