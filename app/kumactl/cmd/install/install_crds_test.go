package install_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/app/kumactl/pkg/test"
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

	It("should write CRD resources to the output file", func() {
		// given
		outputFile := filepath.Join(GinkgoT().TempDir(), "crds.yaml")
		stdout, stderr, rootCmd := test.DefaultTestingRootCmd("install", "crds", "--output-file", outputFile)

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(BeEmpty())

		// and
		actual, err := os.ReadFile(outputFile)
		Expect(err).ToNot(HaveOccurred())
		ExpectMatchesGoldenFiles(actual, filepath.Join("testdata", "install-crds.all.golden.yaml"))
	})
})
