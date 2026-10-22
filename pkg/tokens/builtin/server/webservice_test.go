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
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(context.Context, issuer.DataplaneIdentity, time.Duration) (tokens.Token, error) {
	return s.resp, nil
}

type zoneStaticTokenIssuer struct{}

var _ zone.TokenIssuer = &zoneStaticTokenIssuer{}

func (z *zoneStaticTokenIssuer) Generate(ctx context.Context, identity zone.Identity, validFor time.Duration) (zone.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

type capturingDpTokenAccess struct {
	gotWorkload string
	reject      bool
}

var _ access.DataplaneTokenAccess = &capturingDpTokenAccess{}

func (c *capturingDpTokenAccess) ValidateGenerateDataplaneToken(_ context.Context, _ string, _ string, _ map[string][]string, workload string, _ user.User) error {
	c.gotWorkload = workload
	if c.reject {
		return fmt.Errorf("workload %q not allowed by this token", workload)
	}
	return nil
}

var _ = Describe("Dataplane Token Webservice", func() {
	const credentials = "test"
	var url string

	BeforeEach(func() {
		ws := server.NewWebservice(
			"/tokens",
			&staticTokenIssuer{credentials},
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

	DescribeTable("should return bad request on invalid zone token request",
		func(json string) {
			// given
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("%s/tokens/zone", url), strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			// when
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		},
		Entry("not valid json", `not-valid-json`),
		Entry("scope is a string", `{"zone": "zone-1", "scope": "egress"}`),
		Entry("scope is an object", `{"zone": "zone-1", "scope": {"name": "egress"}}`),
	)

	Context("workload authorization", func() {
		var dpAccess *capturingDpTokenAccess
		var wsURL string

		postWorkload := func(workload string) *http.Response {
			reqBytes, err := json.Marshal(types.DataplaneTokenRequest{Mesh: "default", Workload: workload})
			Expect(err).ToNot(HaveOccurred())
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fmt.Sprintf("%s/tokens/dataplane", wsURL), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			return resp
		}

		BeforeEach(func() {
			dpAccess = &capturingDpTokenAccess{}
			ws := server.NewWebservice(
				"/tokens",
				&staticTokenIssuer{credentials},
				&zoneStaticTokenIssuer{},
				dpAccess,
				&zone_access.NoopZoneTokenAccess{},
			)
			container := restful.NewContainer()
			container.Add(ws)
			srv := httptest.NewServer(container)
			wsURL = srv.URL
			Eventually(func() error {
				_, err := http.DefaultClient.Get(fmt.Sprintf("%s/tokens/dataplane", srv.URL))
				return err
			}).ShouldNot(HaveOccurred())
		})

		It("passes the requested workload to the authorization hook", func() {
			resp := postWorkload("backend")
			Expect(resp.StatusCode).To(Equal(200))
			Expect(dpAccess.gotWorkload).To(Equal("backend"))
		})

		It("does not issue a token when authorization rejects the workload", func() {
			dpAccess.reject = true
			resp := postWorkload("backend")
			Expect(resp.StatusCode).ToNot(Equal(200))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).ToNot(Equal(credentials))
		})
	})
})
