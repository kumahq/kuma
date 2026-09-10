package context_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"

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

// A 3.1 control plane will drop the protobuf wire form and send every spec as
// type-URL-less JSON. Only a 3.0 peer that reads that form correctly makes the
// switch safe, and a released 3.0 cannot be fixed afterwards, so the guarantee
// is pinned here rather than assumed.
var _ = Describe("KDS JSON wire encoding", func() {
	It("reads every synced type back from the JSON form", func() {
		reg := registry.Global()
		for _, filter := range []core_model.TypeFilter{
			core_model.SentFromGlobalToZone(),
			core_model.SentFromZoneToGlobal(),
		} {
			for _, typ := range reg.ObjectTypes(filter) {
				desc, err := reg.DescriptorFor(typ)
				Expect(err).ToNot(HaveOccurred())

				spec := desc.NewObject().GetSpec()
				encoded, err := json.Marshal(spec)
				Expect(err).ToNot(HaveOccurred(), "%s", typ)

				read := desc.NewObject().GetSpec()
				err = core_model.FromAny(&anypb.Any{Value: encoded}, read)
				Expect(err).ToNot(HaveOccurred(), "%s", typ)

				rewritten, err := json.Marshal(read)
				Expect(err).ToNot(HaveOccurred(), "%s", typ)
				Expect(string(rewritten)).To(Equal(string(encoded)), "%s", typ)
			}
		}
	})

	It("replaces rather than merges what the target already held", func() {
		spec := &system_proto.Secret{Data: system_proto.Bytes([]byte("old"))}

		empty, err := json.Marshal(&system_proto.Secret{})
		Expect(err).ToNot(HaveOccurred())
		Expect(core_model.FromAny(&anypb.Any{Value: empty}, spec)).To(Succeed())

		Expect(spec.Data).To(BeNil())
	})
})
