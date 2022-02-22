package generate_test

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	"github.com/kumahq/kuma/pkg/util/test"
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
		ctx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Registry: registry.NewTypeRegistry(),
				NewBaseAPIServerClient: func(server *config_proto.ControlPlaneCoordinates_ApiServer, _ time.Duration) (util_http.Client, error) {
					return nil, nil
				},
				NewZoneTokenClient: func(util_http.Client) tokens.ZoneTokenClient {
					return generator
				},
				NewAPIServerClient: test.GetMockNewAPIServerClient(),
			},
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

})
