package maps_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/maps"
)

var _ = Describe("SortedKeys", func() {
	It("should return sorted keys", func() {
		// given
		m := map[string]string{
			"c": "x",
			"b": "y",
			"a": "z",
		}

		// when
		keys := maps.SortedKeys(m)

		// then
		Expect(keys).To(Equal([]string{"a", "b", "c"}))
	})
})
