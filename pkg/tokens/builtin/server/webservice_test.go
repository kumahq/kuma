package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/access"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/server"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/server/types"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
	zone_access "github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone/access"
)

type staticTokenIssuer struct {
	resp string
	got  issuer.DataplaneIdentity
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(_ context.Context, identity issuer.DataplaneIdentity, _ time.Duration) (tokens.Token, error) {
	s.got = identity
	return s.resp, nil
}

type zoneStaticTokenIssuer struct {
	got zone.Identity
}

var _ zone.TokenIssuer = &zoneStaticTokenIssuer{}

func (z *zoneStaticTokenIssuer) Generate(ctx context.Context, identity zone.Identity, validFor time.Duration) (zone.Token, error) {
	z.got = identity
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

type capturingDpTokenAccess struct {
	got    issuer.DataplaneIdentity
	reject bool
}

var _ access.DataplaneTokenAccess = &capturingDpTokenAccess{}

func (c *capturingDpTokenAccess) ValidateGenerateDataplaneToken(_ context.Context, identity issuer.DataplaneIdentity, _ user.User) error {
	c.got = identity
	if c.reject {
		return fmt.Errorf("workload %q not allowed", identity.Workload)
	}
	return nil
}

type capturingZoneTokenAccess struct {
	got    zone.Identity
	reject bool
}

var _ zone_access.ZoneTokenAccess = &capturingZoneTokenAccess{}

func (c *capturingZoneTokenAccess) ValidateGenerateZoneToken(_ context.Context, identity zone.Identity, _ user.User) error {
	c.got = identity
	if c.reject {
		return fmt.Errorf("scope %v not allowed", identity.Scope)
	}
	return nil
}

var _ = Describe("Dataplane Token Webservice", func() {
	const credentials = "test"
	var url string

	BeforeEach(func() {
		ws := server.NewWebservice(
			"/tokens",
			&staticTokenIssuer{resp: credentials},
			&zoneStaticTokenIssuer{},
			&access.NoopDpTokenAccess{},
			&zone_access.NoopZoneTokenAccess{},
		)

		container := restful.NewContainer()
		container.Add(ws)
		srv := httptest.NewServer(container)
		url = srv.URL

		// wait for the server
		Eventually(func() error {
			_, err := http.DefaultClient.Get(fmt.Sprintf("%s/tokens/dataplane", srv.URL))
			return err
		}).ShouldNot(HaveOccurred())
	})

	It("should respond with generated token", func() {
		// given
		idReq := types.DataplaneTokenRequest{
			Mesh: "default",
			Name: "dp-1",
		}
		reqBytes, err := json.Marshal(idReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("%s/tokens/dataplane", url), bytes.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// when
		respBody, err := io.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(respBody)).To(Equal(credentials))
	})

	DescribeTable("should return bad request on invalid json",
		func(json string) {
			// given
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("%s/tokens/dataplane", url), strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			// when
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		},
		Entry("not valid json", `not-valid-json`),
	)

	Context("authorization hook", func() {
		var dpIssuer *staticTokenIssuer
		var zoneIssuer *zoneStaticTokenIssuer
		var dpAccess *capturingDpTokenAccess
		var zoneAccess *capturingZoneTokenAccess
		var wsURL string

		post := func(path string, body any) *http.Response {
			reqBytes, err := json.Marshal(body)
			Expect(err).ToNot(HaveOccurred())
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("%s/tokens/%s", wsURL, path), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			return resp
		}

		dpReq := types.DataplaneTokenRequest{
			Name:     "dp-1",
			Mesh:     "default",
			Tags:     map[string][]string{"kuma.io/service": {"backend"}},
			Type:     "dataplane",
			Workload: "backend",
		}
		zoneReq := types.ZoneTokenRequest{
			Zone:  "zone-1",
			Scope: []string{"egress"},
		}

		BeforeEach(func() {
			dpIssuer = &staticTokenIssuer{resp: credentials}
			zoneIssuer = &zoneStaticTokenIssuer{}
			dpAccess = &capturingDpTokenAccess{}
			zoneAccess = &capturingZoneTokenAccess{}
			ws := server.NewWebservice("/tokens", dpIssuer, zoneIssuer, dpAccess, zoneAccess)
			container := restful.NewContainer()
			container.Add(ws)
			srv := httptest.NewServer(container)
			wsURL = srv.URL
			Eventually(func() error {
				_, err := http.DefaultClient.Get(fmt.Sprintf("%s/tokens/dataplane", srv.URL))
				return err
			}).ShouldNot(HaveOccurred())
		})

		It("receives the same dataplane identity that is signed", func() {
			resp := post("dataplane", dpReq)
			Expect(resp.StatusCode).To(Equal(200))
			Expect(dpAccess.got).To(Equal(issuer.DataplaneIdentity{
				Name:     "dp-1",
				Mesh:     "default",
				Tags:     mesh_proto.MultiValueTagSet{"kuma.io/service": {"backend": true}},
				Type:     mesh_proto.DataplaneProxyType,
				Workload: "backend",
			}))
			Expect(dpAccess.got).To(Equal(dpIssuer.got))
		})

		It("does not issue a dataplane token when the hook rejects", func() {
			dpAccess.reject = true
			resp := post("dataplane", dpReq)
			Expect(resp.StatusCode).ToNot(Equal(200))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).ToNot(Equal(credentials))
			Expect(dpIssuer.got).To(BeZero())
		})

		It("receives the same zone identity that is signed", func() {
			resp := post("zone", zoneReq)
			Expect(resp.StatusCode).To(Equal(200))
			Expect(zoneAccess.got).To(Equal(zone.Identity{
				Zone:  "zone-1",
				Scope: []string{"egress"},
			}))
			Expect(zoneAccess.got).To(Equal(zoneIssuer.got))
		})

		It("does not issue a zone token when the hook rejects", func() {
			zoneAccess.reject = true
			resp := post("zone", zoneReq)
			Expect(resp.StatusCode).ToNot(Equal(200))
			Expect(zoneIssuer.got).To(BeZero())
		})
	})
})
