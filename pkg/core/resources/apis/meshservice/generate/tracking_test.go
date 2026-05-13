package generate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("propagation tracking", func() {
	Describe("addPropagationTracking", func() {
		It("tracks qualified-name keys containing '/'", func() {
			labels := map[string]string{}
			addPropagationTracking(labels, map[string]string{
				"app.example.com/tier": "gold",
				"simple":               "val",
			})
			count := 0
			for k := range labels {
				if len(k) > len(propagationTrackingPrefix) && k[:len(propagationTrackingPrefix)] == propagationTrackingPrefix {
					count++
				}
			}
			Expect(count).To(Equal(2))
		})

		It("uses empty string as tracking label value", func() {
			labels := map[string]string{}
			addPropagationTracking(labels, map[string]string{"foo.io/bar": "x"})
			for k, v := range labels {
				if k[:len(propagationTrackingPrefix)] == propagationTrackingPrefix {
					Expect(v).To(BeEmpty())
				}
			}
		})

		It("clears old tracking labels before writing new ones", func() {
			labels := map[string]string{
				propagationTrackingPrefix + "0":        "old",
				propagationTrackingPrefix + "deadbeef": "",
				"external":                             "keep",
			}
			addPropagationTracking(labels, map[string]string{"new-key": "v"})
			Expect(labels).ToNot(HaveKey(propagationTrackingPrefix + "0"))
			Expect(labels).ToNot(HaveKey(propagationTrackingPrefix + "deadbeef"))
			Expect(labels).To(HaveKeyWithValue("external", "keep"))
		})

		It("writes no tracking labels for empty propagated map", func() {
			labels := map[string]string{propagationTrackingPrefix + "0": "stale"}
			addPropagationTracking(labels, map[string]string{})
			for k := range labels {
				Expect(k).ToNot(HavePrefix(propagationTrackingPrefix))
			}
		})
	})

	Describe("extractPropagatedKeys", func() {
		It("recognizes dotted qualified-name keys written by addPropagationTracking", func() {
			labels := map[string]string{}
			addPropagationTracking(labels, map[string]string{
				"app.example.com/tier": "gold",
				"simple":               "val",
			})
			wasPropagated := extractPropagatedKeys(labels)
			Expect(wasPropagated("app.example.com/tier")).To(BeTrue())
			Expect(wasPropagated("simple")).To(BeTrue())
			Expect(wasPropagated("not-tracked")).To(BeFalse())
		})

		It("returns false for all keys when no tracking labels present", func() {
			wasPropagated := extractPropagatedKeys(map[string]string{"foo": "bar"})
			Expect(wasPropagated("foo")).To(BeFalse())
			Expect(wasPropagated("app.example.com/tier")).To(BeFalse())
		})

		It("does not recognize old-format tracking labels (key name as value)", func() {
			// Old format: kuma.io/pkey-0 = "foo.example.com/bar"
			// After upgrade, the first reconcile replaces old labels; but before
			// that, extractPropagatedKeys should NOT recognize old labels as tracking
			// entries for their stored key names.
			labels := map[string]string{
				propagationTrackingPrefix + "0": "app.example.com/tier",
			}
			wasPropagated := extractPropagatedKeys(labels)
			Expect(wasPropagated("app.example.com/tier")).To(BeFalse())
		})
	})
})
