package install_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("kumactl install crds", func() {
	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate Kubernetes CRD resources",
		func(given testCase) {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"install", "crds"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

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
		Entry("should generate all Kuma's CRD resources including experimental meshgateway", testCase{
			extraArgs:  []string{"--experimental-meshgateway"},
			goldenFile: "install-crds.experimental-meshgateway.golden.yaml",
		}),
	)
})
