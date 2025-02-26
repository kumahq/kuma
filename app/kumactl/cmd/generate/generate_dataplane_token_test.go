package generate_test

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type staticDataplaneTokenGenerator struct {
	err error
}

var _ tokens.DataplaneTokenClient = &staticDataplaneTokenGenerator{}

func (s *staticDataplaneTokenGenerator) Generate(name string, mesh string, tags map[string][]string, dpType string, validFor time.Duration) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return fmt.Sprintf("token-for-%s-%s-%s-%s", name, mesh, mesh_proto.MultiValueTagSetFrom(tags).String(), dpType), nil
}

var _ = Describe("kumactl generate dataplane-token", func() {
	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var generator *staticDataplaneTokenGenerator
	var ctx *kumactl_cmd.RootContext

	BeforeEach(func() {
		generator = &staticDataplaneTokenGenerator{}
		ctx = test_kumactl.MakeMinimalRootContext()
		ctx.Runtime.NewDataplaneTokenClient = func(util_http.Client) tokens.DataplaneTokenClient {
			return generator
		}

		rootCmd = cmd.NewRootCmd(ctx)

		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
	})

	type testCase struct {
		args   []string
		result string
	}
	DescribeTable("should generate token",
		func(given testCase) {
			// when
			rootCmd.SetArgs(given.args)
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(buf.String()).To(Equal(given.result))
		},
		Entry("for default mesh when it is not specified", testCase{
			args:   []string{"generate", "dataplane-token", "--name=example", "--valid-for", "30s"},
			result: "token-for-example-default--",
		}),
		Entry("for all arguments", testCase{
			args:   []string{"generate", "dataplane-token", "--mesh=demo", "--name=example", "--proxy-type=dataplane", "--tag", "kuma.io/service=web", "--valid-for", "30s"},
			result: "token-for-example-demo-kuma.io/service=web-dataplane",
		}),
	)

	It("should issue token offline", func() {
		// given
		rootCmd.SetArgs([]string{
			"generate", "dataplane-token",
			"--name", "dp-1",
			"--mesh", "demo",
			"--valid-for", "30s",
			"--kid", "1",
			"--signing-key-path", filepath.Join("..", "..", "..", "..", "test", "keys", "samplekey.pem"),
		})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())

		// and the token is valid
		claims := &issuer.DataplaneClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(buf.String(), claims)
		Expect(err).ToNot(HaveOccurred())
		Expect(claims.Name).To(Equal("dp-1"))
		Expect(claims.Mesh).To(Equal("demo"))
	})

	It("should write error when generating token fails", func() {
		// setup
		generator.err = errors.New("could not connect to API")

		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--name=example", "--valid-for", "30s"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to generate a dataplane token: could not connect to API\n"))
	})

	type errTestCase struct {
		args []string
		err  string
	}

	DescribeTable("should trow an error",
		func(given errTestCase) {
			// given
			rootCmd.SetArgs(given.args)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(MatchError(given.err))
		},
		Entry("when kid is specified for online signing", errTestCase{
			args: []string{
				"generate", "dataplane-token",
				"--name", "dp-1",
				"--kid", "1",
				"--valid-for", "30s",
			},
			err: "--kid cannot be used when --signing-key-path is used",
		}),
		Entry("when kid is not specified for offline signing", errTestCase{
			args: []string{
				"generate", "user-token",
				"--name", "dp-1",
				"--valid-for", "30s",
				"--signing-key-path", filepath.Join("..", "..", "..", "..", "..", "..", "..", "test", "keys", "samplekey.pem"),
			},
			err: "--kid is required when --signing-key-path is used",
		}),
	)
})
