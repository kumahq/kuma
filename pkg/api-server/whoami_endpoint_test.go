package api_server_test

import (
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/test"
)

var _ = Describe("Whoami Endpoint", func() {
	stop := func() {}
	var apiServer *api_server.ApiServer
	BeforeEach(func() {
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer())
	})
	AfterEach(func() {
		stop()
	})

	It("should return the user information", test.Within(5*time.Second, func() {
		// when
		resp, err := http.Get("http://" + apiServer.Address() + "/who-am-i")
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// it's admin because of the default KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN=true and the server is running on localhost
		expected := `
        {
          "name": "mesh-system:admin",
          "groups": [
            "mesh-system:admin",
            "mesh-system:authenticated"
          ]
        }`

		Expect(body).To(MatchJSON(expected))
	}))
})
