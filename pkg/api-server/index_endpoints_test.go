package api_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

var _ = Describe("Index Endpoints", func() {

	var backupBuildInfo kuma_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = kuma_version.Build
	})
	AfterEach(func() {
		kuma_version.Build = backupBuildInfo
	})

	It("should return the version of Kuma Control Plane", func(done Done) {
		// given
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "1.2.3",
			GitTag:    "v1.2.3",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig())

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get("http://localhost" + apiServer.Address())
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get("http://localhost" + apiServer.Address())
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		hostname, err := os.Hostname()
		Expect(err).ToNot(HaveOccurred())

		expected := fmt.Sprintf(`
		{
			"hostname": "%s",
			"tagline": "Kuma",
			"version": "1.2.3"
		}`, hostname)

		Expect(body).To(MatchJSON(expected))
		close(done)
	}, 5)
})
