package generate_test

import (
	"bytes"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/cli/generate"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	"github.com/kumahq/kuma/pkg/util/http"
)

type fakeUserTokenClient struct {
}

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
		rootCmd.SetArgs([]string{"generate", "user-token",
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

	It("should throw an error when name is not specified", func() {
		// setup
		rootCmd := cmd.NewRootCmd(kumactl_cmd.DefaultRootContext())
		generate.NewHTTPUserTokenClient = func(client http.Client) client.UserTokenClient {
			return &fakeUserTokenClient{}
		}
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{"generate", "user-token"})

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).To(MatchError(`required flag(s) "name", "valid-for" not set`))
	})
})
