package admin_test

import (
	"os"
	"path"
	"path/filepath"

	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

			configDump := &envoy_admin_v3.ConfigDump{}
			Expect(util_proto.FromJSON(rawConfigDump, configDump)).To(Succeed())

			// when
			Expect(admin.Sanitize(configDump)).To(Succeed())
			// and when
			sanitized, err := util_proto.ToJSONIndent(configDump, "  ")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(sanitized).To(matchers.MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
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
