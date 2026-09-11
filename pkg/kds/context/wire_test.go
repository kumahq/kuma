package context_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
)

// The golden file records the Any type URL every KDS synced type is sent under.
// A control plane on the other side of an upgrade reads the Any by that URL, so
// a type that trades a URL for the empty string (the JSON encoding a spec gets
// once it stops being a protobuf message) stops syncing to it entirely: the
// peer fails the whole DeltaDiscoveryResponse and restarts its stream forever.
// Keeping such a spec on the wire is what core_model.KDSWireSpec is for.
var _ = Describe("KDS wire encoding", func() {
	It("sends every synced type under the type URL its peers expect", func() {
		reg := registry.Global()
		urls := map[string]string{}
		for _, filter := range []core_model.TypeFilter{
			core_model.SentFromGlobalToZone(),
			core_model.SentFromZoneToGlobal(),
		} {
			for _, typ := range reg.ObjectTypes(filter) {
				desc, err := reg.DescriptorFor(typ)
				Expect(err).ToNot(HaveOccurred())
				any, err := core_model.ToAny(desc.NewObject().GetSpec())
				Expect(err).ToNot(HaveOccurred())
				urls[string(typ)] = any.GetTypeUrl()
			}
		}

		actual, err := yaml.Marshal(urls)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(matchers.MatchGoldenYAML("testdata", "kds-wire-type-urls.golden.yaml"))
	})
})
