package envoy

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-dp"

	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
)

var _ = Describe("Bootstrap Config", func() {
	Describe("MinimalBootstrapConfig(..)", func() {
		It("should generate minimal Envoy bootstrap config", func() {
			// given
			input, _ := ioutil.ReadFile(filepath.Join("testdata", "minimal-bootstrap-config.input.yaml"))
			cfg := konvoy_dp.Config{}
			Expect(config.FromYAML(input, &cfg)).To(Succeed())

			// when
			envoyConfig, err := MinimalBootstrapConfig(cfg)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(envoyConfig)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			expected, _ := ioutil.ReadFile(filepath.Join("testdata", "minimal-bootstrap-config.golden.yaml"))
			// then
			Expect(actual).To(MatchYAML(expected))
		})
	})
})
