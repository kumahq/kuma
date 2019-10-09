package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/server"
	"github.com/Kong/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"strings"
)

type staticCredentialGenerator struct {
	resp string
}

var _ auth.CredentialGenerator = &staticCredentialGenerator{}

func (s *staticCredentialGenerator) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	return auth.Credential(s.resp), nil
}

var _ = Describe("Initial Token Server", func() {

	var port int
	const credentials = "test"

	BeforeEach(func() {
		p, err := test.GetFreePort()
		port = p
		Expect(err).ToNot(HaveOccurred())
		srv := server.InitialTokenServer{
			LocalHttpPort:       port,
			CredentialGenerator: &staticCredentialGenerator{credentials},
		}

		ch := make(chan struct{})
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()
			errCh <- srv.Start(ch)
		}()
	})

	It("should respond with generated token", func(done Done) {
		// given
		idReq := server.IdentityRequest{
			Mesh: "defualt",
			Name: "dp-1",
		}
		reqBytes, err := json.Marshal(idReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/token", port), bytes.NewReader(reqBytes))
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// when
		respBody, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(respBody)).To(Equal(credentials))

		// finally
		close(done)
	})

	DescribeTable("should return bad request on invalid json",
		func(json string) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/token", port), strings.NewReader(json))
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		},
		Entry("json does not contain name", `{"mesh": "default"}`),
		Entry("json does not contain mesh", `{"name": "default"}`),
		Entry("not valid json", `not-valid-json`),
	)
})
