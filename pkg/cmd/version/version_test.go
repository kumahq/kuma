package version_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	. "github.com/kumahq/kuma/pkg/cmd/version"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("version", func() {

	var backupBuildInfo kuma_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = kuma_version.Build
	})
	AfterEach(func() {
		kuma_version.Build = backupBuildInfo
	})

	var rootCmd *cobra.Command
	var buf *bytes.Buffer

	BeforeEach(func() {
		rootCmd = &cobra.Command{
			Use: "app",
		}
		rootCmd.AddCommand(NewVersionCmd())

		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	type testCase struct {
		args     []string
		expected string
	}

	DescribeTable("should format output properly",
		func(given testCase) {
			// setup
			kuma_version.Build = kuma_version.BuildInfo{
				Version:   "1.2.3",
				GitTag:    "v1.2.3",
				GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
				BuildDate: "2019-08-07T11:26:06Z",
			}

			// given
			rootCmd.SetArgs(given.args)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(given.expected)))
		},
		Entry("app version", testCase{
			args: []string{"version"},
			expected: `Kuma: 1.2.3
`,
		}),
		Entry("app version --detailed", testCase{
			args: []string{"version", "--detailed"},
			expected: `
Product:    Kuma
Version:    1.2.3
Git Tag:    v1.2.3
Git Commit: 91ce236824a9d875601679aa80c63783fb0e8725
Build Date: 2019-08-07T11:26:06Z
`,
		}),
		Entry("app version -a", testCase{
			args: []string{"version", "-a"},
			expected: `
Product:    Kuma
Version:    1.2.3
Git Tag:    v1.2.3
Git Commit: 91ce236824a9d875601679aa80c63783fb0e8725
Build Date: 2019-08-07T11:26:06Z
`,
		}),
	)
})
