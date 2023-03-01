package api_server_test

import (
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Policies Endpoints", func() {
	stop := func() {}
	var apiServer *api_server.ApiServer
	BeforeEach(func() {
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer())
	})
	AfterEach(func() {
		stop()
	})

	It("should return the list of policies", test.Within(5*time.Second, func() {
		// given

		// when
		resp, err := http.Get("http://" + apiServer.Address() + "/policies")
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(matchers.MatchGoldenJSON("testdata", "policies_list.golden.json"))
	}))
})
