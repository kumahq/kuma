package maps_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/maps"
)

var _ = Describe("Maps", func() {
	It("SortedKeys should return sorted keys", func() {
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

	It("AllKeys should return all keys", func() {
		// given
		m := map[string]string{
			"c": "x",
			"b": "y",
			"a": "z",
		}

		// when
		keys := maps.AllKeys(m)

		// then
		Expect(keys).To(ConsistOf([]string{"c", "b", "a"}))
	})

	It("AllValues should return all values", func() {
		// given
		m := map[string]string{
			"c": "x",
			"b": "y",
			"a": "z",
		}

		// when
		keys := maps.AllValues(m)

		// then
		Expect(keys).To(ConsistOf([]string{"x", "y", "z"}))
	})
})
