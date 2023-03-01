package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/test"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Index Endpoints", func() {
	stop := func() {}
	var backupBuildInfo kuma_version.BuildInfo
	var apiServer *api_server.ApiServer
	BeforeEach(func() {
		backupBuildInfo = kuma_version.Build
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithConfigMutator(func(config *server.ApiServerConfig) {
			config.GUI.Enabled = true
			config.GUI.RootUrl = "https://foo.bar.com:5000/from"
		}))
	})
	AfterEach(func() {
		stop()
		kuma_version.Build = backupBuildInfo
	})

	It("should return the version of Kuma Control Plane", test.Within(5*time.Second, func() {
		// given
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "1.2.3",
			GitTag:    "v1.2.3",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
		hostname, err := os.Hostname()
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := http.Get("http://" + apiServer.Address())
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		expected := fmt.Sprintf(`
		{
			"hostname": "%s",
			"tagline": "Kuma",
			"product": "Kuma",
			"version": "1.2.3",
			"instanceId": "instance-id",
			"clusterId": "cluster-id",
			"gui": "https://foo.bar.com:5000/from"
		}`, hostname)

		Expect(body).To(MatchJSON(expected))
	}))
})
