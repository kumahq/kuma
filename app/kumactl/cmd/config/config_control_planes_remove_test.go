package config_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/kumahq/kuma/app/kumactl/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("kumactl config control-planes remove", func() {

	var configFile *os.File

	BeforeEach(func() {
		var err error
		configFile, err = ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		if configFile != nil {
			Expect(os.Remove(configFile.Name())).To(Succeed())
		}
	})

	var rootCmd *cobra.Command
	var outbuf, errbuf *bytes.Buffer

	BeforeEach(func() {
		rootCmd = cmd.DefaultRootCmd()
		outbuf, errbuf = &bytes.Buffer{}, &bytes.Buffer{}
		rootCmd.SetOut(outbuf)
		rootCmd.SetErr(errbuf)
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
			Expect(outbuf.String()).To(Equal(`Error: required flag(s) "name" not set
`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
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
			Expect(outbuf.String()).To(Equal(`Error: there is no Control Plane with name "example"
`))
			// and
			Expect(errbuf.Bytes()).To(BeEmpty())
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
				initial, err := ioutil.ReadFile(filepath.Join("testdata", given.configFile))
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(configFile.Name(), initial, 0600)
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
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := ioutil.ReadFile(configFile.Name())
				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(actual).To(MatchYAML(expected))
				// and
				Expect(outbuf.String()).To(Equal(strings.TrimLeftFunc(given.expectedOut, unicode.IsSpace)))
				// and
				Expect(errbuf.Bytes()).To(BeEmpty())
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
