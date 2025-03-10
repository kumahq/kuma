package generate_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/cmd/generate"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	"github.com/kumahq/kuma/pkg/util/http"
)

type fakeUserTokenClient struct{}

func (f *fakeUserTokenClient) Generate(name string, groups []string, validFor time.Duration) (string, error) {
	return "token-" + name + "-" + strings.Join(groups, ",") + "-" + validFor.String(), nil
}

var _ client.UserTokenClient = &fakeUserTokenClient{}

var _ = Describe("Generate User Token", func() {
	It("should generate control plane token", func() {
		// setup
		rootCmd := cmd.NewRootCmd(kumactl_cmd.DefaultRootContext())
		generate.NewHTTPUserTokenClient = func(client http.Client) client.UserTokenClient {
			return &fakeUserTokenClient{}
		}
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{
			"generate", "user-token",
			"--name", "john",
			"--group", "team-a",
			"--group", "team-b",
			"--valid-for", "30s",
		})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(Equal("token-john-team-a,team-b-30s"))
	})

	It("should issue token offline", func() {
		rootCmd := cmd.NewRootCmd(kumactl_cmd.DefaultRootContext())
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{
			"generate", "user-token",
			"--name", "john",
			"--group", "team-a",
			"--group", "team-b",
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
		userClaims := &issuer.UserClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(buf.String(), userClaims)
		Expect(err).ToNot(HaveOccurred())
		Expect(userClaims.User.Name).To(Equal("john"))
		Expect(userClaims.User.Groups).To(Equal([]string{"team-a", "team-b"}))
	})

	type errTestCase struct {
		args []string
		err  string
	}

	DescribeTable("should trow an error",
		func(given errTestCase) {
			// setup
			rootCmd := cmd.NewRootCmd(kumactl_cmd.DefaultRootContext())
			generate.NewHTTPUserTokenClient = func(client http.Client) client.UserTokenClient {
				return &fakeUserTokenClient{}
			}
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)

			// given
			rootCmd.SetArgs(given.args)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(MatchError(given.err))
		},
		Entry("when name is not specified", errTestCase{
			args: []string{"generate", "user-token"},
			err:  `required flag(s) "name", "valid-for" not set`,
		}),
		Entry("when kid is specified for online signing", errTestCase{
			args: []string{
				"generate", "user-token",
				"--name", "john",
				"--group", "team-a",
				"--group", "team-b",
				"--valid-for", "30s",
				"--kid", "1",
			},
			err: "--kid cannot be used when --signing-key-path is used",
		}),
		Entry("when kid is not specified for offline signing", errTestCase{
			args: []string{
				"generate", "user-token",
				"--name", "john",
				"--group", "team-a",
				"--group", "team-b",
				"--valid-for", "30s",
				"--signing-key-path", filepath.Join("..", "..", "..", "..", "test", "keys", "samplekey.pem"),
			},
			err: "--kid is required when --signing-key-path is used",
		}),
	)
})
