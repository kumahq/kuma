package xds

import (
	_ "embed"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

func ResourceArrayShouldEqual(resources core_xds.ResourceList, expected []string) {
	Expect(resources).To(HaveLen(len(expected)))

	for i, r := range resources {
		actual, err := util_proto.ToYAML(r.Resource)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchYAML(expected[i]))
	}
	Expect(resources).To(HaveLen(len(expected)))
}

// LegacyServiceName is the value a destination carries in the `kuma.io/service`
// tag, which stays legacy-formatted even under unified resource naming.
func LegacyServiceName(id kri.Identifier, port int32) string {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	Expect(err).ToNot(HaveOccurred())
	return destinationname.ResolveLegacyFromKRI(id, desc.ShortName, port)
}
