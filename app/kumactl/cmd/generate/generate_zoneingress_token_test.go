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

type staticZoneIngressTokenGenerator struct {
	err error
}

var _ tokens.ZoneIngressTokenClient = &staticZoneIngressTokenGenerator{}

func (s *staticZoneIngressTokenGenerator) Generate(zone string, validFor time.Duration) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return fmt.Sprintf("token-for-%s", zone), nil
}

var _ = Describe("kumactl generate zone-ingress-token", func() {

	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var generator *staticZoneIngressTokenGenerator
	var ctx *kumactl_cmd.RootContext

	BeforeEach(func() {
		generator = &staticZoneIngressTokenGenerator{}
		ctx = &kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				Registry: registry.NewTypeRegistry(),
				NewBaseAPIServerClient: func(server *config_proto.ControlPlaneCoordinates_ApiServer, _ time.Duration) (util_http.Client, error) {
					return nil, nil
				},
				NewZoneIngressTokenClient: func(util_http.Client) tokens.ZoneIngressTokenClient {
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
			args:   []string{"generate", "zone-ingress-token", "--zone=my-zone"},
			result: "token-for-my-zone",
		}),
		Entry("for empty zone", testCase{
			args:   []string{"generate", "zone-ingress-token"},
			result: "token-for-",
		}),
	)

	It("should write error when generating token fails", func() {
		// setup
		generator.err = errors.New("could not connect to API")

		// when
		rootCmd.SetArgs([]string{"generate", "zone-ingress-token", "--zone=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to generate a zone ingress token: could not connect to API\n"))
	})

})
