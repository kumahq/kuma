package tokens_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	config_kumactl "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/tokens/builtin/access"
	tokens_server "github.com/kumahq/kuma/pkg/tokens/builtin/server"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
)

type zoneStaticTokenIssuer struct {
}

var _ zone.TokenIssuer = &zoneStaticTokenIssuer{}

func (z *zoneStaticTokenIssuer) Generate(ctx context.Context, identity zone.Identity, validFor time.Duration) (zone.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

var _ = Describe("Zone Egress Tokens Client", func() {

	var server *httptest.Server

	BeforeEach(func() {
		container := restful.NewContainer()
		container.Add(tokens_server.NewWebservice(
			&staticTokenIssuer{},
			&zoneIngressStaticTokenIssuer{},
			&zoneStaticTokenIssuer{},
			access.NoopDpTokenAccess{},
			zone_access.NoopZoneTokenAccess{},
		))
		server = httptest.NewServer(container.ServeMux)
	})

	AfterEach(func() {
		server.Close()
	})

	It("should return a token", func() {
		// given
		baseClient, err := kumactl_client.ApiServerClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		}, time.Second)
		Expect(err).ToNot(HaveOccurred())
		client := tokens.NewZoneTokenClient(baseClient)

		// wait for server
		Eventually(func() error {
			_, err := client.Generate("my-zone-1", zone.FullScope, 24*time.Hour)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// when
		token, err := client.Generate("my-zone-1", zone.FullScope, 24*time.Hour)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(token).To(Equal("token-for-my-zone-1"))
	})

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/tokens/zone", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			writer.WriteHeader(500)
			_, err := writer.Write([]byte("Internal Server Error"))
			Expect(err).ToNot(HaveOccurred())
		})
		baseClient, err := kumactl_client.ApiServerClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		}, time.Second)
		Expect(err).ToNot(HaveOccurred())
		client := tokens.NewZoneTokenClient(baseClient)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("my-zone-2", zone.FullScope, 24*time.Hour)

		// then
		Expect(err).To(MatchError("(500): Internal Server Error"))
	})
})
