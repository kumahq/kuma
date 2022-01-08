package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Index Endpoints", func() {

	var backupBuildInfo kuma_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = kuma_version.Build
	})
	AfterEach(func() {
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

		// setup
		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get("http://" + apiServer.Address())
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get("http://" + apiServer.Address())
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		hostname, err := os.Hostname()
		Expect(err).ToNot(HaveOccurred())

		expected := fmt.Sprintf(`
		{
			"hostname": "%s",
			"tagline": "Kuma",
			"version": "1.2.3",
			"instanceId": "instance-id",
			"clusterId": "cluster-id",
			"gui": "The gui is available at /gui"
		}`, hostname)

		Expect(body).To(MatchJSON(expected))
	}))
})
