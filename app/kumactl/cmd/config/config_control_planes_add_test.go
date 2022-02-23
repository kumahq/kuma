package config_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/util/test"
	"github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("kumactl config control-planes add", func() {

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
				"config", "control-planes", "add"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err.Error()).To(MatchRegexp(requiredFlagNotSet("name")))
			// and
			Expect(outbuf.String()).To(ContainSubstring(`Error: required flag(s) "address", "name" not set
`))
		})

		It("should require API Server URL", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "add",
				"--name", "example"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err.Error()).To(MatchRegexp(requiredFlagNotSet("address")))
			// and
			Expect(outbuf.String()).To(ContainSubstring(`Error: required flag(s) "address" not set
`))
		})

		It("should not allow invalid auth-type", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "add",
				"--address", "http://localhost:1234",
				"--auth-type", "unknown",
				"--name", "example"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(MatchError(`authentication plugin of type "unknown" is not found`))
		})

		It("should fail to add a new Control Plane with duplicate name", func() {
			// setup
			server, port := setupCpIndexServer()
			defer server.Close()

			// given
			rootCmd.SetArgs([]string{"--config-file", filepath.Join("testdata", "config-control-planes-add.01.golden.yaml"),
				"config", "control-planes", "add",
				"--name", "example",
				"--address", fmt.Sprintf("http://localhost:%d", port)})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(MatchError(`Control Plane with name "example" already exists. Use --overwrite to replace an existing one.`))
			// and
			Expect(outbuf.String()).To(Equal(`Error: Control Plane with name "example" already exists. Use --overwrite to replace an existing one.
`))
		})

		It("should fail when CP timeouts", func() {
			// setup
			server, port := setupCpServer(func(writer http.ResponseWriter, req *http.Request) {
				time.Sleep(time.Millisecond * 20)
			})
			defer server.Close()

			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"--api-timeout", "10ms",
				"config", "control-planes", "add",
				"--name", "example",
				"--address", fmt.Sprintf("http://localhost:%d", port),
			})
			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Client.Timeout exceeded"))
			Expect(outbuf.String()).To(ContainSubstring(`Error: could not connect to the Control Plane API Server`))
			Expect(outbuf.String()).To(ContainSubstring(`Client.Timeout exceeded`))
		})

		It("should fail on invalid api server", func() {
			// setup
			server, port := setupCpServer(func(writer http.ResponseWriter, req *http.Request) {
				_, err := writer.Write([]byte("{}"))
				Expect(err).ToNot(HaveOccurred())
			})
			defer server.Close()

			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "add",
				"--name", "example",
				"--address", fmt.Sprintf("http://localhost:%d", port)})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(MatchError(`provided address is not valid Kuma Control Plane API Server`))
			// and
			Expect(outbuf.String()).To(Equal(`Error: provided address is not valid Kuma Control Plane API Server
`))
		})
	})

	Describe("happy path", func() {

		type testCase struct {
			configFile  string
			goldenFile  string
			expectedOut string
			overwrite   bool
			extraArgs   []string
		}

		DescribeTable("should add a new Control Plane by name and address",
			func(given testCase) {
				// setup
				initial, err := os.ReadFile(filepath.Join("testdata", given.configFile))
				Expect(err).ToNot(HaveOccurred())
				err = os.WriteFile(configFile.Name(), initial, 0600)
				Expect(err).ToNot(HaveOccurred())

				// setup cp index server for validation to pass
				server, port := setupCpIndexServer()
				defer server.Close()

				args := []string{"--config-file", configFile.Name(),
					"config", "control-planes", "add",
					"--name", "example",
					"--address", fmt.Sprintf("http://localhost:%d", port),
					"--ca-cert-file", "/tmp/ca-cert.pem",
					"--client-cert-file", "/tmp/client.cert.pem",
					"--client-key-file", "/tmp/client.key.pem",
				}
				if given.overwrite {
					args = append(args, "--overwrite")
				}
				args = append(args, given.extraArgs...)

				// given
				rootCmd.SetArgs(args)
				// when
				err = rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expectedWithPlaceholder, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				expected := strings.ReplaceAll(string(expectedWithPlaceholder), "http://placeholder-address", fmt.Sprintf("http://localhost:%d", port))

				// when
				actual, err := os.ReadFile(configFile.Name())
				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(actual).To(MatchYAML(expected))
				// and
				Expect(outbuf.String()).To(Equal(strings.TrimLeftFunc(given.expectedOut, unicode.IsSpace)))
			},
			Entry("should add a first Control Plane", testCase{
				configFile: "config-control-planes-add.01.initial.yaml",
				goldenFile: "config-control-planes-add.01.golden.yaml",
				expectedOut: `
added Control Plane "example"
switched active Control Plane to "example"
`,
				overwrite: false,
				extraArgs: nil,
			}),
			Entry("should add a second Control Plane", testCase{
				configFile: "config-control-planes-add.02.initial.yaml",
				goldenFile: "config-control-planes-add.02.golden.yaml",
				expectedOut: `
added Control Plane "example"
switched active Control Plane to "example"
`,
				overwrite: false,
				extraArgs: nil,
			}),
			Entry("should replace the example Control Plane", testCase{
				configFile: "config-control-planes-add.03.initial.yaml",
				goldenFile: "config-control-planes-add.03.golden.yaml",
				expectedOut: `
added Control Plane "example"
switched active Control Plane to "example"
`,
				overwrite: true,
				extraArgs: nil,
			}),
			Entry("should add the example Control Plane with headers", testCase{
				configFile: "config-control-planes-add.04.initial.yaml",
				goldenFile: "config-control-planes-add.04.golden.yaml",
				expectedOut: `
added Control Plane "example"
switched active Control Plane to "example"
`,
				overwrite: true,
				extraArgs: []string{"--headers", "abc=xyz", "--headers", "def=pqr"},
			}),
			Entry("should replace the example Control Plane with headers", testCase{
				configFile: "config-control-planes-add.05.initial.yaml",
				goldenFile: "config-control-planes-add.05.golden.yaml",
				expectedOut: `
added Control Plane "example"
switched active Control Plane to "example"
`,
				overwrite: true,
				extraArgs: []string{"--headers", "abc=xyz"},
			}),
		)
	})
})

func setupCpIndexServer() (*httptest.Server, int) {
	return setupCpServer(func(writer http.ResponseWriter, req *http.Request) {
		response := types.IndexResponse{
			Tagline: version.Product,
			Version: "unknown",
		}
		marshaled, err := json.Marshal(response)
		Expect(err).ToNot(HaveOccurred())
		_, err = writer.Write(marshaled)
		Expect(err).ToNot(HaveOccurred())
	})
}

func setupCpServer(fn func(http.ResponseWriter, *http.Request)) (*httptest.Server, int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		defer GinkgoRecover()
		fn(writer, req)
	})
	server := httptest.NewServer(mux)
	port, err := strconv.Atoi(strings.Split(server.Listener.Addr().String(), ":")[1])
	Expect(err).ToNot(HaveOccurred())
	return server, port
}
