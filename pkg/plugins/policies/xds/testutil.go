package xds

import (
	_ "embed"
	"time"

	. "github.com/onsi/gomega"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ResourceArrayShouldEqual(resources core_xds.ResourceList, expected []string) {
	for i, r := range resources {
		actual, err := util_proto.ToYAML(r.Resource)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchYAML(expected[i]))
	}
	Expect(len(resources)).To(Equal(len(expected)))
}

func ParseDuration(duration string) *k8s.Duration {
	d, _ := time.ParseDuration(duration)
	return &k8s.Duration{Duration: d}
}

func PointerOf[T any](value T) *T {
	return &value
}
