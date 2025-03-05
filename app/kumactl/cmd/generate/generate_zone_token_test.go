package generate_test

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type staticZoneTokenGenerator struct {
	err error
}

var _ tokens.ZoneTokenClient = &staticZoneTokenGenerator{}

func (s *staticZoneTokenGenerator) Generate(
	zone string,
	_ []string,
	_ time.Duration,
) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return fmt.Sprintf("token-for-%s", zone), nil
}

var _ = Describe("kumactl generate zone-token", func() {
	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var generator *staticZoneTokenGenerator
	var ctx *kumactl_cmd.RootContext

	BeforeEach(func() {
		generator = &staticZoneTokenGenerator{}
		ctx = test_kumactl.MakeMinimalRootContext()
		ctx.Runtime.NewZoneTokenClient = func(util_http.Client) tokens.ZoneTokenClient {
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
		Entry("for zone", testCase{
			args:   []string{"generate", "zone-token", "--zone=my-zone", "--valid-for=24h"},
			result: "token-for-my-zone",
		}),
	)

	It("should issue token offline", func() {
		// when
		rootCmd.SetArgs([]string{
			"generate", "zone-token",
			"--zone=my-zone",
			"--valid-for=24h",
			"--kid", "1",
			"--signing-key-path", filepath.Join("..", "..", "..", "..", "test", "keys", "samplekey.pem"),
		})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())

		// and the token is valid
		claims := &zone.ZoneClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(buf.String(), claims)
		Expect(err).ToNot(HaveOccurred())
		Expect(claims.Zone).To(Equal("my-zone"))
		Expect(claims.Scope).To(Equal([]string{"ingress", "egress"}))
	})

	It("should write error when generating token fails", func() {
		// setup
		generator.err = errors.New("could not connect to API")

		// when
		rootCmd.SetArgs([]string{"generate", "zone-token", "--zone=example", "--valid-for=24h"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to generate a zone token: could not connect to API\n"))
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
		Entry("when zone is not specified", errTestCase{
			args: []string{"generate", "zone-token"},
			err:  `required flag(s) "valid-for", "zone" not set`,
		}),
		Entry("when kid is specified for online signing", errTestCase{
			args: []string{
				"generate", "zone-token",
				"--zone", "east",
				"--valid-for", "30s",
				"--kid", "1",
			},
			err: "--kid cannot be used when --signing-key-path is used",
		}),
		Entry("when kid is not specified for offline signing", errTestCase{
			args: []string{
				"generate", "zone-token",
				"--zone", "east",
				"--valid-for", "30s",
				"--signing-key-path", filepath.Join("..", "..", "..", "..", "..", "..", "..", "test", "keys", "samplekey.pem"),
			},
			err: "--kid is required when --signing-key-path is used",
		}),
	)
})
