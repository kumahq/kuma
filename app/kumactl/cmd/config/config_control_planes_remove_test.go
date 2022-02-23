package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("kumactl config control-planes remove", func() {

	var configFile *os.File

	BeforeEach(func() {
		var err error
		configFile, err = os.CreateTemp("", "")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		if configFile != nil {
			Expect(os.Remove(configFile.Name())).To(Succeed())
		}
	})

	var rootCmd *cobra.Command
	var outbuf *bytes.Buffer

	BeforeEach(func() {
		rootCmd = test.DefaultTestingRootCmd()

		// Different versions of cobra might emit errors to stdout
		// or stderr. It's too fragile to depend on precidely what
		// it does, and that's not something that needs to be tested
		// within Kuma anyway. So we just combine all the output
		// and validate the aggregate.
		outbuf = &bytes.Buffer{}
		rootCmd.SetOut(outbuf)
		rootCmd.SetErr(outbuf)
	})

	Describe("error cases", func() {

		It("should require name", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "remove"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err.Error()).To(MatchRegexp(requiredFlagNotSet("name")))
			// and
			Expect(outbuf.String()).To(ContainSubstring(`Error: required flag(s) "name" not set
`))
		})

		It("should fail to remove unknown Control Plane", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", filepath.Join("testdata", "config-control-planes-remove.01.initial.yaml"),
				"config", "control-planes", "remove",
				"--name", "example"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(MatchError(`there is no Control Plane with name "example"`))
			// and
			Expect(outbuf.String()).To(ContainSubstring(`Error: there is no Control Plane with name "example"
`))
		})
	})

	Describe("happy path", func() {

		type testCase struct {
			configFile  string
			goldenFile  string
			expectedOut string
		}

		DescribeTable("should remove an existing Control Plane",
			func(given testCase) {
				// setup
				initial, err := os.ReadFile(filepath.Join("testdata", given.configFile))
				Expect(err).ToNot(HaveOccurred())
				err = os.WriteFile(configFile.Name(), initial, 0600)
				Expect(err).ToNot(HaveOccurred())

				// given
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "remove",
					"--name", "example"})
				// when
				err = rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := os.ReadFile(configFile.Name())
				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(actual).To(MatchYAML(expected))
				// and
				Expect(outbuf.String()).To(Equal(strings.TrimLeftFunc(given.expectedOut, unicode.IsSpace)))
			},
			Entry("should remove active Control Plane", testCase{
				configFile: "config-control-planes-remove.11.initial.yaml",
				goldenFile: "config-control-planes-remove.11.golden.yaml",
				expectedOut: `
removed Control Plane "example"
switched active Control Plane to "other"
`,
			}),
			Entry("should remove non-active Control Plane", testCase{
				configFile: "config-control-planes-remove.12.initial.yaml",
				goldenFile: "config-control-planes-remove.12.golden.yaml",
				expectedOut: `
removed Control Plane "example"
switched active Control Plane to "other"
`,
			}),
			Entry("should remove the last Control Plane", testCase{
				configFile: "config-control-planes-remove.13.initial.yaml",
				goldenFile: "config-control-planes-remove.13.golden.yaml",
				expectedOut: `
removed Control Plane "example"
there is no active Control Plane left. Use ` + "`" + `kumactl config control-planes add` + "`" + ` to add a Control Plane and make it active
`,
			}),
		)
	})
})
