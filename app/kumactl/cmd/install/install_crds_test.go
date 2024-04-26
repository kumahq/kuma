package install_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/test"
)

var _ = Context("kumactl install crds", func() {
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate Kubernetes CRD resources",
		func(given testCase) {
			// given
			args := append([]string{"install", "crds"}, given.extraArgs...)
			stdout, stderr, rootCmd := test.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(stderr.String()).To(BeEmpty())

			// when
			actual := stdout.Bytes()
			ExpectMatchesGoldenFiles(actual, filepath.Join("testdata", given.goldenFile))
		},
		Entry("should generate all Kuma's CRD resources", testCase{
			extraArgs:  nil,
			goldenFile: "install-crds.all.golden.yaml",
		}),
	)
})
