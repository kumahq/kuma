package admin_test

import (
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/test/matchers"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("Sanitize ConfigDump", func() {
	type testCase struct {
		configFile string
		goldenFile string
	}

	DescribeTable("should redact sensitive information",
		func(given testCase) {
			// given
			rawConfigDump, err := os.ReadFile(filepath.Join("testdata", given.configFile))
			Expect(err).ToNot(HaveOccurred())

			// when
			sanitizedDump, err := admin.Sanitize(rawConfigDump)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(sanitizedDump).To(matchers.MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
		},
		Entry("full config", testCase{
			configFile: "full_config.json",
			goldenFile: "golden.full_config.json",
		}),
		Entry("no hds", testCase{
			configFile: "no_hds.json",
			goldenFile: "golden.no_hds.json",
		}),
	)
})
