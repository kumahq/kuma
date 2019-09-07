package version_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/cmd/version"

	"github.com/spf13/cobra"

	konvoy_version "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/version"
)

var _ = Describe("version", func() {

	var backupBuildInfo konvoy_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = konvoy_version.Build
	})
	AfterEach(func() {
		konvoy_version.Build = backupBuildInfo
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
		buildInfo konvoy_version.BuildInfo
		args      []string
		expected  string
	}

	DescribeTable("should format output properly",
		func(given testCase) {
			// setup
			konvoy_version.Build = konvoy_version.BuildInfo{
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
			args:     []string{"version"},
			expected: `1.2.3`,
		}),
		Entry("app version --detailed", testCase{
			args: []string{"version", "--detailed"},
			expected: `
Version:    1.2.3
Git Tag:    v1.2.3
Git Commit: 91ce236824a9d875601679aa80c63783fb0e8725
Build Date: 2019-08-07T11:26:06Z
`,
		}),
		Entry("app version -a", testCase{
			args: []string{"version", "-a"},
			expected: `
Version:    1.2.3
Git Tag:    v1.2.3
Git Commit: 91ce236824a9d875601679aa80c63783fb0e8725
Build Date: 2019-08-07T11:26:06Z
`,
		}),
	)
})
