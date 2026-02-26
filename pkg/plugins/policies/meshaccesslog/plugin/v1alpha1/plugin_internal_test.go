package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	plugin_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/plugin/xds"
)

var _ = Describe("toOrderedDpBackends", func() {
	It("should sort backends by name and map fields correctly", func() {
		backends := map[string]plugin_xds.OtelPipeBackendInfo{
			"backend-b": {
				SocketPath: "/tmp/b.sock",
				Endpoint:   "collector-b:4317",
			},
			"backend-a": {
				SocketPath: "/tmp/a.sock",
				Endpoint:   "collector-a:4317",
				UseHTTP:    true,
				Path:       "/custom",
			},
		}

		ordered := toOrderedDpBackends(backends)

		Expect(ordered).To(HaveLen(2))
		Expect(ordered[0].SocketPath).To(Equal("/tmp/a.sock"))
		Expect(ordered[0].Endpoint).To(Equal("collector-a:4317"))
		Expect(ordered[0].UseHTTP).To(BeTrue())
		Expect(ordered[0].Path).To(Equal("/custom"))
		Expect(ordered[1].SocketPath).To(Equal("/tmp/b.sock"))
		Expect(ordered[1].Endpoint).To(Equal("collector-b:4317"))
		Expect(ordered[1].UseHTTP).To(BeFalse())
	})
})
