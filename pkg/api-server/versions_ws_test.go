package api_server_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Versions WS", func() {
	It("should return the supported versions", func() {
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
			_, err := http.Get(fmt.Sprintf("http://%s/versions", apiServer.Address()))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://%s/versions", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		var data struct {
			KumaDp map[string]struct {
				Envoy string
			}
		}

		Expect(json.NewDecoder(resp.Body).Decode(&data)).ToNot(HaveOccurred())

		// 1.0.0
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.0"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.0"].Envoy).To(Equal("1.16.0"))

		// 1.0.1
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.1"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.1"].Envoy).To(Equal("1.16.0"))

		// 1.0.2
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.2"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.2"].Envoy).To(Equal("1.16.1"))

		// 1.0.3
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.3"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.3"].Envoy).To(Equal("1.16.1"))

		// 1.0.4
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.4"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.4"].Envoy).To(Equal("1.16.1"))

		// 1.0.5
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.5"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.5"].Envoy).To(Equal("1.16.2"))

		// 1.0.6
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.6"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.6"].Envoy).To(Equal("1.16.2"))

		// 1.0.7
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.7"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.7"].Envoy).To(Equal("1.16.2"))

		// 1.0.8
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.0.8"]).ToNot(BeNil())
		Expect(data.KumaDp["1.0.8"].Envoy).To(Equal("1.16.2"))

		// 1.1.0
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.1.0"]).ToNot(BeNil())
		Expect(data.KumaDp["1.1.0"].Envoy).To(Equal("~1.17.0"))

		// 1.1.1
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.1.1"]).ToNot(BeNil())
		Expect(data.KumaDp["1.1.1"].Envoy).To(Equal("~1.17.0"))

		// 1.1.2
		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())
		Expect(data.KumaDp["1.1.2"]).ToNot(BeNil())
		Expect(data.KumaDp["1.1.2"].Envoy).To(Equal("~1.17.0"))
	})
})
