package xds

import (
	_ "embed"

	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
